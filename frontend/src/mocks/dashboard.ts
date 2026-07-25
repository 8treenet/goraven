/* ============================================
   Mock data — Dashboard
   Extracted from features/dashboard/DashboardPage.tsx
   ============================================ */

import type { DashboardData, TokenTrendItem } from './types'
import { listDelay } from './delay'

/* ---------- Trend data generator ---------- */

export function generateTrendData(days: number): TokenTrendItem[] {
  const now = new Date()
  return Array.from({ length: days }, (_, i) => {
    const d = new Date(now)
    d.setDate(d.getDate() - (days - 1 - i))
    const dateStr = `${d.getMonth() + 1}/${d.getDate()}`
    const base = 3000 + Math.sin(i * 0.3) * 1000 + Math.random() * 2000
    return {
      date: dateStr,
      promptTokens: Math.round(base * (0.5 + Math.random() * 0.3)),
      completionTokens: Math.round(base * (0.2 + Math.random() * 0.3)),
    }
  })
}

/* ---------- Pre-generated trend arrays ---------- */

export const TREND_7 = generateTrendData(7)
export const TREND_30 = generateTrendData(30)
export const TREND_90 = generateTrendData(90)

/* ---------- Mock dashboard data ---------- */

export const MOCK_DASHBOARD: DashboardData = {
  overview: {
    todayTokens: 12453,
    weekTokens: 156234,
    totalTokens: 2345678,
    dailyTokenLimit: 10,
    totalSessions: 42,
    newSessions: 3,
    sparkline: Array.from({ length: 7 }, (_, i) => ({
      date: `D-${7 - i}`,
      tokens: Math.round(4000 + Math.sin(i * 0.8) * 2000 + Math.random() * 1500),
    })),
  },
  toolUsageRank: [
    { name: 'filesystem', count: 89 },
    { name: 'execute', count: 45 },
    { name: 'file_read', count: 32 },
    { name: 'web_fetch', count: 28 },
    { name: 'file_write', count: 21 },
    { name: 'file_delete', count: 15 },
    { name: 'search', count: 12 },
    { name: 'dir_list', count: 10 },
    { name: 'git_diff', count: 7 },
    { name: 'git_log', count: 5 },
  ],
  mcpUsageRank: [
    { name: 'brave-search', count: 234 },
    { name: 'github', count: 187 },
    { name: 'postgres', count: 156 },
    { name: 'slack', count: 98 },
    { name: 'redis', count: 67 },
    { name: 'docker', count: 45 },
    { name: 'filesystem', count: 34 },
    { name: 'jira', count: 23 },
    { name: 'notion', count: 18 },
    { name: 'linear', count: 12 },
  ],
  skillUsageRank: [
    { name: 'code-review', count: 89 },
    { name: 'pdf-reader', count: 76 },
    { name: 'web-search', count: 65 },
    { name: 'api-doc-gen', count: 54 },
    { name: 'excel-analyzer', count: 43 },
    { name: 'db-migration', count: 32 },
    { name: 'unit-test-gen', count: 28 },
    { name: 'docker-deploy', count: 19 },
    { name: 'log-analyzer', count: 14 },
    { name: 'k8s-ops', count: 8 },
  ],
  storageStats: {
    usedBytes: 1_342_177_280,  // 1.25 GB
    freeBytes: 9_395_264_512,  // 8.75 GB
    totalBytes: 10_737_418_240, // 10 GB
    items: [
      { name: 'documents', bytesSize: 483_183_820, percentage: 0.045 },
      { name: 'images', bytesSize: 268_435_456, percentage: 0.025 },
      { name: 'projects', bytesSize: 429_496_729, percentage: 0.04 },
      { name: 'other', bytesSize: 161_061_273, percentage: 0.015 },
    ],
  },
}

/* ---------- Async mock functions ---------- */

export async function getDashboard(): Promise<DashboardData> {
  await listDelay()
  return MOCK_DASHBOARD
}

export async function getTokenTrend(
  days?: number,
): Promise<{ items: TokenTrendItem[] }> {
  await listDelay()
  const data =
    days === 7
      ? TREND_7
      : days === 90
        ? TREND_90
        : TREND_30
  return { items: data }
}
