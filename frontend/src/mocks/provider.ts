/* ============================================
   Mock data — Provider / Available Models
   GET /api/providers/models
   ============================================ */

import type { AvailableModel } from './types'
import { listDelay } from './delay'

/* ---------- Mock data ---------- */

export const MOCK_MODELS: AvailableModel[] = [
  { aiModelId: 1, providerDisplayName: 'DeepSeek', displayName: 'DeepSeek V4 Pro', modelName: 'DeepSeek V4 Pro', icon: 'brain', contextLen: 131072, isDefault: true },
  { aiModelId: 2, providerDisplayName: 'Alibaba', displayName: 'Qwen Max', modelName: 'Qwen Max', icon: 'cpu', contextLen: 32768, isDefault: false },
  { aiModelId: 3, providerDisplayName: 'Zhipu', displayName: 'GLM-4', modelName: 'GLM-4', icon: 'cpu', contextLen: 131072, isDefault: false },
  { aiModelId: 4, providerDisplayName: 'OpenAI', displayName: 'GPT-4o', modelName: 'GPT-4o', icon: 'cpu', contextLen: 131072, isDefault: false },
  { aiModelId: 5, providerDisplayName: 'Anthropic', displayName: 'Claude 3.5 Sonnet', modelName: 'Claude 3.5 Sonnet', icon: 'cpu', contextLen: 200000, isDefault: false },
  { aiModelId: 6, providerDisplayName: 'Google', displayName: 'Gemini 2.0 Flash', modelName: 'Gemini 2.0 Flash', icon: 'cpu', contextLen: 1048576, isDefault: false },
]

/* ---------- Async function ---------- */

export async function getAvailableModels(): Promise<AvailableModel[]> {
  await listDelay()
  return [...MOCK_MODELS]
}
