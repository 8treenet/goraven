import http from './http'
import type { DashboardData, TokenTrendRsp } from './types'

/** GET /api/dashboard */
export function getDashboard(params?: { refresh?: boolean }) {
  return http.get<DashboardData>('/dashboard', { params })
}

/** GET /api/dashboard/tokenTrend */
export function getTokenTrend(params?: { days?: number; refresh?: boolean }) {
  return http.get<TokenTrendRsp>('/dashboard/tokenTrend', { params })
}
