/* ============================================
   Admin Models — mock data and handlers
   ============================================ */

import type { AdminModelItem, ProviderItem, RecommendModelItem, PaginatedResponse } from '../types'
import { listDelay, itemDelay, mutationDelay, heavyDelay } from '../delay'

/* ============================================
   Constants
   ============================================ */

export const PAGE_SIZE = 20

/* ============================================
   Stored model (includes full apiKey internally)
   ============================================ */

interface StoredModel {
  aiModelId: number
  providerDisplayName: string
  displayName: string
  providerId: string
  modelName: string
  icon: string
  apiKey: string
  baseUrl: string
  proxyUrl: string
  contextLen: number
  extraFields: string
  isDefault: number
  isFlash: number
  isVisual: number
  status: number
  remark: string
  created: string
  updated: string
}

function maskApiKey(key: string): string {
  if (!key || key.length <= 9) return key
  return key.slice(0, 5) + '****' + key.slice(-4)
}

function storedToItem(m: StoredModel): AdminModelItem {
  return {
    aiModelId: m.aiModelId,
    providerDisplayName: m.providerDisplayName,
    displayName: m.displayName,
    providerId: m.providerId,
    modelName: m.modelName,
    icon: m.icon,
    apiKey: maskApiKey(m.apiKey),
    baseUrl: m.baseUrl,
    proxyUrl: m.proxyUrl,
    contextLen: m.contextLen,
    extraFields: m.extraFields,
    isDefault: m.isDefault,
    isFlash: m.isFlash,
    isVisual: m.isVisual,
    status: m.status,
    remark: m.remark,
    created: m.created,
    updated: m.updated,
  }
}

function storedToDetail(m: StoredModel): AdminModelItem {
  return {
    aiModelId: m.aiModelId,
    providerDisplayName: m.providerDisplayName,
    displayName: m.displayName,
    providerId: m.providerId,
    modelName: m.modelName,
    icon: m.icon,
    apiKey: m.apiKey,
    baseUrl: m.baseUrl,
    proxyUrl: m.proxyUrl,
    contextLen: m.contextLen,
    extraFields: m.extraFields,
    isDefault: m.isDefault,
    isFlash: m.isFlash,
    isVisual: m.isVisual,
    status: m.status,
    remark: m.remark,
    created: m.created,
    updated: m.updated,
  }
}

/* ============================================
   Generate mock models (16 models, aiModelId 1–16)
   ============================================ */

function generateMockModels(): StoredModel[] {
  return [
    {
      aiModelId: 1, providerDisplayName: 'DeepSeek', displayName: 'DeepSeek V3', providerId: 'deepseek', icon: '/logos/deepseek.svg', modelName: 'deepseek-chat',
      apiKey: 'sk-d1xxxxxxxx87f2', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 1, isFlash: 0, isVisual: 0,
      status: 1, remark: '', created: '2026-01-15T08:30:00Z', updated: '2026-03-20T14:22:00Z',
    },
    {
      aiModelId: 2, providerDisplayName: 'DeepSeek 代码', displayName: 'DeepSeek Coder', providerId: 'deepseek', icon: '/logos/deepseek.svg', modelName: 'deepseek-coder',
      apiKey: 'sk-d2xxxxxxxx3a91', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 1, remark: '', created: '2026-02-01T10:00:00Z', updated: '2026-02-01T10:00:00Z',
    },
    {
      aiModelId: 3, providerDisplayName: 'DeepSeek 推理', displayName: 'DeepSeek Reasoner', providerId: 'deepseek', icon: '/logos/deepseek.svg', modelName: 'deepseek-reasoner',
      apiKey: 'sk-d3xxxxxxxxbc22', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 1, isVisual: 0,
      status: 1, remark: '用于复杂推理任务', created: '2026-03-01T09:15:00Z', updated: '2026-04-10T16:30:00Z',
    },
    {
      aiModelId: 4, providerDisplayName: 'OpenAI', displayName: 'GPT-4o', providerId: 'openai', icon: '/logos/openai.svg', modelName: 'gpt-4o',
      apiKey: 'sk-o1xxxxxxxxee45', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 1,
      status: 1, remark: '多模态模型', created: '2026-01-20T11:00:00Z', updated: '2026-05-01T08:00:00Z',
    },
    {
      aiModelId: 5, providerDisplayName: 'OpenAI Mini', displayName: 'GPT-4o Mini', providerId: 'openai', icon: '/logos/openai.svg', modelName: 'gpt-4o-mini',
      apiKey: 'sk-o2xxxxxxxxdf56', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 1, remark: '轻量任务', created: '2026-02-10T13:00:00Z', updated: '2026-02-10T13:00:00Z',
    },
    {
      aiModelId: 6, providerDisplayName: 'OpenAI o3', displayName: 'o3-mini', providerId: 'openai', icon: '/logos/openai.svg', modelName: 'o3-mini',
      apiKey: 'sk-o3xxxxxxxxgh78', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 0, remark: '', created: '2026-03-15T08:00:00Z', updated: '2026-03-15T08:00:00Z',
    },
    {
      aiModelId: 7, providerDisplayName: 'Claude', displayName: 'Claude Sonnet 4', providerId: 'claude', icon: '/logos/claude.svg', modelName: 'claude-sonnet-4-20250514',
      apiKey: 'sk-c1xxxxxxxxij90', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 1, remark: '', created: '2026-02-20T15:00:00Z', updated: '2026-02-20T15:00:00Z',
    },
    {
      aiModelId: 8, providerDisplayName: 'Claude Opus', displayName: 'Claude Opus 4', providerId: 'claude', icon: '/logos/claude.svg', modelName: 'claude-opus-4-20250514',
      apiKey: 'sk-c2xxxxxxxxkl12', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 1, remark: '复杂分析', created: '2026-03-01T10:00:00Z', updated: '2026-03-01T10:00:00Z',
    },
    {
      aiModelId: 9, providerDisplayName: 'Claude Haiku', displayName: 'Claude Haiku 3.5', providerId: 'claude', icon: '/logos/claude.svg', modelName: 'claude-haiku-3-5-20241022',
      apiKey: 'sk-c3xxxxxxxxmn34', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 0, remark: '', created: '2026-03-10T09:00:00Z', updated: '2026-03-10T09:00:00Z',
    },
    {
      aiModelId: 10, providerDisplayName: '百炼', displayName: 'Qwen Turbo', providerId: 'bailian', icon: '/logos/bailian.svg', modelName: 'qwen-turbo-latest',
      apiKey: 'sk-b1xxxxxxxxop56', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 1, remark: '', created: '2026-01-25T16:00:00Z', updated: '2026-01-25T16:00:00Z',
    },
    {
      aiModelId: 11, providerDisplayName: '百炼 Plus', displayName: 'Qwen Plus', providerId: 'bailian', icon: '/logos/bailian.svg', modelName: 'qwen-plus-latest',
      apiKey: 'sk-b2xxxxxxxxqr78', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 1, remark: '', created: '2026-02-05T14:00:00Z', updated: '2026-02-05T14:00:00Z',
    },
    {
      aiModelId: 12, providerDisplayName: '百炼 Max', displayName: 'Qwen Max', providerId: 'bailian', icon: '/logos/bailian.svg', modelName: 'qwen-max-latest',
      apiKey: 'sk-b3xxxxxxxxst90', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 1, remark: '长文本专用', created: '2026-03-20T11:00:00Z', updated: '2026-03-20T11:00:00Z',
    },
    {
      aiModelId: 13, providerDisplayName: '本地 Ollama', displayName: 'Qwen 2.5 14B', providerId: 'ollama', icon: '/logos/ollama.svg', modelName: 'qwen2.5:14b',
      apiKey: '', baseUrl: 'http://localhost:11434', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 1, remark: '内网部署', created: '2026-02-15T08:00:00Z', updated: '2026-04-01T12:00:00Z',
    },
    {
      aiModelId: 14, providerDisplayName: '本地 Llama', displayName: 'Llama 3.1 8B', providerId: 'ollama', icon: '/logos/ollama.svg', modelName: 'llama3.1:8b',
      apiKey: '', baseUrl: 'http://localhost:11434', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 1, remark: '', created: '2026-03-05T17:00:00Z', updated: '2026-03-05T17:00:00Z',
    },
    {
      aiModelId: 15, providerDisplayName: '本地 CodeLlama', displayName: 'CodeLlama 13B', providerId: 'ollama', icon: '/logos/ollama.svg', modelName: 'codellama:13b',
      apiKey: '', baseUrl: 'http://localhost:11434', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 0, remark: '', created: '2026-03-10T10:00:00Z', updated: '2026-03-10T10:00:00Z',
    },
    {
      aiModelId: 16, providerDisplayName: 'Gemini', displayName: 'Gemini 2.0 Flash', providerId: 'gemini', icon: '', modelName: 'gemini-2.0-flash',
      apiKey: 'sk-g1xxxxxxxxyz23', baseUrl: '', proxyUrl: '', contextLen: 200, extraFields: '',
      isDefault: 0, isFlash: 0, isVisual: 0,
      status: 1, remark: 'Google 多模态模型', created: '2026-04-12T09:00:00Z', updated: '2026-04-12T09:00:00Z',
    },
  ]
}

/* ============================================
   MOCK_PROVIDERS (12 providers)
   ============================================ */

export const MOCK_PROVIDERS: ProviderItem[] = [
  { providerId: 'deepseek', providerDisplayNameZh: 'DeepSeek', providerDisplayNameEn: 'DeepSeek', icon: '/logos/deepseek.svg', defaultBaseUrl: 'https://api.deepseek.com/v1', requireApiKey: true, requireBaseUrl: true },
  { providerId: 'openai', providerDisplayNameZh: 'OpenAI', providerDisplayNameEn: 'OpenAI', icon: '/logos/openai.svg', defaultBaseUrl: '', requireApiKey: true, requireBaseUrl: false },
  { providerId: 'claude', providerDisplayNameZh: 'Claude', providerDisplayNameEn: 'Claude', icon: '/logos/claude.svg', defaultBaseUrl: '', requireApiKey: true, requireBaseUrl: false },
  { providerId: 'gemini', providerDisplayNameZh: 'Gemini', providerDisplayNameEn: 'Gemini', icon: '/logos/gemini.svg', defaultBaseUrl: '', requireApiKey: true, requireBaseUrl: false },
  { providerId: 'bailian', providerDisplayNameZh: 'Bailian', providerDisplayNameEn: 'Bailian', icon: '/logos/bailian.svg', defaultBaseUrl: 'https://dashscope.aliyuncs.com/compatible-mode/v1', requireApiKey: true, requireBaseUrl: true },
  { providerId: 'qwen', providerDisplayNameZh: 'Qwen', providerDisplayNameEn: 'Qwen', icon: '/logos/qwen.svg', defaultBaseUrl: 'https://dashscope.aliyuncs.com/compatible-mode/v1', requireApiKey: true, requireBaseUrl: true },
  { providerId: 'glm', providerDisplayNameZh: 'GLM', providerDisplayNameEn: 'GLM', icon: '/logos/zhipu.svg', defaultBaseUrl: 'https://open.bigmodel.cn/api/paas/v4', requireApiKey: true, requireBaseUrl: true },
  { providerId: 'minimax', providerDisplayNameZh: 'MiniMax', providerDisplayNameEn: 'MiniMax', icon: '/logos/minimax.svg', defaultBaseUrl: 'https://api.minimaxi.com/anthropic', requireApiKey: true, requireBaseUrl: true },
  { providerId: 'volcano', providerDisplayNameZh: 'Volcano', providerDisplayNameEn: 'Volcano', icon: '/logos/huoshan.png', defaultBaseUrl: 'https://ark.cn-beijing.volces.com/api/v3', requireApiKey: true, requireBaseUrl: true },
  { providerId: 'ollama', providerDisplayNameZh: 'Ollama', providerDisplayNameEn: 'Ollama', icon: '/logos/ollama.svg', defaultBaseUrl: '', requireApiKey: false, requireBaseUrl: true },
  { providerId: 'openai_compatible', providerDisplayNameZh: 'OpenAI Compatible', providerDisplayNameEn: 'OpenAI Compatible', icon: '', defaultBaseUrl: '', requireApiKey: true, requireBaseUrl: true },
  { providerId: 'claude_compatible', providerDisplayNameZh: 'Claude Compatible', providerDisplayNameEn: 'Claude Compatible', icon: '', defaultBaseUrl: '', requireApiKey: true, requireBaseUrl: true },
  { providerId: 'openrouter', providerDisplayNameZh: 'OpenRouter', providerDisplayNameEn: 'OpenRouter', icon: '/logos/openrouter.svg', defaultBaseUrl: '', requireApiKey: true, requireBaseUrl: false },
]

/* ============================================
   MOCK_RECOMMENDED_MODELS
   ============================================ */

export const MOCK_RECOMMENDED_MODELS: Record<string, Array<RecommendModelItem & { ownedBy: string }>> = {
  deepseek: [
    { id: 'deepseek-chat', object: 'model', ownedBy: 'deepseek' },
    { id: 'deepseek-coder', object: 'model', ownedBy: 'deepseek' },
    { id: 'deepseek-reasoner', object: 'model', ownedBy: 'deepseek' },
    { id: 'deepseek-v3', object: 'model', ownedBy: 'deepseek' },
    { id: 'deepseek-r1', object: 'model', ownedBy: 'deepseek' },
  ],
  openai: [
    { id: 'gpt-4o', object: 'model', ownedBy: 'openai' },
    { id: 'gpt-4o-mini', object: 'model', ownedBy: 'openai' },
    { id: 'gpt-4-turbo', object: 'model', ownedBy: 'openai' },
    { id: 'gpt-4', object: 'model', ownedBy: 'openai' },
    { id: 'gpt-3.5-turbo', object: 'model', ownedBy: 'openai' },
    { id: 'o3-mini', object: 'model', ownedBy: 'openai' },
    { id: 'o1', object: 'model', ownedBy: 'openai' },
    { id: 'o1-mini', object: 'model', ownedBy: 'openai' },
  ],
  claude: [
    { id: 'claude-sonnet-4-20250514', object: 'model', ownedBy: 'anthropic' },
    { id: 'claude-opus-4-20250514', object: 'model', ownedBy: 'anthropic' },
    { id: 'claude-haiku-3-5-20241022', object: 'model', ownedBy: 'anthropic' },
    { id: 'claude-3-5-sonnet-20241022', object: 'model', ownedBy: 'anthropic' },
    { id: 'claude-3-opus-20240229', object: 'model', ownedBy: 'anthropic' },
  ],
  bailian: [
    { id: 'qwen-turbo-latest', object: 'model', ownedBy: 'bailian' },
    { id: 'qwen-plus-latest', object: 'model', ownedBy: 'bailian' },
    { id: 'qwen-max-latest', object: 'model', ownedBy: 'bailian' },
    { id: 'qwen3-235b-a22b', object: 'model', ownedBy: 'bailian' },
    { id: 'qwen2.5-72b-instruct', object: 'model', ownedBy: 'bailian' },
    { id: 'qwen2.5-32b-instruct', object: 'model', ownedBy: 'bailian' },
    { id: 'qwen2.5-14b-instruct', object: 'model', ownedBy: 'bailian' },
    { id: 'qwen2.5-7b-instruct', object: 'model', ownedBy: 'bailian' },
    { id: 'qwen-vl-max', object: 'model', ownedBy: 'bailian' },
  ],
  ollama: [
    { id: 'qwen2.5:14b', object: 'model', ownedBy: 'library' },
    { id: 'qwen2.5:7b', object: 'model', ownedBy: 'library' },
    { id: 'qwen2.5:32b', object: 'model', ownedBy: 'library' },
    { id: 'llama3.1:8b', object: 'model', ownedBy: 'library' },
    { id: 'llama3.1:70b', object: 'model', ownedBy: 'library' },
    { id: 'codellama:13b', object: 'model', ownedBy: 'library' },
    { id: 'codellama:34b', object: 'model', ownedBy: 'library' },
    { id: 'mistral:7b', object: 'model', ownedBy: 'library' },
    { id: 'gemma2:9b', object: 'model', ownedBy: 'library' },
    { id: 'deepseek-r1:8b', object: 'model', ownedBy: 'library' },
  ],
}

/* ============================================
   AVAILABLE_LOGOS
   ============================================ */

export const AVAILABLE_LOGOS: string[] = [
  '/logos/aicodemirror.svg',
  '/logos/aicoding.svg',
  '/logos/aihubmix-color.svg',
  '/logos/algocode.svg',
  '/logos/alibaba.svg',
  '/logos/anthropic.svg',
  '/logos/apikeyfun.png',
  '/logos/apinebula_icon.png',
  '/logos/atlascloud_icon.png',
  '/logos/aws.svg',
  '/logos/azure.svg',
  '/logos/baidu.svg',
  '/logos/bailian.svg',
  '/logos/bytedance.svg',
  '/logos/byteplus.png',
  '/logos/catcoder.svg',
  '/logos/chatglm.svg',
  '/logos/claude.svg',
  '/logos/ClaudeApi.png',
  '/logos/claudecn.png',
  '/logos/claw.svg',
  '/logos/cloudflare.svg',
  '/logos/cohere.svg',
  '/logos/copilot.svg',
  '/logos/crazyrouter.svg',
  '/logos/ctok.svg',
  '/logos/cubence.svg',
  '/logos/dds.svg',
  '/logos/deepseek.svg',
  '/logos/doubao.svg',
  '/logos/eflowcode.png',
  '/logos/gemini.svg',
  '/logos/gemma.svg',
  '/logos/github.svg',
  '/logos/githubcopilot.svg',
  '/logos/google.svg',
  '/logos/googlecloud.svg',
  '/logos/grok.svg',
  '/logos/hermes.png',
  '/logos/huawei.svg',
  '/logos/huggingface.svg',
  '/logos/hunyuan.svg',
  '/logos/huoshan.png',
  '/logos/kimi.svg',
  '/logos/lemondata.png',
  '/logos/lioncc.svg',
  '/logos/longcat-color.svg',
  '/logos/mcp.svg',
  '/logos/meta.svg',
  '/logos/micu.svg',
  '/logos/midjourney.svg',
  '/logos/minimax.svg',
  '/logos/mistral.svg',
  '/logos/modelscope-color.svg',
  '/logos/newapi.svg',
  '/logos/notion.svg',
  '/logos/novita.svg',
  '/logos/nvidia.svg',
  '/logos/ollama.svg',
  '/logos/openai.svg',
  '/logos/opencode-logo-light.svg',
  '/logos/openrouter.svg',
  '/logos/packycode.svg',
  '/logos/palm.svg',
  '/logos/pateway.jpg',
  '/logos/perplexity.svg',
  '/logos/pipellm.png',
  '/logos/qwen.svg',
  '/logos/rc.svg',
  '/logos/relaxcode.png',
  '/logos/runapi.jpg',
  '/logos/shengsuanyun.svg',
  '/logos/siliconflow.svg',
  '/logos/sssaicode.svg',
  '/logos/stability.svg',
  '/logos/stepfun.svg',
  '/logos/sudocode.png',
  '/logos/tencent.svg',
  '/logos/ucloud.svg',
  '/logos/vercel.svg',
  '/logos/wenxin.svg',
  '/logos/xai.svg',
  '/logos/xiaomimimo.svg',
  '/logos/yi.svg',
  '/logos/zeroone.svg',
  '/logos/zhipu.svg',
]

/* ============================================
   Mutable state — shared across all handlers
   ============================================ */

let models: StoredModel[] = generateMockModels()
let nextAiModelId: number = models.length > 0 ? Math.max(...models.map((m) => m.aiModelId)) + 1 : 1

/* ============================================
   Request types
   ============================================ */

export interface GetModelsParams {
  providerId?: string
  search?: string
  page: number
  pageSize: number
}

export interface CreateModelRequest {
  providerDisplayName: string
  displayName: string
  providerId: string
  modelName: string
  icon: string
  apiKey: string
  baseUrl: string
  proxyUrl: string
  contextLen: number
  extraFields: string
  isDefault: number
  isFlash: number
  isVisual: number
  remark: string
}

export interface UpdateModelRequest {
  providerDisplayName: string
  displayName: string
  modelName: string
  icon: string
  apiKey: string
  baseUrl: string
  proxyUrl: string
  contextLen: number
  extraFields: string
  isDefault: number
  isFlash: number
  isVisual: number
  status: number
  remark: string
}

export interface TestConnectionRequest {
  providerId?: string
  apiKey?: string
  baseUrl?: string
  modelName?: string
}

export interface TestConnectionResult {
  ok: boolean
  latency: number
}

/* ============================================
   Helpers
   ============================================ */

function paginate<T>(list: T[], page: number, pageSize: number): PaginatedResponse<T> {
  const totalCount = list.length
  const totalPage = Math.max(1, Math.ceil(totalCount / pageSize))
  const safePage = Math.min(page, totalPage)
  const start = (safePage - 1) * pageSize
  const paged = list.slice(start, start + pageSize)

  return {
    list: paged,
    totalPage,
    totalCount,
    page: safePage,
    pageSize,
  }
}

function applyExclusiveFlags(ms: StoredModel[], data: { isDefault?: number; isFlash?: number; isVisual?: number }): StoredModel[] {
  return ms.map((m) => ({
    ...m,
    ...(data.isDefault === 1 && m.isDefault === 1 ? { isDefault: 0 } : {}),
    ...(data.isFlash === 1 && m.isFlash === 1 ? { isFlash: 0 } : {}),
    ...(data.isVisual === 1 && m.isVisual === 1 ? { isVisual: 0 } : {}),
  }))
}

/* ============================================
   Handlers
   ============================================ */

/** Reset to initial mock data */
export function resetModels(): void {
  models = generateMockModels()
  nextAiModelId = models.length > 0 ? Math.max(...models.map((m) => m.aiModelId)) + 1 : 1
}

/** List models with optional search, provider filter, and pagination */
export async function getModels(params: GetModelsParams): Promise<PaginatedResponse<AdminModelItem>> {
  await listDelay()

  let filtered = models

  if (params.providerId && params.providerId !== 'all') {
    filtered = filtered.filter((m) => m.providerId === params.providerId)
  }

  if (params.search?.trim()) {
    const q = params.search.trim().toLowerCase()
    filtered = filtered.filter(
      (m) =>
        m.modelName.toLowerCase().includes(q) ||
        m.providerDisplayName.toLowerCase().includes(q),
    )
  }

  const paged = paginate(filtered, params.page, params.pageSize)
  return {
    ...paged,
    list: paged.list.map(storedToItem),
  }
}

/** Get single model detail (with full apiKey) */
export async function getModelDetail(id: number): Promise<AdminModelItem> {
  await itemDelay()

  const model = models.find((m) => m.aiModelId === id)
  if (!model) {
    throw new Error(`Model not found: ${id}`)
  }

  return storedToDetail(model)
}

/** Create a new model */
export async function createModel(req: CreateModelRequest): Promise<AdminModelItem> {
  await mutationDelay()

  const now = new Date().toISOString()
  const newModel: StoredModel = {
    aiModelId: nextAiModelId++,
    providerDisplayName: req.providerDisplayName,
    displayName: req.displayName || req.modelName,
    providerId: req.providerId,
    modelName: req.modelName,
    icon: req.icon,
    apiKey: req.apiKey,
    baseUrl: req.baseUrl,
    proxyUrl: req.proxyUrl,
    contextLen: req.contextLen,
    extraFields: req.extraFields,
    isDefault: req.isDefault,
    isFlash: req.isFlash,
    isVisual: req.isVisual,
    status: 1,
    remark: req.remark,
    created: now,
    updated: now,
  }

  models = [newModel, ...applyExclusiveFlags(models, req)]
  return storedToDetail(newModel)
}

/** Update an existing model */
export async function updateModel(id: number, req: UpdateModelRequest): Promise<AdminModelItem> {
  await mutationDelay()

  const idx = models.findIndex((m) => m.aiModelId === id)
  if (idx === -1) {
    throw new Error(`Model not found: ${id}`)
  }

  const now = new Date().toISOString()

  models = models.map((m) => {
    if (m.aiModelId !== id) {
      let updated = { ...m }
      if (req.isDefault === 1 && m.isDefault === 1) updated = { ...updated, isDefault: 0 }
      if (req.isFlash === 1 && m.isFlash === 1) updated = { ...updated, isFlash: 0 }
      if (req.isVisual === 1 && m.isVisual === 1) updated = { ...updated, isVisual: 0 }
      return updated
    }
    return {
      ...m,
      providerDisplayName: req.providerDisplayName,
      displayName: req.displayName || req.modelName,
      modelName: req.modelName,
      icon: req.icon,
      apiKey: req.apiKey || m.apiKey,
      baseUrl: req.baseUrl,
      proxyUrl: req.proxyUrl,
      contextLen: req.contextLen,
      extraFields: req.extraFields,
      isDefault: req.isDefault,
      isFlash: req.isFlash,
      isVisual: req.isVisual,
      status: req.status,
      remark: req.remark,
      updated: now,
    }
  })

  const updated = models[idx]
  return storedToDetail(updated)
}

/** Toggle model status */
export async function toggleModelStatus(id: number, status: number): Promise<void> {
  await mutationDelay()

  models = models.map((m) =>
    m.aiModelId === id ? { ...m, status } : m,
  )
}

/** Set a model as the default */
export async function setDefaultModel(id: number): Promise<void> {
  await mutationDelay()

  models = models.map((m) => ({
    ...m,
    isDefault: m.aiModelId === id ? 1 : 0,
  }))
}

/** Set a model as the flash model */
export async function setFlashModel(id: number): Promise<void> {
  await mutationDelay()

  models = models.map((m) => ({
    ...m,
    isFlash: m.aiModelId === id ? 1 : 0,
  }))
}

/** Set a model as the visual model */
export async function setVisualModel(id: number): Promise<void> {
  await mutationDelay()

  models = models.map((m) => ({
    ...m,
    isVisual: m.aiModelId === id ? 1 : 0,
  }))
}

/** Delete a model */
export async function deleteModel(id: number): Promise<void> {
  await mutationDelay()

  models = models.filter((m) => m.aiModelId !== id)
}

/** Get all providers */
export async function getProviders(): Promise<ProviderItem[]> {
  await listDelay()
  return MOCK_PROVIDERS
}

/** Get recommended models for a provider */
export async function getRecommendedModels(
  providerId: string,
  _apiKey?: string,
  _baseUrl?: string,
): Promise<Array<RecommendModelItem & { ownedBy: string }>> {
  await itemDelay()

  return MOCK_RECOMMENDED_MODELS[providerId] || []
}

/** Simulate model connectivity test */
export async function testModelConnection(_req: TestConnectionRequest): Promise<TestConnectionResult> {
  await heavyDelay()

  const pass = Math.random() > 0.3
  return {
    ok: pass,
    latency: pass ? Math.floor(40 + Math.random() * 160) : 0,
  }
}
