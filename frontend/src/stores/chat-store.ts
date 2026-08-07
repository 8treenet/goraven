import { create } from 'zustand'
import type { StateCreator } from 'zustand'
import { toast } from 'sonner'
import { providersApi, personasApi, mcpApi, skillsApi, chatApi, sessionsApi } from '@/api'
import { connectChatStream } from '@/api/sse'
import type { ChatRequest, ChatResponse, SessionDetail, Message as ApiMessage, SharedProjectInfo } from '@/api/types'
import { uuid } from '@/lib/utils'

export type RoleType = 'user' | 'assistant'

export interface FileRef {
  type: 'file' | 'dir'
  name: string
  path: string
}

export interface Model {  id: number
  name: string
  provider: string
  icon: string
  isDefault: boolean
  isFlash: boolean
  isVisual: boolean
}

export interface McpEndpoint {
  id: number
  name: string
  description: string
  icon: string
}

export interface Skill {
  id: number
  name: string
  description: string
  icon: string
}

export interface Persona {
  id: number
  name: string
  icon: string
  modelId?: number
  mcpIds?: number[]
  skillIds?: number[]
}

export interface ToolCall {
  toolName: string
  displayName: string
  icon: string
  action: string
  duration?: number
  success?: boolean
}

export type ThinkingSegment =
  | { type: 'reasoning', content: string }
  | { type: 'tool', toolName: string, displayName: string, icon: string, action: string }
  | { type: 'retry', attempt: number, maxRetries: number, error: string }

export interface RetryInfo {
  attempt: number
  maxRetries: number
  error: string
}

export interface Message {
  id: string
  role: RoleType
  content: string
  reasoningContent?: string
  toolCalls?: ToolCall[]
  thinkingSegments?: ThinkingSegment[]
  timestamp: string
  roundId: string
  contextState: number
}

export interface Session {
  id: string
  title: string
  modelId: number
  personaId: number | null
  project: string
  sharedProject?: SharedProjectInfo
  mcpIds: number[]
  skillIds: number[]
  lastChatTime: string
  status: number
  messages: Message[]
  modelName: string
  modelIcon: string
  personaName: string
  personaIcon: string
  contextTokens: number
  contextLimit: number
  promptTokensCount: number
  completionTokensCount: number
}

/* ============================================
   Helper: map API message to store message
   ============================================ */

function mapApiMessage(m: ApiMessage): Message {
  const thinkingSegments: ThinkingSegment[] = []
  if (m.reasoningContent?.length) {
    for (const item of m.reasoningContent) {
      if (item.eventType === 'reasoning' && item.content) {
        thinkingSegments.push({ type: 'reasoning', content: item.content })
      } else if (item.eventType === 'tool' && item.tool) {
        thinkingSegments.push({
          type: 'tool',
          toolName: item.tool.name,
          displayName: item.tool.displayName,
          icon: item.tool.icon,
          action: item.tool.action,
        })
      }
    }
  }
  return {
    id: m.msgId,
    role: m.roleType === 'summary' ? 'assistant' : m.roleType,
    content: m.content,
    thinkingSegments: thinkingSegments.length > 0 ? thinkingSegments : undefined,
    timestamp: m.created,
    roundId: m.roundId,
    contextState: m.contextState,
  }
}

const pollIntervals = new Map<string, ReturnType<typeof setInterval>>()
const backgroundTimeouts = new Map<string, ReturnType<typeof setTimeout>>()
const BACKGROUND_TIMEOUT_MS = 5 * 60 * 1000

const LAST_USED_MODEL_KEY = 'goraven.lastUsedModelId'

function getLastUsedModelId(): number {
  const v = localStorage.getItem(LAST_USED_MODEL_KEY)
  return v ? Number(v) || 0 : 0
}

function setLastUsedModelId(id: number) {
  localStorage.setItem(LAST_USED_MODEL_KEY, String(id))
}

function startPolling(
  sessionId: string,
  get: () => ChatState,
  set: Parameters<StateCreator<ChatState>>[0],
) {
  if (pollIntervals.has(sessionId)) return
  if (get().generatingSessionId === sessionId) return
  const id = setInterval(async () => {
    try {
      const detail = await sessionsApi.getSessionDetail(sessionId)
      if (detail.status === 0) {
        clearInterval(id)
        pollIntervals.delete(sessionId)
        if (get().currentSessionId === sessionId) {
          get().loadSession(sessionId)
        } else {
          set((s) => ({
            sessions: s.sessions.map((se) => se.id === sessionId ? { ...se, status: 0 } : se),
          }))
        }
      } else {
        // Defensive: ensure the local session reflects the in-progress status.
        // Any transition path that forgets to set status=1 is self-corrected here,
        // so the background UI appears even if the stream silently died.
        // Skip the set() when already 1 to avoid creating a new sessions array
        // (which would trigger a re-render) every 3s.
        const se = get().sessions.find((s) => s.id === sessionId)
        if (se && se.status !== 1) {
          set((s) => ({
            sessions: s.sessions.map((se) =>
              se.id === sessionId ? { ...se, status: 1 } : se
            ),
          }))
        }
      }
    } catch {
      // ignore polling errors
    }
  }, 3000)
  pollIntervals.set(sessionId, id)
}

export function stopPolling(sessionId: string) {
  const id = pollIntervals.get(sessionId)
  if (id) {
    clearInterval(id)
    pollIntervals.delete(sessionId)
  }
}

/** Shared SSE handler factory to eliminate duplication between sendMessage and recovery flow. */
function createStreamHandlers(
  sessionId: string,
  set: Parameters<StateCreator<ChatState>>[0],
  get: () => ChatState,
  isNewSession = false,
) {
  return {
    onReasoning(chunk: string) {
      set((s) => {
        // 收到实际内容，retry 已过期，清掉
        const segs = s.streamingThinkingSegments.filter((seg) => seg.type !== 'retry')
        const last = segs[segs.length - 1]
        if (last && last.type === 'reasoning') {
          segs[segs.length - 1] = { ...last, content: last.content + chunk }
        } else {
          segs.push({ type: 'reasoning', content: chunk })
        }
        return { streamingThinkingSegments: segs, streamingRetry: null }
      })
    },
    onContent(chunk: string) {
      set((s) => ({
        streamingContent: s.streamingContent + chunk,
        // 收到实际内容，retry 已过期，清掉
        streamingThinkingSegments: s.streamingThinkingSegments.some((seg) => seg.type === 'retry')
          ? s.streamingThinkingSegments.filter((seg) => seg.type !== 'retry') : s.streamingThinkingSegments,
        streamingRetry: null,
      }))
    },
    onTool(name: string, displayName: string, icon: string, action: string) {
      set((s) => ({
        // 收到实际内容，retry 已过期，清掉
        streamingThinkingSegments: [
          ...s.streamingThinkingSegments.filter((seg) => seg.type !== 'retry'),
          {
            type: 'tool' as const,
            toolName: name,
            displayName,
            icon,
            action,
          },
        ],
        streamingRetry: null,
      }))
    },
    onRetry(attempt: number, maxRetries: number, error: string) {
      set((s) => ({
        streamingThinkingSegments: [...s.streamingThinkingSegments, {
          type: 'retry' as const,
          attempt,
          maxRetries,
          error,
        }],
        streamingRetry: { attempt, maxRetries, error },
      }))
    },
    onContext(tokens: number, limit: number) {
      set((s) => ({
        sessionDetail: s.sessionDetail ? { ...s.sessionDetail, contextTokens: tokens, contextLimit: limit } : null,
        sessions: s.sessions.map((se) =>
          se.id === sessionId ? { ...se, contextTokens: tokens, contextLimit: limit } : se
        ),
      }))
    },
    onEnd() {
      clearTimeout(backgroundTimeouts.get(sessionId))
      backgroundTimeouts.delete(sessionId)
      stopPolling(sessionId)
      const s = get()
      // 持久化前过滤掉 retry（重试已过期，不需要保留在最终消息中）
      const persistSegs = s.streamingThinkingSegments.filter((seg) => seg.type !== 'retry')
      const thinkingSegments = persistSegs.length > 0 ? persistSegs : undefined
      const finalMsg: Message = {
        id: uuid(),
        role: 'assistant',
        content: s.streamingContent,
        reasoningContent: persistSegs
          .filter((seg): seg is { type: 'reasoning', content: string } => seg.type === 'reasoning')
          .map((seg) => seg.content)
          .join('') || undefined,
        toolCalls: persistSegs
          .filter((seg): seg is { type: 'tool', toolName: string, displayName: string, icon: string, action: string } => seg.type === 'tool')
          .map((seg) => ({
            toolName: seg.toolName,
            displayName: seg.displayName,
            icon: seg.icon,
            action: seg.action,
          })) || undefined,
        thinkingSegments,
        timestamp: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
        roundId: uuid(),
        contextState: 0,
      }

      set((prev) => ({
        messages: [...prev.messages, finalMsg],
        generatingSessionId: null,
        streamingContent: '',
        streamingThinkingSegments: [],
        streamingRetry: null,
        streamController: null,
        sessions: prev.sessions.map((s) => (s.id === sessionId ? { ...s, status: 0 } : s)),
      }))

      // Refresh session detail to get updated context tokens
      sessionsApi.getSessionDetail(sessionId)
        .then((detail) => set({ sessionDetail: detail }))
        .catch(() => {})

      // 新会话首轮结束后，后端异步生成标题，延迟 15s 刷新侧边栏
      if (isNewSession) {
        setTimeout(() => get().refreshSessions(), 15_000)
      }
    },
    onError() {
      clearTimeout(backgroundTimeouts.get(sessionId))
      backgroundTimeouts.delete(sessionId)
      const state = get()
      const wasGenerating = state.generatingSessionId === sessionId
      set((prev) => {
        // If already in background (timeout fired before), just clean up stream state
        if (prev.generatingSessionId !== sessionId) {
          return { generatingSessionId: null, streamController: null }
        }
        // SSE disconnected (network issue) — mark status=1 so the UI shows the
        // background state (session.status===1 && !isGenerating) while polling
        // waits for the backend to finish. Clear stream state.
        return {
          generatingSessionId: null,
          streamController: null,
          streamingContent: '',
          streamingThinkingSegments: [],
          streamingRetry: null,
          sessions: prev.sessions.map((se) =>
            se.id === sessionId ? { ...se, status: 1 } : se
          ),
        }
      })
      // Start polling so we detect when the backend finishes generating
      if (wasGenerating) {
        startPolling(sessionId, get, set)
      }
    },
  }
}

/* ============================================
   Store
   ============================================ */

interface ChatState {
  models: Model[]
  mcpEndpoints: McpEndpoint[]
  skills: Skill[]
  personas: Persona[]
  sessions: Session[]
  sessionPage: number
  sessionHasMore: boolean

  formPersonaId: number | null
  formModelId: number
  formMcpIds: number[]
  formSkillIds: number[]
  formThinking: boolean
  formProjectPath: string | null
  formSharedProjectId: number | null
  formAttachments: string[]
  formRefs: FileRef[]

  currentSessionId: string | null
  sessionDetail: SessionDetail | null
  messages: Message[]
  generatingSessionId: string | null
  compressing: boolean
  streamingContent: string
  streamingThinkingSegments: ThinkingSegment[]
  streamingRetry: RetryInfo | null
  streamController: AbortController | null
  input: string
  creatingSession: boolean
  stopDisabled: boolean

  setFormPersonaId: (id: number | null) => void
  setFormModelId: (id: number) => void
  toggleFormMcp: (id: number) => void
  toggleFormSkill: (id: number) => void
  setFormThinking: (v: boolean) => void
  setFormProjectPath: (path: string | null) => void
  setFormSharedProjectId: (id: number | null) => void
  addFormAttachment: (uploadId: string) => void
  removeFormAttachment: (uploadId: string) => void
  addFormRef: (ref: FileRef) => void
  removeFormRef: (index: number) => void

  loadModels: () => Promise<void>
  loadPersonas: () => Promise<void>
  loadMcpEndpoints: () => Promise<void>
  loadSkills: () => Promise<void>

  sendMessage: (content: string) => Promise<string | undefined>
  loadSession: (id: string) => Promise<void>
  stopGeneration: () => void
  compressSession: () => void
  refreshSessions: () => Promise<void>
  loadMoreSessions: () => Promise<void>
  setInput: (v: string) => void
  backgroundActiveSession: (except?: string) => void
}

export const useChatStore = create<ChatState>((set, get) => ({
  models: [],
  mcpEndpoints: [],
  skills: [],
  personas: [],
  sessions: [],
  sessionPage: 1,
  sessionHasMore: false,

  formPersonaId: null,
  formModelId: 0,
  formMcpIds: [],
  formSkillIds: [],
  formThinking: true,
  formProjectPath: null,
  formSharedProjectId: null,
  formAttachments: [],
  formRefs: [],

  currentSessionId: null,
  sessionDetail: null,
  messages: [],
  generatingSessionId: null,
  compressing: false,
  streamingContent: '',
  streamingThinkingSegments: [],
  streamingRetry: null,
  streamController: null,
  input: '',
  creatingSession: false,
  stopDisabled: false,

  setFormPersonaId: async (id) => {
    if (id === null) {
      set({ formPersonaId: null, formMcpIds: [], formSkillIds: [] })
      return
    }
    set({ formPersonaId: id })
    try {
      const detail = await personasApi.getPersonaDetail(id)
      const modelId = detail.aiModelId > 0 ? detail.aiModelId : get().formModelId
      set({
        formModelId: modelId,
        formMcpIds: detail.mcpIds ?? [],
        formSkillIds: detail.skillIds ?? [],
      })
    } catch {
      // keep current values on error
    }
  },

  setFormModelId: (id) => set({ formModelId: id }),

  toggleFormMcp: (id) =>
    set((s) => {
      if (s.formPersonaId) return s
      return {
        formMcpIds: s.formMcpIds.includes(id)
          ? s.formMcpIds.filter((i) => i !== id)
          : [...s.formMcpIds, id],
      }
    }),

  toggleFormSkill: (id) =>
    set((s) => {
      if (s.formPersonaId) return s
      return {
        formSkillIds: s.formSkillIds.includes(id)
          ? s.formSkillIds.filter((i) => i !== id)
          : [...s.formSkillIds, id],
      }
    }),

  setFormThinking: (v) => set({ formThinking: v }),

  setFormProjectPath: (path) => set({ formProjectPath: path, formSharedProjectId: null }),

  setFormSharedProjectId: (id) => set({ formSharedProjectId: id, formProjectPath: null }),

  addFormAttachment: (uploadId) =>
    set((s) => ({
      formAttachments: s.formAttachments.includes(uploadId)
        ? s.formAttachments
        : [...s.formAttachments, uploadId],
    })),

  removeFormAttachment: (uploadId) =>
    set((s) => ({
      formAttachments: s.formAttachments.filter((id) => id !== uploadId),
    })),

  addFormRef: (ref) =>
    set((s) => ({
      formRefs: [...s.formRefs, ref],
    })),

  removeFormRef: (index) =>
    set((s) => ({
      formRefs: s.formRefs.filter((_, i) => i !== index),
    })),

  loadModels: async () => {
    try {
      const data = await providersApi.getAvailableModels()
      const models: Model[] = data.map((m) => ({
        id: m.aiModelId,
        name: m.displayName,
        provider: m.providerDisplayName,
        icon: m.icon,
        isDefault: Boolean(m.isDefault),
        isFlash: Boolean(m.isFlash),
        isVisual: Boolean(m.isVisual),
      }))
      const defaultModel = models.find((m) => m.isDefault)
      let nextModelId = get().formModelId
      if (!nextModelId) {
        const lastId = getLastUsedModelId()
        nextModelId = lastId && models.some((m) => m.id === lastId)
          ? lastId
          : (defaultModel?.id ?? models[0]?.id ?? 0)
      }
      set({ models, formModelId: nextModelId })
    } catch {
      // keep empty models on error
    }
  },

  loadPersonas: async () => {
    try {
      const data = await personasApi.getPersonasSimple()
      const personas: Persona[] = data.map((p) => ({
        id: p.personaId,
        name: p.name,
        icon: p.icon,
      }))
      set({ personas })
    } catch {
      // keep empty personas on error
    }
  },

  loadMcpEndpoints: async () => {
    try {
      const data = await mcpApi.getMcpEndpoints()
      const endpoints: McpEndpoint[] = data.map((m) => ({
        id: m.mcpId,
        name: m.displayName,
        description: m.description,
        icon: m.icon,
      }))
      set({ mcpEndpoints: endpoints })
    } catch {
      // keep existing endpoints on error
    }
  },

  loadSkills: async () => {
    try {
      const data = await skillsApi.getSimpleSkills()
      const skills: Skill[] = data.map((s) => ({
        id: s.userSkillId,
        name: s.skillName,
        description: s.description,
        icon: s.icon,
      }))
      set({ skills })
    } catch {
      // keep existing skills on error
    }
  },

  sendMessage: async (content) => {
    const state = get()
    if (state.currentSessionId && state.generatingSessionId === state.currentSessionId) return undefined
    if (!state.currentSessionId && state.creatingSession) return undefined

    // Abort any existing stream before starting a new one
    if (state.streamController) {
      state.streamController.abort()
      // Clear background timeout for the generating session
      const genId = state.generatingSessionId
      if (genId) {
        clearTimeout(backgroundTimeouts.get(genId))
        backgroundTimeouts.delete(genId)
      }
      // If the existing stream is for a different session, clean up generating state
      // and start polling so it transitions to background mode
      if (genId && genId !== state.currentSessionId) {
        set({ generatingSessionId: null, streamController: null, streamingContent: '', streamingThinkingSegments: [] })
        startPolling(genId, get, set)
      }
    }

    // Optimistically add user message
    const userMsg: Message = {
      id: uuid(),
      role: 'user',
      content,
      timestamp: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
      roundId: uuid(),
      contextState: 0,
    }

    set((s) => ({ messages: [...s.messages, userMsg], streamController: null }))

    // Build the request
    const refTags = state.formRefs.map(r => `<goraven-ref type="${r.type}" name="${r.name}">${r.path}</goraven-ref>`).join('\n')
    const fullContent = content + (refTags ? '\n' + refTags : '')
    const request: ChatRequest = {
      sessionId: state.currentSessionId || undefined,
      content: fullContent,
      attachments: state.formAttachments,
      aiModelId: state.formModelId,
      personaId: state.formPersonaId || undefined,
      mcpIds: state.formMcpIds,
      skillIds: state.formSkillIds,
      reasoning: state.formThinking ? 1 : 0,
      project: state.formSharedProjectId ? '' : (state.formProjectPath || ''),
      sharedProjectId: state.formSharedProjectId || undefined,
    }

    let sessionId: string
    let rsp: ChatResponse
    const isNewSession = !state.currentSessionId
    if (isNewSession) {
      set({ creatingSession: true })
    } else {
      set({ generatingSessionId: state.currentSessionId })
    }
    set({ stopDisabled: true })
    setTimeout(() => set({ stopDisabled: false }), 3000)
    try {
      rsp = await chatApi.createChat(request)
      sessionId = rsp.sessionId
      set({ input: '' })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '发送失败')
      if (isNewSession) {
        set({ creatingSession: false })
      } else {
        set({ generatingSessionId: null })
      }
      // Remove the optimistic user message on error
      set((s) => ({ messages: s.messages.filter((m) => m.id !== userMsg.id) }))
      return undefined
    }
    if (isNewSession) {
      set({ creatingSession: false })
      const usedModelId = rsp.session?.aiModelId ?? state.formModelId
      if (usedModelId > 0) setLastUsedModelId(usedModelId)
    }


    // For new sessions, build session from chat response (backend returns full detail)
    if (isNewSession) {
      const d = rsp.session
      const newSession: Session = d ? {
        id: d.sessionId,
        title: d.title,
        modelId: d.aiModelId,
        personaId: d.personaId > 0 ? d.personaId : null,
        project: d.project,
        sharedProject: d.sharedProject,
        mcpIds: d.mcpIds ?? [],
        skillIds: d.skillIds ?? [],
        lastChatTime: d.lastChatTime,
        // Backend returns status=0 in the ChatRsp because session.Status is read
        // before the runner goroutine flips the DB to 1. We just initiated a
        // generation, so treat it as in-progress.
        status: 1,
        messages: [...state.messages],
        modelName: d.modelName ?? '',
        modelIcon: d.modelIcon ?? '',
        personaName: d.personaName ?? '',
        personaIcon: d.personaIcon ?? '',
        contextTokens: d.contextTokens,
        contextLimit: d.contextLimit,
        promptTokensCount: d.promptTokensCount,
        completionTokensCount: d.completionTokensCount,
      } : {
        id: sessionId,
        title: content.slice(0, 50),
        modelId: state.formModelId,
        personaId: state.formPersonaId,
        project: state.formProjectPath || '',
        sharedProject: undefined,
        mcpIds: state.formMcpIds,
        skillIds: state.formSkillIds,
        lastChatTime: new Date().toISOString(),
        status: 1,
        messages: [...state.messages],
        modelName: '',
        modelIcon: '',
        personaName: '',
        personaIcon: '',
        contextTokens: 0,
        contextLimit: 0,
        promptTokensCount: 0,
        completionTokensCount: 0,
      }

      set((s) => ({
        sessions: [newSession, ...s.sessions],
        currentSessionId: sessionId,
        generatingSessionId: sessionId,
        streamingContent: '',
        streamingThinkingSegments: [],
        streamingRetry: null,
        formAttachments: [],
        formRefs: [],
        sessionDetail: d ?? null,
      }))
    } else {
      set((s) => ({
        currentSessionId: sessionId,
        generatingSessionId: sessionId,
        streamingContent: '',
        streamingThinkingSegments: [],
        streamingRetry: null,
        formAttachments: [],
        formRefs: [],
        // Optimistically mark the session as generating so any transition to
        // background (timeout / SSE error / navigation) reads status===1.
        sessions: s.sessions.map((se) =>
          se.id === sessionId ? { ...se, status: 1 } : se
        ),
      }))
    }

    // Connect SSE stream
    const handlers = createStreamHandlers(sessionId, set, get, isNewSession)
    const controller = connectChatStream(sessionId, handlers)

    set({ streamController: controller })

    // 3-minute background timer: if no response, cut SSE and go to background
    clearTimeout(backgroundTimeouts.get(sessionId))
    const timeoutId = setTimeout(() => {
      const s = get()
      // Only transition if still actively generating for this session
      if (s.generatingSessionId !== sessionId) return
      set({
        generatingSessionId: null,
        streamController: null,
        sessions: s.sessions.map((se) =>
          se.id === sessionId ? { ...se, status: 1 } : se
        ),
      })
      controller.abort()
      startPolling(sessionId, get, set)
      backgroundTimeouts.delete(sessionId)
    }, BACKGROUND_TIMEOUT_MS)
    backgroundTimeouts.set(sessionId, timeoutId)

    // Clear input after sending
    set({ input: '' })

    // Return sessionId for navigation (only meaningful for new sessions)
    return isNewSession ? sessionId : undefined
  },

  loadSession: async (id) => {
    // Background any actively generating session (abort SSE, start polling)
    get().backgroundActiveSession(id)

    try {
      const [detail, apiMessages] = await Promise.all([
        sessionsApi.getSessionDetail(id),
        chatApi.getSessionMessages(id),
      ])

      const messages: Message[] = (apiMessages ?? []).map(mapApiMessage)

      const session: Session = {
        id: detail.sessionId,
        title: detail.title,
        modelId: detail.aiModelId,
        personaId: detail.personaId > 0 ? detail.personaId : null,
        project: detail.project,
        sharedProject: detail.sharedProject,
        mcpIds: detail.mcpIds ?? [],
        skillIds: detail.skillIds ?? [],
        lastChatTime: detail.lastChatTime,
        status: detail.status,
        messages,
        modelName: detail.modelName ?? '',
        modelIcon: detail.modelIcon ?? '',
        personaName: detail.personaName ?? '',
        personaIcon: detail.personaIcon ?? '',
        contextTokens: detail.contextTokens,
        contextLimit: detail.contextLimit,
        promptTokensCount: detail.promptTokensCount,
        completionTokensCount: detail.completionTokensCount,
      }

      console.warn(`[loadSession/${id}] Zustand set() — ${messages.length} messages ready`)
      set((s) => {
        const sessions = s.sessions.some((sess) => sess.id === id)
          ? s.sessions.map((sess) => sess.id === id ? session : sess)
          : [session, ...s.sessions]

        return {
          currentSessionId: id,
          sessionDetail: detail,
          messages,
          sessions,
          input: '',
          formThinking: true,
          streamController: s.generatingSessionId === id ? s.streamController : null,
          formAttachments: [],
          formRefs: [],
        }
      })

      // Start polling if session is generating in the background (no active SSE)
      if (detail.status === 1 && get().generatingSessionId !== id) {
        startPolling(id, get, set)
      }
    } catch {
      // On error, still set the sessionId so the UI can show an error state
      set({ currentSessionId: id, messages: [], input: '', streamController: null })
    }
  },

  stopGeneration: () => {
    const state = get()
    const controller = state.streamController

    // Background thinking: no active stream, just notify backend and let polling handle the rest
    if (!controller && !state.generatingSessionId) {
      if (state.currentSessionId) {
        chatApi.stopChat(state.currentSessionId).catch(() => {})
      }
      return
    }

    // Save partial streaming content as a terminated message before clearing
    // 持久化前过滤掉 retry（重试已过期，不需要保留在被中止的消息中）
    const persistSegs = state.streamingThinkingSegments.filter((seg) => seg.type !== 'retry')
    if (state.streamingContent || persistSegs.length > 0) {
      const partialMsg: Message = {
        id: uuid(),
        role: 'assistant',
        content: state.streamingContent,
        reasoningContent: persistSegs
          .filter((seg): seg is { type: 'reasoning', content: string } => seg.type === 'reasoning')
          .map((seg) => seg.content)
          .join('') || undefined,
        toolCalls: persistSegs
          .filter((seg): seg is { type: 'tool', toolName: string, displayName: string, icon: string, action: string } => seg.type === 'tool')
          .map((seg) => ({
            toolName: seg.toolName,
            displayName: seg.displayName,
            icon: seg.icon,
            action: seg.action,
          })) || undefined,
        thinkingSegments: persistSegs.length > 0 ? persistSegs : undefined,
        timestamp: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
        roundId: uuid(),
        contextState: 1,
      }
      set((prev) => ({
        messages: [...prev.messages, partialMsg],
      }))
    }

    // Clear background timeout if any
    clearTimeout(backgroundTimeouts.get(state.currentSessionId!))
    backgroundTimeouts.delete(state.currentSessionId!)

    // Reset all generating state
    set({
      generatingSessionId: null,
      streamController: null,
      streamingContent: '',
      streamingThinkingSegments: [],
      streamingRetry: null,
      sessions: state.sessions.map(s => s.id === state.currentSessionId ? { ...s, status: 0 } : s),
    })

    // Then abort the stream
    controller?.abort()

    if (state.currentSessionId) {
      chatApi.stopChat(state.currentSessionId).catch(() => {})
    }
  },

  compressSession: () => {
    const state = get()
    if (state.compressing || !state.currentSessionId) return
    const sessionId = state.currentSessionId
    set({ compressing: true })

    chatApi.compressChat(sessionId)
      .then((rsp) => {
        const taskId = rsp.taskId
        let attempts = 0
        const poll = () => {
          if (attempts++ > 60) {
            set({ compressing: false })
            return
          }
          chatApi.getCompressStatus(taskId)
            .then((status) => {
              if (status.status === 'done') {
                get().loadSession(sessionId).finally(() => {
                  set({ compressing: false })
                })
              } else if (status.status === 'failed') {
                set({ compressing: false })
              } else {
                setTimeout(poll, 1000)
              }
            })
            .catch(() => {
              set({ compressing: false })
            })
        }
        poll()
      })
      .catch(() => {
        set({ compressing: false })
      })
  },

  refreshSessions: async () => {
    try {
      const data = await sessionsApi.getSessions({ page: 1 })
      const existing = get().sessions
      const existingMap = new Map(existing.map((s) => [s.id, s]))

      const cutoff = new Date()
      cutoff.setDate(cutoff.getDate() - 30)
      cutoff.setHours(0, 0, 0, 0)

      const list = data.list ?? []
      const sessions: Session[] = []
      let hitCutoff = false

      for (const s of list) {
        if (new Date(s.lastChatTime) < cutoff) {
          hitCutoff = true
          break
        }
        const prev = existingMap.get(s.sessionId)
        sessions.push({
          id: s.sessionId,
          title: s.title,
          status: s.status,
          lastChatTime: s.lastChatTime,
          modelId: prev?.modelId ?? 0,
          personaId: s.personaId > 0 ? s.personaId : (prev?.personaId ?? null),
          project: s.project || prev?.project || '',
          sharedProject: s.sharedProject || prev?.sharedProject,
          mcpIds: prev?.mcpIds ?? [],
          skillIds: prev?.skillIds ?? [],
          messages: prev?.messages ?? [],
          modelName: prev?.modelName ?? '',
          modelIcon: prev?.modelIcon ?? '',
          personaName: prev?.personaName ?? '',
          personaIcon: prev?.personaIcon ?? '',
          contextTokens: prev?.contextTokens ?? 0,
          contextLimit: prev?.contextLimit ?? 0,
          promptTokensCount: prev?.promptTokensCount ?? 0,
          completionTokensCount: prev?.completionTokensCount ?? 0,
        })
      }

      set({ sessions, sessionPage: 1, sessionHasMore: !hitCutoff && data.page < data.totalPage })
    } catch(e) {
      console.error('refreshSessions failed:', e)
    }
  },

  loadMoreSessions: async () => {
    const { sessionPage, sessionHasMore } = get()
    if (!sessionHasMore) return
    try {
      const nextPage = sessionPage + 1
      const data = await sessionsApi.getSessions({ page: nextPage })

      const cutoff = new Date()
      cutoff.setDate(cutoff.getDate() - 30)
      cutoff.setHours(0, 0, 0, 0)

      const list = data.list ?? []
      const sessions: Session[] = []
      let hitCutoff = false

      for (const s of list) {
        if (new Date(s.lastChatTime) < cutoff) {
          hitCutoff = true
          break
        }
        sessions.push({
          id: s.sessionId,
          title: s.title,
          status: s.status,
          lastChatTime: s.lastChatTime,
          modelId: 0,
          personaId: s.personaId > 0 ? s.personaId : null,
          project: s.project || '',
          sharedProject: s.sharedProject,
          mcpIds: [],
          skillIds: [],
          messages: [],
          modelName: '',
          modelIcon: '',
          personaName: '',
          personaIcon: '',
          contextTokens: 0,
          contextLimit: 0,
          promptTokensCount: 0,
          completionTokensCount: 0,
        })
      }

      set({
        sessions: [...get().sessions, ...sessions],
        sessionPage: nextPage,
        sessionHasMore: !hitCutoff && nextPage < data.totalPage,
      })
    } catch(e) {
      console.error('loadMoreSessions failed:', e)
    }
  },

  setInput: (v) => set({ input: v }),

  backgroundActiveSession: (except) => {
    const genId = get().generatingSessionId
    if (!genId || genId === except) return
    clearTimeout(backgroundTimeouts.get(genId))
    backgroundTimeouts.delete(genId)
    get().streamController?.abort()
    set((s) => ({
      generatingSessionId: null,
      streamController: null,
      streamingContent: '',
      streamingThinkingSegments: [],
      // The backgrounded session is still generating on the backend; reflect
      // that locally so the UI renders the background state while polling.
      sessions: s.sessions.map((se) =>
        se.id === genId ? { ...se, status: 1 } : se
      ),
    }))
    startPolling(genId, get, set)
  },
}))
