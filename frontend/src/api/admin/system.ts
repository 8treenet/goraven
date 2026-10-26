import http from '../http'
import type { SettingGroup, AdminDashboardData, TokenTrendItem, ActiveTrendItem, ModelUsageItem, UserTokenRankItem } from '../types'

/* ---------- System info ---------- */

/** GET /api/admin/systemInfo */
export function getSystemInfo(params?: { forceRefresh?: boolean }) {
  return http.get<Record<string, unknown>>('/admin/systemInfo', { params })
}

/* ---------- Dashboard ---------- */

/** GET /api/admin/dashboard */
export function getDashboard() {
  return http.get<AdminDashboardData>('/admin/dashboard')
}

/** GET /api/admin/dashboard/tokenTrend */
export function getTokenTrend(params?: { days?: number }) {
  return http.get<{ items: TokenTrendItem[] }>('/admin/dashboard/tokenTrend', { params })
}

/** GET /api/admin/dashboard/modelUsage */
export function getModelUsage(params?: { days?: number }) {
  return http.get<{ items: ModelUsageItem[] }>('/admin/dashboard/modelUsage', { params })
}

/** GET /api/admin/dashboard/userTokenRank */
export function getUserTokenRank(params?: { days?: number }) {
  return http.get<{ items: UserTokenRankItem[] }>('/admin/dashboard/userTokenRank', { params })
}

/** GET /api/admin/dashboard/activeUsers */
export function getActiveUsers(params?: { days?: number }) {
  return http.get<{ items: ActiveTrendItem[] }>('/admin/dashboard/activeUsers', { params })
}

/* ---------- Settings ---------- */

/** GET /api/admin/settings */
export function getSettings() {
  return http.get<{ groups: SettingGroup[] }>('/admin/settings')
}

/** PUT /api/admin/settings */
export function updateSettings(updates: { key: string; value: string }[]) {
  return http.put('/admin/settings', { settings: updates })
}
