import { useState, useCallback, useRef, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Share2, Shrink, Paperclip, ArrowUp, CircleStop, Loader2, TriangleAlert, FolderGit2, FolderOpen, FolderClock, Menu } from 'lucide-react'
import { cn } from '@/lib/utils'
import { listFiles } from '@/api/files'
import type { FileItem } from '@/api/types'
import { useChatStore, stopPolling, type Model, type McpEndpoint, type Skill } from '@/stores/chat-store'
import { useSidebarStore } from '@/stores/sidebar-store'
import { ShareDialog } from './ShareDialog'
import { MessageList } from './MessageList'
import { FilePreviews, MAX_FILE_SIZE } from './FilePreviews'
import type { UploadedFile } from './FilePreviews'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
} from '@/components/ui/dialog'
import { Icon } from '@/components/common/Icon'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { useChunkUpload } from '@/hooks/useChunkUpload'
import { useT } from '@/i18n'
import { sessionsApi } from '@/api'
import { PersonaInfoDialog } from './PersonaInfoDialog'
import { ModelInfoDialog } from './ModelInfoDialog'

function formatTokenK(n: number) {
  if (n < 1024) return String(n)
  const v = n / 1024
  return `${v % 1 === 0 ? v : v.toFixed(1)}K`
}

import type { Persona } from '@/stores/chat-store'

const SCROLL_THRESHOLD = 80

function useAutoScroll(deps: unknown[], resetKey: string | null = null) {
  const elRef = useRef<HTMLDivElement | null>(null)
  const userScrolledUp = useRef(false)
  const instanceId = useRef(Math.random().toString(36).slice(2, 7)).current

  // Reset scroll flag on session switch so the first load auto-scrolls.
  const prevKeyRef = useRef<string | null>(null)
  useEffect(() => {
    if (resetKey !== null && resetKey !== prevKeyRef.current) {
      console.warn(`[scroll/${instanceId}] resetKey changed: "${prevKeyRef.current}" → "${resetKey}" — resetting userScrolledUp to false`)
      userScrolledUp.current = false
    } else {
      console.log(`[scroll/${instanceId}] resetKey effect: resetKey="${resetKey}" prevKey="${prevKeyRef.current}" — no reset`)
    }
    prevKeyRef.current = resetKey
  })

  // Callback ref — notified when scroll container mounts/unmounts.
  // When it mounts (e.g. after loading → content), scroll to bottom immediately.
  const ref = useCallback((el: HTMLDivElement | null) => {
    elRef.current = el
    const wasMounted = !!el
    console.warn(`[scroll/${instanceId}] callback ref: el=${wasMounted ? `mounted (scrollHeight=${el?.scrollHeight})` : 'unmounted'}, userScrolledUp=${userScrolledUp.current}`)
    if (el && !userScrolledUp.current) {
      console.warn(`[scroll/${instanceId}] callback ref → scrollTo(instant, ${el.scrollHeight})`)
      el.scrollTo({ top: el.scrollHeight, behavior: 'instant' as ScrollBehavior })
    } else if (el && userScrolledUp.current) {
      console.warn(`[scroll/${instanceId}] callback ref BLOCKED — userScrolledUp=true, NOT scrolling`)
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [resetKey])

  useEffect(() => {
    const el = elRef.current
    if (!el) return
    const onScroll = () => {
      const gap = el.scrollHeight - el.scrollTop - el.clientHeight
      const scrolledUp = gap > SCROLL_THRESHOLD
      if (userScrolledUp.current !== scrolledUp) {
        console.log(`[scroll/${instanceId}] scroll handler: gap=${gap} → userScrolledUp=${scrolledUp}`)
      }
      userScrolledUp.current = scrolledUp
    }
    el.addEventListener('scroll', onScroll, { passive: true })
    return () => el.removeEventListener('scroll', onScroll)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ref])

  useEffect(() => {
    const el = elRef.current
    console.log(`[scroll/${instanceId}] deps-scroll effect: el=${el ? `present (scrollHeight=${el.scrollHeight}, messages=${(deps[0] as []).length})` : 'null'}, userScrolledUp=${userScrolledUp.current}`)
    if (!el || userScrolledUp.current) return
    console.log(`[scroll/${instanceId}] deps-scroll → scrollTo(smooth, ${el.scrollHeight})`)
    el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' })
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps)

  return ref
}

/* ============================================
   Component
   ============================================ */

export function Component() {
  const { sessionId } = useParams<{ sessionId: string }>()

  return sessionId ? (
    <SessionChat sessionId={sessionId} />
  ) : (
    <NewChat />
  )
}

/* ============================================
   New Chat (no session)
   ============================================ */

function NewChat() {
  const navigate = useNavigate()
  const {
    personas, models, mcpEndpoints, skills,
    formPersonaId, formModelId, formMcpIds, formSkillIds,
    setFormPersonaId, setFormModelId, toggleFormMcp, toggleFormSkill,
    messages, sendMessage, input, setInput,
    streamingContent, streamingThinkingSegments,
    loadModels, loadPersonas, loadMcpEndpoints, loadSkills,
    formProjectPath, setFormProjectPath, generatingSessionId, creatingSession, stopDisabled,
  } = useChatStore()
  const generating = generatingSessionId !== null || creatingSession

  useEffect(() => {
    useChatStore.getState().backgroundActiveSession()
    useChatStore.setState({
      messages: [],
      currentSessionId: null,
      input: '',
      formThinking: true,
      formAttachments: [],
      formPersonaId: null,
      formMcpIds: [],
      formSkillIds: [],
      formProjectPath: null,
      formModelId: 0,
    })
    loadModels()
    loadPersonas()
    loadMcpEndpoints()
    loadSkills()
  }, [loadModels, loadPersonas, loadMcpEndpoints, loadSkills])

  const [plusOpen, setPlusOpen] = useState(false)
  const scrollContainerRef = useAutoScroll([messages, streamingContent, streamingThinkingSegments])
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const plusRef = useRef<HTMLDivElement>(null)
  const isComposingRef = useRef({ composing: false, justEnded: false })
  const personaLocked = formPersonaId !== null
  const chatting = messages.length > 0

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (plusRef.current && !plusRef.current.contains(e.target as Node)) {
        setPlusOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  const handleSend = useCallback(async () => {
    const trimmed = input.trim()
    if (!trimmed || generating) return
    const newSessionId = await sendMessage(trimmed)
    if (newSessionId) {
      navigate(`/chat/${newSessionId}`, { replace: true })
    }
  }, [input, generating, navigate, sendMessage])

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      const { composing, justEnded } = isComposingRef.current
      if (composing || justEnded || e.nativeEvent.isComposing || e.keyCode === 229) {
        isComposingRef.current.justEnded = false
        e.preventDefault()
        return
      }
      e.preventDefault()
      handleSend()
    } else if (!e.nativeEvent.isComposing) {
      isComposingRef.current.justEnded = false
    }
  }, [handleSend])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      <ChatToolbar
        personas={personas}
        models={models}
        formPersonaId={formPersonaId}
        formModelId={formModelId}
        onPersonaChange={setFormPersonaId}
        onModelChange={setFormModelId}
      />

      <div ref={scrollContainerRef} className="flex-1 overflow-y-auto [scrollbar-gutter:stable] [scrollbar-width:thin] [scrollbar-color:var(--color-bg-layer-3)_transparent] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-bg-layer-3 [&::-webkit-scrollbar-track]:bg-transparent">
        {!chatting ? (
          <div className="flex h-full flex-col items-center justify-center px-4 pb-20">
            <div className="mb-8 flex w-full max-w-2xl flex-col items-center gap-3">
              <img src="/favicon.svg" alt="Raven" className="size-10 opacity-60" />
              <p className="text-xl text-text-2">
                What can I help you build?
              </p>
            </div>
            <div className="w-full max-w-2xl">
              <NewChatInput
                value={input}
                onChange={setInput}
                onSend={handleSend}
                onKeyDown={handleKeyDown}
                generating={generating}
                plusOpen={plusOpen}
                onPlusToggle={() => setPlusOpen(!plusOpen)}
                plusRef={plusRef}
                personaLocked={personaLocked}
                mcpEndpoints={mcpEndpoints}
                skills={skills}
                formMcpIds={formMcpIds}
                formSkillIds={formSkillIds}
                onMcpToggle={toggleFormMcp}
                onSkillToggle={toggleFormSkill}
                formProjectPath={formProjectPath}
                onProjectChange={setFormProjectPath}
                stopDisabled={stopDisabled}
                isComposingRef={isComposingRef}
              />
            </div>
          </div>
        ) : (
        <MessageList messages={messages} isGenerating={generating} isBackground={false} />
        )}
        <div ref={messagesEndRef} />
      </div>

      {chatting && (
        <ChatInput
          value={input}
          onChange={setInput}
          onSend={handleSend}
          onKeyDown={handleKeyDown}
          isGenerating={generating}
          isBackground={false}
          stopDisabled={stopDisabled}
          isComposingRef={isComposingRef}
        />
      )}
    </div>
  )
}

/* ============================================
   Session Chat (existing session)
   ============================================ */

function SessionChat({ sessionId }: { sessionId: string }) {
  const { sessions, loadSession, messages, compressing, compressSession,
    streamingContent, streamingThinkingSegments,
    models, loadModels, sessionDetail, stopGeneration, input, setInput,
    generatingSessionId, sendMessage, stopDisabled,
  } = useChatStore()
  const t = useT()
  const session = sessions.find(s => s.id === sessionId)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const [loading, setLoading] = useState(true)
  const [personaDialogOpen, setPersonaDialogOpen] = useState(false)
  const [modelDialogOpen, setModelDialogOpen] = useState(false)
  const isComposingRef = useRef({ composing: false, justEnded: false })
  const isGenerating = generatingSessionId === sessionId
  const isBackground = session?.status === 1 && !isGenerating
  const personaId = session?.personaId ?? sessionDetail?.personaId ?? null

  useEffect(() => {
    const state = useChatStore.getState()
    if (state.currentSessionId === sessionId && state.messages.length > 0) {
      console.log(`[SessionChat/${sessionId}] loadSession effect: skipping, session already active (${state.messages.length} msgs)`)
      setLoading(false)
      return
    }
    console.log(`[SessionChat/${sessionId}] loadSession effect: fetching…`)
    setLoading(true)
    loadSession(sessionId).then(() => {
      console.log(`[SessionChat/${sessionId}] loadSession resolved, store now has ${useChatStore.getState().messages.length} messages`)
    }).finally(() => {
      console.warn(`[SessionChat/${sessionId}] .finally() → setLoading(false) — scroll container will mount next render`)
      setLoading(false)
    })
  }, [sessionId, loadSession])

  // Fallback: 会话关联的模型可能已被删除（sessionDetail.modelName 为空），
  // 此时拉取模型列表，取默认模型展示在工具栏；若未设置默认模型则取列表第一个。
  useEffect(() => {
    if (!sessionDetail || sessionDetail.modelName) return
    if (models.length > 0) return
    loadModels()
  }, [sessionDetail, models.length, loadModels])

  // Cleanup: stop polling for this session on unmount — but only if the
  // session has finished (status 0). If it's still in background (status 1),
  // keep polling so it transitions to done even after the user navigates away.
  // Do NOT destroy stream state (streamController, generatingSessionId, streaming*)
  // here — those live in the global Zustand store and are managed by sendMessage /
  // loadSession / stopGeneration. Destroying them on unmount breaks React StrictMode
  // double-mount and HMR, killing the active SSE mid-stream.
  useEffect(() => {
    return () => {
      const state = useChatStore.getState()
      const session = state.sessions.find(s => s.id === sessionId)
      if (!session || session.status === 0) {
        stopPolling(sessionId)
      }
    }
  }, [sessionId])

  const scrollContainerRef = useAutoScroll([messages, streamingContent, streamingThinkingSegments], sessionId)

  const handleSend = useCallback(async () => {
    const trimmed = input.trim()
    if (!trimmed || isGenerating || isBackground) return
    sendMessage(trimmed)
  }, [input, isGenerating, isBackground, sendMessage])

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      const { composing, justEnded } = isComposingRef.current
      if (composing || justEnded || e.nativeEvent.isComposing || e.keyCode === 229) {
        isComposingRef.current.justEnded = false
        e.preventDefault()
        return
      }
      e.preventDefault()
      handleSend()
    } else if (!e.nativeEvent.isComposing) {
      isComposingRef.current.justEnded = false
    }
  }, [handleSend])

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center bg-bg-base">
        <Loader2 className="size-5 animate-spin text-text-muted" />
      </div>
    )
  }

  if (!session) {
    return (
      <div className="flex h-full items-center justify-center bg-bg-base">
        <p className="text-sm text-text-3">{t('chat.sessionNotFound')}</p>
      </div>
    )
  }

  // 模型被删除时的降级：取默认模型，无默认则取列表第一个
  const fallbackModel = models.length > 0
    ? (models.find(m => m.isDefault) ?? models[0])
    : null
  const displayName = sessionDetail?.modelName || session.modelName || fallbackModel?.name || null
  const displayIcon = models.find(m => m.id === (sessionDetail?.aiModelId ?? session.modelId))?.icon || fallbackModel?.icon || null

  return (
    <div className="flex h-full flex-col bg-bg-base">
      <ChatToolbar
        showSession={true}
        sessionTitle={session.title}
        domainConfigured={true}
        modelName={displayName}
        modelIcon={displayIcon}
        personaName={sessionDetail?.personaName || session.personaName || null}
        personaIcon={sessionDetail?.personaIcon || session.personaIcon || null}
        onPersonaClick={() => setPersonaDialogOpen(true)}
        onModelClick={() => setModelDialogOpen(true)}
        tokenUsed={sessionDetail?.contextTokens ?? session.contextTokens}
        tokenMax={sessionDetail?.contextLimit ?? session.contextLimit}
        compressing={compressing}
        onCompress={compressSession}
        project={session.project}
        sessionId={sessionId}
      />

      <div ref={scrollContainerRef} className="flex-1 overflow-y-auto [scrollbar-gutter:stable] [scrollbar-width:thin] [scrollbar-color:var(--color-bg-layer-3)_transparent] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-bg-layer-3 [&::-webkit-scrollbar-track]:bg-transparent">
        <MessageList messages={messages} isGenerating={isGenerating} isBackground={isBackground} startTime={session.lastChatTime} />
        <div ref={messagesEndRef} />
      </div>

      <ChatInput
        value={input}
        onChange={setInput}
        onSend={handleSend}
        onKeyDown={handleKeyDown}
        isGenerating={isGenerating}
        isBackground={isBackground}
        compressing={compressing}
        onStop={stopGeneration}
        stopDisabled={stopDisabled}
        isComposingRef={isComposingRef}
      />
      <PersonaInfoDialog
        personaId={personaId}
        contextTokens={sessionDetail?.contextTokens ?? session.contextTokens ?? 0}
        promptTokensCount={sessionDetail?.promptTokensCount ?? session.promptTokensCount ?? 0}
        completionTokensCount={sessionDetail?.completionTokensCount ?? session.completionTokensCount ?? 0}
        open={personaDialogOpen}
        onOpenChange={setPersonaDialogOpen}
      />
      <ModelInfoDialog
        aiModelId={sessionDetail?.aiModelId ?? session.modelId ?? 0}
        mcpIds={sessionDetail?.mcpIds ?? []}
        skillIds={sessionDetail?.skillIds ?? []}
        contextTokens={sessionDetail?.contextTokens ?? session.contextTokens ?? 0}
        promptTokensCount={sessionDetail?.promptTokensCount ?? session.promptTokensCount ?? 0}
        completionTokensCount={sessionDetail?.completionTokensCount ?? session.completionTokensCount ?? 0}
        open={modelDialogOpen}
        onOpenChange={setModelDialogOpen}
      />
    </div>
  )
}

/* ============================================
   Toolbar
   ============================================ */

function ModelIcon({ icon }: { icon?: string }) {
  const [err, setErr] = useState(false)
  if (!icon || err) return <Icon name="brain" className="size-4 shrink-0 text-text-2" />
  return <img src={icon} alt="" className="size-4 shrink-0 rounded object-cover" onError={() => setErr(true)} />
}

function ChatToolbar({
  showSession,
  modelName,
  modelIcon,
  personaName,
  personaIcon,
  onPersonaClick,
  sessionTitle,
  domainConfigured,
  personas,
  models,
  formPersonaId,
  formModelId,
  onPersonaChange,
  onModelChange,
  onModelClick,
  tokenUsed,
  tokenMax,
  compressing,
  onCompress,
  project,
  sessionId,
}: {
  showSession?: boolean
  modelName?: string | null
  modelIcon?: string | null
  personaName?: string | null
  personaIcon?: string | null
  onPersonaClick?: () => void
  onModelClick?: () => void
  sessionTitle?: string
  domainConfigured?: boolean
  personas?: Persona[]
  models?: Model[]
  formPersonaId?: number | null
  formModelId?: number
  onPersonaChange?: (id: number | null) => void
  onModelChange?: (id: number) => void
  tokenUsed?: number
  tokenMax?: number
  compressing?: boolean
  onCompress?: () => void
  project?: string
  sessionId?: string
}) {
  const t = useT()
  const openMobile = useSidebarStore((s) => s.openMobile)
  const [open, setOpen] = useState(false)
  const [shareOpen, setShareOpen] = useState(false)
  const [noModelOpen, setNoModelOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const selectedPersona = formPersonaId ? personas?.find(p => p.id === formPersonaId) : null
  const selectedModel = !selectedPersona && formModelId ? models?.find(m => m.id === formModelId) : null

  const handleSelectModel = (id: number) => {
    if (onPersonaChange) onPersonaChange(null)
    if (onModelChange) onModelChange(id)
    setOpen(false)
  }

  const handleSelectPersona = (id: number) => {
    if (onPersonaChange) onPersonaChange(id)
    setOpen(false)
  }

  if (showSession) {
    return (
      <div className="flex h-10 shrink-0 items-center gap-3 border-b border-border px-4">
        <button
          onClick={openMobile}
          aria-label={t('sidebar.menu')}
          className="shrink-0 text-text-muted transition-colors hover:text-text-3 md:hidden"
        >
          <Menu className="size-4" />
        </button>
        {personaName ? (
          <button
            onClick={onPersonaClick}
            className="flex items-center gap-1.5 text-base text-text-2 min-w-0 rounded px-1 -mx-1 py-0.5 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
          >
            {personaIcon && <Icon name={personaIcon} className="size-4 shrink-0" />}
            <span className="truncate max-w-48">{personaName}</span>
          </button>
        ) : (
          <button
            onClick={onModelClick}
            className="flex items-center gap-1.5 text-base text-text-2 min-w-0 rounded px-1 -mx-1 py-0.5 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
          >
            <ModelIcon icon={modelIcon ?? undefined} />
            <span className="truncate max-w-48">{modelName}</span>
          </button>
        )}

        <div className="flex-1" />

        <div className="hidden items-center gap-2 md:flex">
          {project ? (
            <div className="flex items-center gap-1.5 max-w-36 text-xs text-highlight">
              <FolderGit2 className="size-3.5 shrink-0" />
              <span className="truncate">{project}</span>
            </div>
          ) : null}
          {tokenUsed !== undefined && tokenMax !== undefined && tokenMax > 0 && (
            <span className="text-xs text-highlight tabular-nums">
              {formatTokenK(tokenUsed)} / {formatTokenK(tokenMax)}
            </span>
          )}
          {compressing ? (
            <button
              disabled
              className="rounded p-1 cursor-not-allowed text-text-muted/50"
            >
              <Loader2 className="size-4 animate-spin" />
            </button>
          ) : (
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  onClick={onCompress}
                  className="rounded p-1 text-highlight transition-colors hover:bg-bg-layer-2 hover:opacity-90"
                >
                  <Shrink className="size-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent>
                {t('chat.compressTooltip')}
              </TooltipContent>
            </Tooltip>
          )}
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                onClick={() => setShareOpen(true)}
                className="rounded p-1 text-highlight transition-colors hover:bg-bg-layer-2 hover:opacity-90"
              >
                <Share2 className="size-4" />
              </button>
            </TooltipTrigger>
            <TooltipContent>{t('chat.shareTooltip')}</TooltipContent>
          </Tooltip>
        </div>

        {sessionTitle && sessionId && (
          <ShareDialog
            open={shareOpen}
            onOpenChange={setShareOpen}
            sessionTitle={sessionTitle}
            domainConfigured={domainConfigured}
            onGenerate={async (params) => {
              const share = await sessionsApi.createShare(sessionId, {
                title: params.title,
                expiresIn: params.expiresIn,
              })
              return `${window.location.origin}/share/${share.shareId}`
            }}
          />
        )}
      </div>
    )
  }

  return (
    <div className="flex h-10 shrink-0 items-center border-b border-border px-4">
      <button
        onClick={openMobile}
        aria-label={t('sidebar.menu')}
        className="shrink-0 text-text-muted transition-colors hover:text-text-3 md:hidden"
      >
        <Menu className="size-4" />
      </button>
      {!models || models.length === 0 ? (
        <>
          <button
            onClick={() => setNoModelOpen(true)}
            className="flex items-center gap-1.5 rounded px-2.5 py-0.5 text-sm text-text-muted transition-colors hover:bg-bg-layer-2 hover:text-text-3"
          >
            {t('chat.noModel')}
          </button>
          <Dialog open={noModelOpen} onOpenChange={setNoModelOpen}>
            <DialogContent className="max-w-md gap-0">
              <div className="rounded-lg border border-border px-4 py-5">
                <div className="flex items-center gap-2.5 mb-3">
                  <TriangleAlert className="size-5 shrink-0 text-highlight" />
                  <span className="text-base font-semibold text-text-1">{t('chat.noModelTitle')}</span>
                </div>
                <p className="text-sm text-text-2 leading-relaxed mb-3">
                  {t('chat.noModelDesc')}
                </p>
                <ol className="space-y-1.5 text-sm text-text-2">
                  <li className="flex gap-2">
                    <span className="shrink-0 text-text-muted">1.</span>
                    <span>{t('chat.noModelStep1')}</span>
                  </li>
                  <li className="flex gap-2">
                    <span className="shrink-0 text-text-muted">2.</span>
                    <span>{t('chat.noModelStep2')}</span>
                  </li>
                  <li className="flex gap-2">
                    <span className="shrink-0 text-text-muted">3.</span>
                    <span>{t('chat.noModelStep3')}</span>
                  </li>
                  <li className="flex gap-2">
                    <span className="shrink-0 text-text-muted">4.</span>
                    <span>{t('chat.noModelStep4')}</span>
                  </li>
                </ol>
              </div>
              <div className="mt-6 flex items-center justify-end">
                <Button variant="outline" onClick={() => setNoModelOpen(false)}>
                  {t('chat.gotIt')}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </>
      ) : (
        <>
          <div ref={ref} className="relative">
            <button
              onClick={() => setOpen(!open)}
              className={cn(
                'flex items-center gap-1.5 rounded px-2.5 py-0.5 text-sm transition-colors',
                selectedPersona || selectedModel ? 'text-text-1 hover:bg-bg-layer-2' : 'text-text-3 hover:bg-bg-layer-2 hover:text-text-2',
                open && 'bg-bg-layer-2',
              )}>
              <span className="truncate max-w-48">
                {selectedPersona ? (
                  <span className="flex items-center gap-1.5"><Icon name={selectedPersona.icon} className="size-3.5" />{selectedPersona.name}</span>
                ) : selectedModel ? (
                  <span className="flex items-center gap-1.5"><ModelIcon icon={selectedModel.icon} />{selectedModel.name}</span>
                ) : t('chat.selectModelOrPersona')}
              </span>
              <span className="text-xs text-text-muted">{open ? '▴' : '▾'}</span>
            </button>

            {open && (
              <div className="absolute left-0 top-full z-40 mt-1 max-w-[calc(100vw-2rem)] rounded-lg border border-border bg-bg-layer-2 py-1 shadow-pop md:w-56">
                <div className="px-3 py-0.5 text-xs text-text-muted">{t('chat.modelsDropdown')}</div>
                {models?.map((m) => (
                  <button key={m.id} onClick={() => handleSelectModel(m.id)}
                    className={cn(
                      'flex w-full items-center gap-2 px-3 py-0.5 text-sm transition-colors hover:bg-bg-layer-3 min-w-0',
                      !selectedPersona && formModelId === m.id ? 'text-text-1' : 'text-text-2',
                    )}>
                    <span className={cn('shrink-0 text-sm', !selectedPersona && formModelId === m.id ? 'text-text-1' : 'text-text-muted opacity-0')}>
                      {'✓'}
                    </span>
                    <ModelIcon icon={m.icon} />
                    <span className="truncate">{m.name}</span>
                  </button>
                ))}
                <div className="my-1 h-px bg-border" />
                <div className="px-3 py-0.5 text-xs text-text-muted">{t('chat.personasDropdown')}</div>
                {personas?.map((p) => (
                  <button key={p.id} onClick={() => handleSelectPersona(p.id)}
                    className={cn(
                      'flex w-full items-center gap-2 px-3 py-0.5 text-sm transition-colors hover:bg-bg-layer-3 min-w-0',
                      formPersonaId === p.id ? 'text-text-1' : 'text-text-2',
                    )}>
                    <span className={cn('shrink-0 text-sm', formPersonaId === p.id ? 'text-text-1' : 'text-text-muted opacity-0')}>
                      {'✓'}
                    </span>
                    <Icon name={p.icon} className="size-3.5 shrink-0 text-text-2" />
                    <span className="truncate">{p.name}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}

/* ============================================
   Init Form (new conversation)
   ============================================ */

/* ============================================
   New Chat Input (with + popover for config)
   ============================================ */

function NewChatInput({
  value, onChange, onSend, onKeyDown, generating,
  plusOpen, onPlusToggle, plusRef,
  personaLocked,
  mcpEndpoints, skills,
  formMcpIds, formSkillIds,
  onMcpToggle, onSkillToggle,
  formProjectPath, onProjectChange,
  stopDisabled,
  isComposingRef,
}: {
  value: string
  onChange: (v: string) => void
  onSend: () => void
  onKeyDown: (e: React.KeyboardEvent) => void
  generating: boolean
  plusOpen: boolean
  onPlusToggle: () => void
  plusRef: React.RefObject<HTMLDivElement>
  personaLocked: boolean
  mcpEndpoints: McpEndpoint[]
  skills: Skill[]
  formMcpIds: number[]
  formSkillIds: number[]
  onMcpToggle: (id: number) => void
  onSkillToggle: (id: number) => void
  formProjectPath: string | null
  onProjectChange: (path: string | null) => void
  stopDisabled: boolean
  isComposingRef: React.MutableRefObject<{ composing: boolean; justEnded: boolean }>
}) {
  const t = useT()
  const thinking = useChatStore((s) => s.formThinking)
  const setThinking = useChatStore((s) => s.setFormThinking)
  const addFormAttachment = useChatStore((s) => s.addFormAttachment)
  const removeFormAttachment = useChatStore((s) => s.removeFormAttachment)
  const stopGeneration = useChatStore((s) => s.stopGeneration)
  const { upload: chunkUpload } = useChunkUpload()
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [files, setFiles] = useState<UploadedFile[]>([])
  const [projectOpen, setProjectOpen] = useState(false)
  const [projectItems, setProjectItems] = useState<FileItem[]>([])
  const [projectLoading, setProjectLoading] = useState(false)
  const [projectError, setProjectError] = useState(false)

  const loadProjects = useCallback(() => {
    setProjectLoading(true)
    setProjectError(false)
    listFiles('projects')
      .then((data) => {
        const dirs = data.items.filter((item) => item.isDir)
        setProjectItems(dirs)
        setProjectLoading(false)
      })
      .catch(() => {
        setProjectError(true)
        setProjectLoading(false)
      })
  }, [])

  const handleProjectOpen = useCallback(() => {
    setProjectOpen(true)
    loadProjects()
  }, [loadProjects])

  const handleProjectSelect = useCallback((name: string) => {
    if (formProjectPath === name) {
      onProjectChange(null)
    } else {
      onProjectChange(name)
    }
    setProjectOpen(false)
  }, [formProjectPath, onProjectChange])

  const sidebarCollapsed = useSidebarStore((s) => s.collapsed)
  const dialogLeft = sidebarCollapsed
    ? 'calc(50% + 32px)'
    : 'calc(50% + 128px)'

  const handleFiles = useCallback(async (fileList: FileList) => {
    const remaining = 10 - files.length
    const selected = Array.from(fileList)
      .filter((f) => f.size <= MAX_FILE_SIZE)
      .slice(0, remaining)
    if (selected.length === 0) return

    const newFiles: UploadedFile[] = selected.map((f) => ({
      id: crypto.randomUUID(),
      name: f.name,
      size: f.size,
      type: f.type,
      status: 'uploading' as const,
      previewUrl: f.type.startsWith('image/') ? URL.createObjectURL(f) : undefined,
    }))

    setFiles((prev) => [...prev, ...newFiles])

    for (const f of newFiles) {
      const sourceFile = selected.find((s) => s.name === f.name)
      if (!sourceFile) continue
      try {
        const result = await chunkUpload(sourceFile)
        addFormAttachment(result.uploadId)
        setFiles((prev) => prev.map((pf) => pf.id === f.id ? { ...pf, status: 'done' as const, uploadId: result.uploadId } : pf))
      } catch {
        setFiles((prev) => prev.map((pf) => pf.id === f.id ? { ...pf, status: 'error' as const } : pf))
      }
    }
  }, [files.length, chunkUpload, addFormAttachment])

  const removeFile = useCallback((id: string) => {
    setFiles((prev) => {
      const file = prev.find((f) => f.id === id)
      if (file?.uploadId) removeFormAttachment(file.uploadId)
      return prev.filter((f) => f.id !== id)
    })
  }, [removeFormAttachment])

  useEffect(() => {
    const el = textareaRef.current
    if (el) {
      el.style.height = 'auto'
      el.style.height = Math.min(el.scrollHeight, 160) + 'px'
    }
  }, [value])

  const handleSendOrStop = useCallback(() => {
    if (generating) {
      stopGeneration()
    } else {
      onSend()
    }
  }, [generating, stopGeneration, onSend])

  return (
    <div className="rounded-lg bg-bg-layer-2">
      <FilePreviews files={files} onRemove={removeFile} />
      <textarea
        ref={textareaRef}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={onKeyDown}
        onCompositionStart={() => { isComposingRef.current.composing = true; isComposingRef.current.justEnded = false }}
        onCompositionEnd={() => { isComposingRef.current.composing = false; isComposingRef.current.justEnded = true }}
        placeholder={t('chat.inputPlaceholder')}
        disabled={generating}
        rows={1}
        className={cn(
          'w-full resize-none overflow-y-auto rounded-t-lg border-0 bg-transparent px-3 pt-3 text-base text-text-1 placeholder:text-text-muted outline-none max-h-40 [scrollbar-width:thin] [scrollbar-color:var(--color-bg-layer-3)_transparent] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-bg-layer-3 [&::-webkit-scrollbar-track]:bg-transparent',
          generating && 'opacity-50',
        )}
      />

      <div className="flex items-center justify-between px-3 pb-2">
        <div className="flex items-center gap-1">
          <div ref={plusRef} className="relative">
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  onClick={personaLocked ? undefined : onPlusToggle}
                  className={cn(
                    'rounded-md p-1 transition-colors',
                    personaLocked
                      ? 'pointer-events-none text-text-muted/30'
                      : (formMcpIds.length > 0 || formSkillIds.length > 0)
                        ? 'text-text-1 bg-bg-hover'
                        : 'text-text-muted hover:bg-bg-hover hover:text-text-2',
                    plusOpen && !personaLocked && 'bg-bg-hover text-text-2',
                  )}
                >
                  <Icon name="wrench" className="size-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent>{t('chat.configTooltip')}</TooltipContent>
            </Tooltip>

            {plusOpen && !personaLocked && (
              <div className="fixed inset-x-0 bottom-0 z-50 max-h-80 overflow-y-auto rounded-t-lg border border-border bg-bg-layer-2 p-3 shadow-pop [scrollbar-width:thin] [scrollbar-color:var(--color-bg-layer-3)_transparent] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-bg-layer-3 [&::-webkit-scrollbar-track]:bg-transparent md:absolute md:inset-x-auto md:bottom-full md:left-0 md:mb-2 md:w-72 md:rounded-lg md:rounded-t-none">
                <div className="mb-3">
                  <label className="mb-1.5 block text-sm text-text-3">{t('chat.mcpTools')}</label>
                  <div className="space-y-0.5">
                    {mcpEndpoints.map((mcp) => {
                      const checked = formMcpIds.includes(mcp.id)
                      return (
                        <Tooltip key={mcp.id}>
                          <TooltipTrigger asChild>
                            <label className={cn(
                              'flex cursor-pointer items-center gap-2 rounded px-1.5 py-1 text-sm hover:bg-bg-hover min-w-0',
                              checked ? 'bg-bg-hover text-text-1' : 'text-text-2',
                            )}>
                              <input type="checkbox" checked={checked}
                                onChange={() => onMcpToggle(mcp.id)} className="size-3.5 shrink-0" />
                              <Icon name={mcp.icon} className={cn('size-3.5 shrink-0', checked ? 'text-text-1' : 'text-text-3')} />
                              <span className="truncate">{mcp.name}</span>
                            </label>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-xs">
                            {mcp.description}
                          </TooltipContent>
                        </Tooltip>
                      )
                    })}
                  </div>
                </div>

                <div>
                  <label className="mb-1.5 block text-sm text-text-3">{t('chat.skillsLabel')}</label>
                  <div className="space-y-0.5">
                    {skills.map((skill) => {
                      const checked = formSkillIds.includes(skill.id)
                      return (
                        <Tooltip key={skill.id}>
                          <TooltipTrigger asChild>
                            <label className={cn(
                              'flex cursor-pointer items-center gap-2 rounded px-1.5 py-1 text-sm hover:bg-bg-hover min-w-0',
                              checked ? 'bg-bg-hover text-text-1' : 'text-text-2',
                            )}>
                              <input type="checkbox" checked={checked}
                                onChange={() => onSkillToggle(skill.id)} className="size-3.5 shrink-0" />
                              <Icon name={skill.icon} className={cn('size-3.5 shrink-0', checked ? 'text-text-1' : 'text-text-3')} />
                              <span className="truncate">{skill.name}</span>
                            </label>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-xs">
                            {skill.description}
                          </TooltipContent>
                        </Tooltip>
                      )
                    })}
                  </div>
                </div>
              </div>
            )}
          </div>

          <input
            ref={fileInputRef}
            type="file"
            multiple
            className="hidden"
            onChange={(e) => {
              if (e.target.files) handleFiles(e.target.files)
              e.target.value = ''
            }}
          />
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                className="rounded-md p-1 text-text-muted transition-colors hover:bg-bg-hover hover:text-text-2"
                onClick={() => fileInputRef.current?.click()}
              >
                <Paperclip className="size-3.5" />
              </button>
            </TooltipTrigger>
            <TooltipContent>{t('chat.maxFilesHint')}</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <button
                onClick={handleProjectOpen}
                className={cn(
                  'rounded-md p-1 transition-colors',
                  formProjectPath
                    ? 'text-text-1 bg-bg-hover'
                    : 'text-text-muted hover:bg-bg-hover hover:text-text-2',
                )}
              >
                <FolderGit2 className="size-3.5" />
              </button>
            </TooltipTrigger>
            <TooltipContent>{t('chat.selectProject')}</TooltipContent>
          </Tooltip>

          <label className={cn(
            'flex items-center gap-1.5 rounded-md px-1.5 py-1 text-xs text-text-2 transition-colors',
            !generating && 'cursor-pointer hover:bg-bg-hover', thinking && 'text-text-1',
          )}>
            <span className={cn('relative inline-flex h-3.5 w-8 items-center rounded-full transition-colors',
              thinking ? 'bg-highlight' : 'bg-bg-hover')}>
              <span className={cn('inline-block h-3 w-3 rounded-full transition-transform',
                thinking ? 'bg-highlight-fg translate-x-5' : 'bg-text-muted')} />
            </span>
            <span className="hidden sm:inline">{t('chat.thinkingToggle')}</span>
            <input type="checkbox" checked={thinking}
              onChange={() => !generating && setThinking(!thinking)} className="sr-only" />
          </label>
        </div>

        <button onClick={handleSendOrStop} disabled={(!generating && !value.trim()) || (generating && stopDisabled)}
          className={cn(
            'flex items-center justify-center rounded-md p-1.5 transition-colors bg-highlight text-highlight-fg hover:opacity-90',
            !value.trim() && !generating && 'pointer-events-none bg-bg-layer-3 text-text-2 opacity-40',
            generating && stopDisabled && 'opacity-40 cursor-not-allowed',
          )}>
          {generating ? <CircleStop className="size-4" /> : <ArrowUp className="size-4" />}
        </button>
      </div>

      <Dialog open={projectOpen} onOpenChange={setProjectOpen}>
        <DialogContent
          className="!max-w-lg p-5 md:left-[var(--dialog-left)]"
          style={{ '--dialog-left': dialogLeft } as React.CSSProperties}
        >
          <p className="text-base font-semibold text-text-1 mb-3">{t('chat.projectsTitle')}</p>

          {projectLoading ? (
            <div className="flex items-center justify-center py-10">
              <Loader2 className="size-4 animate-spin text-text-muted" />
            </div>
          ) : projectError ? (
            <div className="flex flex-col items-center gap-3 py-8">
              <FolderClock className="size-7 text-text-3" />
              <p className="text-sm text-text-3">{t('common.loadFailed')}</p>
              <button
                onClick={loadProjects}
                className="rounded-md bg-bg-layer-2 px-3 py-1.5 text-xs text-text-1 transition-colors hover:bg-bg-hover"
              >
                {t('common.retry')}
              </button>
            </div>
          ) : projectItems.length === 0 ? (
            <div className="flex flex-col items-center gap-2 py-8">
              <FolderOpen className="size-7 text-text-3" />
              <p className="text-[13px] text-text-2">{t('chat.noProject')}</p>
              <p className="text-xs text-text-3">{t('chat.noProjectHint')}</p>
            </div>
          ) : (
            <div className="-mx-5 max-h-64 overflow-y-auto [scrollbar-width:thin] [scrollbar-color:var(--color-bg-layer-3)_transparent] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-bg-layer-3 [&::-webkit-scrollbar-track]:bg-transparent">
              <button
                onClick={() => { onProjectChange(null); setProjectOpen(false) }}
                className={cn(
                  'flex w-full items-center gap-2.5 px-5 py-2 text-[13px] transition-colors text-left',
                  !formProjectPath
                    ? 'bg-bg-layer-3 text-text-1'
                    : 'text-text-3 hover:bg-bg-hover hover:text-text-2',
                )}
              >
                <FolderGit2 className={cn(
                  'size-4 shrink-0',
                  !formProjectPath ? 'text-text-1' : 'text-text-3',
                )} />
                <span>{t('chat.clearProject')}</span>
                {!formProjectPath && (
                  <span className="ml-auto shrink-0 text-xs">{'✓'}</span>
                )}
              </button>
              {projectItems.map((item, i) => {
                const isSelected = formProjectPath === item.name
                return (
                  <button
                    key={item.name}
                    onClick={() => handleProjectSelect(item.name)}
                    className={cn(
                      'flex w-full items-center gap-2.5 px-5 py-2 text-[13px] transition-colors text-left',
                      isSelected
                        ? 'bg-bg-layer-3 text-text-1'
                        : i % 2 === 0
                          ? 'bg-bg-layer-1 hover:bg-bg-hover text-text-2'
                          : 'hover:bg-bg-hover text-text-2',
                    )}
                  >
                    <FolderOpen className={cn(
                      'size-4 shrink-0',
                      isSelected ? 'text-text-1' : 'text-text-3',
                    )} />
                    <span className="truncate">{item.name}</span>
                    {isSelected && (
                      <span className="ml-auto shrink-0 text-xs">{'✓'}</span>
                    )}
                  </button>
                )
              })}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}

/* ============================================
   Chat Input (existing session, no + popover)
   ============================================ */

function ChatInput({
  value, onChange, onSend, onKeyDown, isGenerating, isBackground, compressing, onStop, stopDisabled,
  isComposingRef,
}: {
  value: string
  onChange: (v: string) => void
  onSend: () => void
  onKeyDown: (e: React.KeyboardEvent) => void
  isGenerating: boolean
  isBackground: boolean
  compressing?: boolean
  onStop?: () => void
  stopDisabled?: boolean
  isComposingRef: React.MutableRefObject<{ composing: boolean; justEnded: boolean }>
}) {
  const t = useT()
  const thinking = useChatStore((s) => s.formThinking)
  const setThinking = useChatStore((s) => s.setFormThinking)
  const addFormAttachment = useChatStore((s) => s.addFormAttachment)
  const removeFormAttachment = useChatStore((s) => s.removeFormAttachment)
  const { upload: chunkUpload } = useChunkUpload()
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [files, setFiles] = useState<UploadedFile[]>([])

  const handleFiles = useCallback(async (fileList: FileList) => {
    const remaining = 10 - files.length
    const selected = Array.from(fileList)
      .filter((f) => f.size <= MAX_FILE_SIZE)
      .slice(0, remaining)
    if (selected.length === 0) return

    const newFiles: UploadedFile[] = selected.map((f) => ({
      id: crypto.randomUUID(),
      name: f.name,
      size: f.size,
      type: f.type,
      status: 'uploading' as const,
      previewUrl: f.type.startsWith('image/') ? URL.createObjectURL(f) : undefined,
    }))

    setFiles((prev) => [...prev, ...newFiles])

    for (const f of newFiles) {
      const sourceFile = selected.find((s) => s.name === f.name)
      if (!sourceFile) continue
      try {
        const result = await chunkUpload(sourceFile)
        addFormAttachment(result.uploadId)
        setFiles((prev) => prev.map((pf) => pf.id === f.id ? { ...pf, status: 'done' as const, uploadId: result.uploadId } : pf))
      } catch {
        setFiles((prev) => prev.map((pf) => pf.id === f.id ? { ...pf, status: 'error' as const } : pf))
      }
    }
  }, [files.length, chunkUpload, addFormAttachment])

  const removeFile = useCallback((id: string) => {
    setFiles((prev) => {
      const file = prev.find((f) => f.id === id)
      if (file?.uploadId) removeFormAttachment(file.uploadId)
      return prev.filter((f) => f.id !== id)
    })
  }, [removeFormAttachment])

  useEffect(() => {
    const el = textareaRef.current
    if (el) {
      el.style.height = 'auto'
      el.style.height = Math.min(el.scrollHeight, 160) + 'px'
    }
  }, [value])

  const handleSendOrStop = useCallback(() => {
    if ((isGenerating || isBackground) && onStop) {
      onStop()
    } else {
      onSend()
    }
  }, [isGenerating, isBackground, onStop, onSend])

  return (
    <div className="shrink-0 px-4 pb-6 pt-3">
      <div className="mx-auto max-w-3xl">
        <div className="overflow-hidden rounded-lg bg-bg-layer-2">
          <FilePreviews files={files} onRemove={removeFile} />
          <textarea
            ref={textareaRef}
            value={value} onChange={(e) => onChange(e.target.value)}
            onKeyDown={onKeyDown}
            onCompositionStart={() => { isComposingRef.current.composing = true; isComposingRef.current.justEnded = false }}
            onCompositionEnd={() => { isComposingRef.current.composing = false; isComposingRef.current.justEnded = true }}
            placeholder={isBackground ? t('chat.backgroundThinkingPlaceholder') : t('chat.inputPlaceholderSession')} disabled={isGenerating || isBackground || compressing} rows={1}
            className={cn(
              'w-full resize-none overflow-y-auto border-0 bg-transparent px-3 pt-3 text-base text-text-1 placeholder:text-text-muted outline-none max-h-40 [scrollbar-width:thin] [scrollbar-color:var(--color-bg-layer-3)_transparent] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-bg-layer-3 [&::-webkit-scrollbar-track]:bg-transparent',
              (isGenerating || isBackground || compressing) && 'opacity-50',
            )} />

          <div className="flex items-center justify-between px-3 pb-2">
            <div className="flex items-center gap-1">
              <input
                ref={fileInputRef}
                type="file"
                multiple
                className="hidden"
                onChange={(e) => {
                  if (e.target.files) handleFiles(e.target.files)
                  e.target.value = ''
                }}
              />
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    className={cn(
                      'rounded-md p-1 transition-colors',
                      compressing
                        ? 'cursor-not-allowed text-text-muted/30'
                        : 'text-text-muted hover:bg-bg-hover hover:text-text-2',
                    )}
                    onClick={() => !compressing && fileInputRef.current?.click()}
                  >
                    <Paperclip className="size-3.5" />
                  </button>
                </TooltipTrigger>
                <TooltipContent>{t('chat.maxFilesHint')}</TooltipContent>
              </Tooltip>

              <label className={cn(
                'flex items-center gap-1.5 rounded-md px-1.5 py-1 text-xs text-text-2 transition-colors',
                !isGenerating && !isBackground && !compressing && 'cursor-pointer hover:bg-bg-hover', thinking && 'text-text-1',
              )}>
                <span className={cn('relative inline-flex h-3.5 w-8 items-center rounded-full transition-colors',
                  thinking ? 'bg-highlight' : 'bg-bg-hover')}>
                  <span className={cn('inline-block h-3 w-3 rounded-full transition-transform',
                    thinking ? 'bg-highlight-fg translate-x-5' : 'bg-text-muted')} />
                </span>
                <span className="hidden sm:inline">{t('chat.thinkingToggle')}</span>
                <input type="checkbox" checked={thinking}
                  onChange={() => !isGenerating && !isBackground && !compressing && setThinking(!thinking)} className="sr-only" />
              </label>
            </div>

            <div className="flex items-center gap-3">
              <button onClick={handleSendOrStop} disabled={(!isGenerating && !isBackground && (compressing || !value.trim())) || ((isGenerating || isBackground) && stopDisabled)}
                className={cn(
                  'flex items-center justify-center rounded-md p-1.5 transition-colors',
                  compressing && !isGenerating && !isBackground ? 'bg-bg-layer-3 text-text-2' : 'bg-highlight text-highlight-fg hover:opacity-90',
                  !value.trim() && !isGenerating && !isBackground && !compressing && 'pointer-events-none bg-bg-layer-3 text-text-2 opacity-40',
                  (isGenerating || isBackground) && stopDisabled && 'opacity-40 cursor-not-allowed',
                )}>
                {(isGenerating || isBackground) ? <CircleStop className="size-4" /> : <ArrowUp className="size-4" />}
              </button>
            </div>
          </div>
        </div>
        <p className="mt-2 text-center text-xs text-text-muted/50">{t('chat.aiDisclaimer')}</p>
      </div>
    </div>
  )
}


