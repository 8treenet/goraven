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

/** GET /api/admin/providers/recommend */
export function getRecommendModels(params: {
  providerId: string
  apiKey?: string
  baseUrl?: string
}) {
  return http.get<{ list: RecommendModelItem[] }>('/admin/providers/recommend', { params })
}
