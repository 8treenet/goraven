import http from '../http'
import type { ProviderItem } from '../types'

export interface RecommendModelItem {
  id: string
  object: string
  ownedBy: string
}

/** GET /api/admin/providers */
export function getProviders() {
  return http.get<{ list: ProviderItem[] }>('/admin/providers')
}

/** POST /api/admin/providers/recommend（密钥走 JSON body，不进 URL / 访问日志） */
export function getRecommendModels(params: {
  providerId: string
  apiKey?: string
  baseUrl?: string
}) {
  return http.post<{ list: RecommendModelItem[] }>('/admin/providers/recommend', params)
}
