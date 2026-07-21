import http from './http'
import type { DashboardData, TokenTrendRsp, ModelUsageRsp } from './types'

/** GET /api/dashboard */
export function getDashboard() {
  return http.get<DashboardData>('/dashboard')
}

/** GET /api/dashboard/tokenTrend */
export function getTokenTrend(params?: { days?: number }) {
  return http.get<TokenTrendRsp>('/dashboard/tokenTrend', { params })
}

/** GET /api/dashboard/modelUsage */
export function getModelUsage(params?: { days?: number }) {
  return http.get<ModelUsageRsp>('/dashboard/modelUsage', { params })
}
