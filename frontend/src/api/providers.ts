import http from './http'
import type { ModelInfo } from './types'

/** GET /api/providers/models */
export function getAvailableModels() {
  return http.get<ModelInfo[]>('/providers/models')
}

/** GET /api/providers/models/:id */
export function getModel(id: number) {
  return http.get<ModelInfo>(`/providers/models/${id}`)
}
