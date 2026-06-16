import http from '../http'
import type { SettingGroup, AdminDashboardData, TokenTrendItem, ActiveTrendItem } from '../types'

/* ---------- System info ---------- */

/** GET /api/admin/systemInfo */
export function getSystemInfo(params?: { forceRefresh?: boolean }) {
  return http.get<Record<string, unknown>>('/admin/systemInfo', { params })
}

/* ---------- Dashboard ---------- */

/** GET /api/admin/dashboard */
export function getDashboard(params?: { refresh?: boolean }) {
  return http.get<AdminDashboardData>('/admin/dashboard', { params })
}

/** GET /api/admin/dashboard/tokenTrend */
export function getTokenTrend(params?: { days?: number; refresh?: boolean }) {
  return http.get<{ items: TokenTrendItem[] }>('/admin/dashboard/tokenTrend', { params })
}

/** GET /api/admin/dashboard/activeUsers */
export function getActiveUsers(params?: { days?: number; refresh?: boolean }) {
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
