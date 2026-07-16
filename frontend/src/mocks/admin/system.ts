/* ============================================
   Admin System — mock data
   Settings, System Info, Dashboard
   ============================================ */

import type {
  SettingGroupData,
  AdminDashboardData,
  TokenTrendItem,
  ActiveTrendItem,
  SparklineItem,
} from '../types'
import { listDelay, mutationDelay, heavyDelay } from '../delay'

/* ========================================================================
   Settings (6 groups, 21 settings)
   Aligned with AdminSettingsPage.tsx
   ======================================================================== */

export const MOCK_SETTINGS: SettingGroupData[] = [
  {
    name: 'general',
    displayName: '基本配置',
    displayOrder: 1,
    settings: [
      {
        key: 'general.domain',
        value: 'https://goraven.dev',
        valueType: 'string',
        defaultValue: '',
        displayName: '系统域名',
        description: '系统对外服务域名，用于生成文件外链和分享链接',
        inputType: 'text',
        displayOrder: 1,
      },
    ],
  },
  {
    name: 'tools',
    displayName: '工具',
    displayOrder: 2,
    settings: [
      {
        key: 'tools.webfetch_enabled',
        value: 'true',
        valueType: 'bool',
        defaultValue: 'true',
        displayName: '网页读取',
        description:
          '启用后 Agent 可通过 HTTP GET 读取指定网页内容。注意：仅能获取服务端返回的原始 HTML，无法读取前端渲染的动态页面。',
        inputType: 'switch',
        displayOrder: 1,
      },
      {
        key: 'tools.visual_enabled',
        value: 'false',
        valueType: 'bool',
        defaultValue: 'false',
        displayName: '多模态识别',
        description:
          '启用图像、视频、音频识别能力。需在模型管理中设置多模态模型，且该模型需支持多模态。',
        inputType: 'switch',
        displayOrder: 2,
      },
    ],
  },
  {
    name: 'clawhub',
    displayName: 'ClawHub',
    displayOrder: 2,
    settings: [
      {
        key: 'clawhub.api_url',
        value: 'https://clawhub.ai',
        valueType: 'string',
        defaultValue: 'https://clawhub.ai',
        displayName: 'ClawHub API 地址',
        description: 'ClawHub 服务接口地址',
        inputType: 'text',
        displayOrder: 1,
      },
      {
        key: 'clawhub.token',
        value: 'sk-xxxx-token-xxxx',
        valueType: 'string',
        defaultValue: '',
        displayName: '加速 Token',
        description: 'ClawHub 加速令牌，留空则不走加速',
        inputType: 'text',
        displayOrder: 2,
      },
    ],
  },
  {
    name: 'agent',
    displayName: 'Agent 配置',
    displayOrder: 3,
    settings: [
      {
        key: 'agent.compress_threshold_percent',
        value: '80',
        valueType: 'int',
        defaultValue: '80',
        displayName: '压缩阈值百分比',
        description: '上下文占模型窗口百分比超过此值时触发压缩',
        inputType: 'slider',
        min: 40,
        max: 80,
        displayOrder: 1,
      },
      {
        key: 'agent.compress_keep_rounds',
        value: '4',
        valueType: 'int',
        defaultValue: '4',
        displayName: '压缩保留轮数',
        description: '压缩时保留最近几轮对话不压缩',
        inputType: 'number',
        min: 1,
        max: 20,
        displayOrder: 2,
      },
      {
        key: 'agent.max_iterations',
        value: '120',
        valueType: 'int',
        defaultValue: '120',
        displayName: '最大迭代步数',
        description: 'Agent 单次对话最大执行步骤',
        inputType: 'number',
        min: 1,
        max: 500,
        displayOrder: 3,
      },
      {
        key: 'agent.pruning_token_threshold',
        value: '96',
        valueType: 'int',
        defaultValue: '96',
        displayName: '剪枝 Token 阈值/K',
        description: '总 token 超过此阈值（单位 K）时触发剪枝',
        inputType: 'number',
        min: 1,
        max: 1000,
        displayOrder: 4,
      },
      {
        key: 'agent.pruning_keep_recent_rounds',
        value: '12',
        valueType: 'int',
        defaultValue: '12',
        displayName: '剪枝保留最近轮数',
        description: '剪枝时保留最近几轮工具调用',
        inputType: 'number',
        min: 1,
        max: 50,
        displayOrder: 5,
      },
      {
        key: 'agent.pruning_keep_intact_rounds',
        value: '6',
        valueType: 'int',
        defaultValue: '6',
        displayName: '剪枝完整保留轮数',
        description: '最近几轮不做截断完整保留',
        inputType: 'number',
        min: 1,
        max: 20,
        displayOrder: 6,
      },
      {
        key: 'agent.pruning_recent_time_window_sec',
        value: '600',
        valueType: 'int',
        defaultValue: '600',
        displayName: '剪枝时间窗口/秒',
        description: '最近时间窗口内的消息优先保留',
        inputType: 'number',
        min: 60,
        max: 3600,
        displayOrder: 7,
      },
      {
        key: 'agent.pruning_max_tool_result_length',
        value: '2000',
        valueType: 'int',
        defaultValue: '2000',
        displayName: '工具结果最大长度',
        description: '工具返回结果超过此长度时截断',
        inputType: 'number',
        min: 100,
        max: 50000,
        displayOrder: 8,
      },
      {
        key: 'agent.pruning_head_truncate_length',
        value: '1000',
        valueType: 'int',
        defaultValue: '1000',
        displayName: '截断保留头部',
        description: '截断时保留头部的字符数',
        inputType: 'number',
        min: 0,
        max: 10000,
        displayOrder: 9,
      },
      {
        key: 'agent.pruning_tail_truncate_length',
        value: '1000',
        valueType: 'int',
        defaultValue: '1000',
        displayName: '截断保留尾部',
        description: '截断时保留尾部的字符数',
        inputType: 'number',
        min: 0,
        max: 10000,
        displayOrder: 10,
      },
      {
        key: 'agent.llm_request_delay_ms',
        value: '500',
        valueType: 'int',
        defaultValue: '500',
        displayName: 'LLM 请求延迟/ms',
        description: '两次 LLM 请求之间的延迟，用于避免限流',
        inputType: 'number',
        min: 0,
        max: 10000,
        displayOrder: 11,
      },
      {
        key: 'agent.max_retries',
        value: '3',
        valueType: 'int',
        defaultValue: '3',
        displayName: 'LLM 最大重试次数',
        description: 'LLM 调用失败时的最大重试次数',
        inputType: 'number',
        min: 0,
        max: 10,
        displayOrder: 12,
      },
      {
        key: 'agent.rate_limit_wait_sec',
        value: '8',
        valueType: 'int',
        defaultValue: '8',
        displayName: '429 限流等待/秒',
        description: '遇到 429 限流时的固定等待秒数',
        inputType: 'number',
        min: 1,
        max: 60,
        displayOrder: 13,
      },
      {
        key: 'agent.backoff_base_sec',
        value: '3',
        valueType: 'int',
        defaultValue: '3',
        displayName: '重试退避',
        description: '第 N 次重试等待 N×此值 秒',
        inputType: 'number',
        min: 1,
        max: 30,
        displayOrder: 14,
      },
    ],
  },
  {
    name: 'sharing',
    displayName: '分享配置',
    displayOrder: 4,
    settings: [
      {
        key: 'sharing.file_expires_hours',
        value: '72',
        valueType: 'int',
        defaultValue: '72',
        displayName: '文件外链有效期/小时',
        description: '文件分享外链的有效期',
        inputType: 'number',
        min: 1,
        max: 720,
        displayOrder: 1,
      },
    ],
  },
  {
    name: 'knowledge',
    displayName: '知识库',
    displayOrder: 5,
    settings: [
      {
        key: 'knowledge.enable_ocr',
        value: 'false',
        valueType: 'bool',
        defaultValue: 'false',
        displayName: 'OCR 开关',
        description: '是否启用 OCR 解析',
        inputType: 'switch',
        displayOrder: 1,
      },
    ],
  },
]

/** Live values store — seeded from MOCK_SETTINGS on first access */
const settingsValues: Record<string, string> = {}

function ensureSettingsInit(): void {
  if (Object.keys(settingsValues).length > 0) return
  for (const g of MOCK_SETTINGS) {
    for (const s of g.settings) {
      settingsValues[s.key] = s.value as string
    }
  }
}

export async function getSettings(): Promise<SettingGroupData[]> {
  await listDelay()
  ensureSettingsInit()

  // Return settings with live values merged in
  return MOCK_SETTINGS.map((g) => ({
    ...g,
    settings: g.settings.map((s) => ({
      ...s,
      value: settingsValues[s.key] ?? s.value,
    })),
  }))
}

export async function updateSettings(
  updates: { key: string; value: string }[],
): Promise<void> {
  await mutationDelay()
  ensureSettingsInit()

  for (const { key, value } of updates) {
    settingsValues[key] = value
  }
}

/* ========================================================================
   System Info
   Aligned with AdminSystemInfoPage.tsx
   ======================================================================== */

export const MOCK_SYSTEM_INFO = {
  overview: {
    version: '1.0.0',
    language: 'zh',
    cacheType: 'local',
    cacheMemory: '1500 items',
    timezone: 'Asia/Shanghai',
    uploadBytes: 20971520,
    tempBytes: 0,
  },
  database: {
    type: 'sqlite',
    version: '3.45.0',
    name: '/data/raven.db',
    dataSizeBytes: 10485760,
    pool: {
      maxOpenConnections: 25,
      openConnections: 2,
      inUse: 1,
      idle: 1,
      waitCount: 0,
      waitDurationMs: 0,
      maxIdleClosed: 0,
      maxLifetimeClosed: 0,
    },
  },
  disks: [
    { mountPoint: '/', fsType: 'apfs', device: '/dev/disk3s1', totalBytes: 500000000000, usedBytes: 300000000000, freeBytes: 200000000000, usedPercent: 60.0 },
    { mountPoint: '/home', fsType: 'ext4', device: '/dev/sda1', totalBytes: 1000000000000, usedBytes: 850000000000, freeBytes: 150000000000, usedPercent: 85.0 },
    { mountPoint: '/data', fsType: 'xfs', device: '/dev/sdb1', totalBytes: 2000000000000, usedBytes: 1860000000000, freeBytes: 140000000000, usedPercent: 93.0 },
    { mountPoint: '/backup', fsType: 'ext4', device: '/dev/sdc1', totalBytes: 4000000000000, usedBytes: 1200000000000, freeBytes: 2800000000000, usedPercent: 30.0 },
  ],
  mcpHealth: [
    { mcpId: 1, name: 'filesystem', displayName: '文件系统', icon: 'folder', healthLatency: 150, healthCheckedAt: '2025-06-15T10:30:00Z' },
    { mcpId: 2, name: 'brave-search', displayName: 'Brave 搜索', icon: 'search', healthLatency: 1200, healthCheckedAt: '2025-06-15T10:30:00Z' },
    { mcpId: 3, name: 'github', displayName: 'GitHub', icon: 'github', healthLatency: 0, healthCheckedAt: '2025-06-15T10:25:00Z' },
    { mcpId: 4, name: 'postgres', displayName: 'PostgreSQL', icon: 'database', healthLatency: 85, healthCheckedAt: '2025-06-15T10:30:00Z' },
    { mcpId: 5, name: 'redis', displayName: 'Redis 缓存', icon: 'server', healthLatency: 42, healthCheckedAt: '2025-06-15T10:30:00Z' },
    { mcpId: 6, name: 'slack', displayName: 'Slack 通知', icon: 'message-square', healthLatency: 2300, healthCheckedAt: '2025-06-15T10:29:00Z' },
    { mcpId: 7, name: 'notion', displayName: 'Notion 文档', icon: 'file-text', healthLatency: 0, healthCheckedAt: '2025-06-15T10:20:00Z' },
    { mcpId: 8, name: 'jira', displayName: 'Jira 工单', icon: 'ticket', healthLatency: 320, healthCheckedAt: '2025-06-15T10:30:00Z' },
    { mcpId: 9, name: 'docker', displayName: 'Docker 引擎', icon: 'container', healthLatency: 67, healthCheckedAt: '2025-06-15T10:30:00Z' },
    { mcpId: 10, name: 'linear', displayName: 'Linear', icon: 'list', healthLatency: 910, healthCheckedAt: '2025-06-15T10:30:00Z' },
    { mcpId: 11, name: 'chromadb', displayName: 'ChromaDB', icon: 'layers', healthLatency: 1800, healthCheckedAt: '2025-06-15T10:28:00Z' },
    { mcpId: 12, name: 'prometheus', displayName: 'Prometheus', icon: 'bar-chart', healthLatency: 0, healthCheckedAt: '2025-06-15T09:00:00Z' },
  ],
  ecosystem: {
    totalUsers: 10, activeUsers: 8, adminUsers: 1,
    totalModels: 5, enabledModels: 3,
    totalMcps: 3, enabledMcps: 2,
    systemSkills: 5, marketSkills: 10,
    personaTemplates: 6,
    totalSessions: 100, totalMessages: 5000,
    totalSharedProjects: 3,
    totalShareLinks: 5, activeShareLinks: 3, totalShareViews: 42,
  },
  plugins: [
    { name: 'audit', version: '1.0.0' },
    { name: 'search', version: '0.2.3' },
    { name: 'sanitize', version: '2.1.0' },
  ],
  collectedAt: '2025-06-15T10:30:00Z',
}

export async function getSystemInfo() {
  await listDelay()
  return { ...MOCK_SYSTEM_INFO }
}

/* ========================================================================
   Admin Dashboard
   Aligned with AdminDashboardPage.tsx
   ======================================================================== */

/* ---- Trend generators (return fresh random data each call) ---- */

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

export function generateActiveTrend(days: number): ActiveTrendItem[] {
  const now = new Date()
  return Array.from({ length: days }, (_, i) => {
    const d = new Date(now)
    d.setDate(d.getDate() - (days - 1 - i))
    const dateStr = `${d.getMonth() + 1}/${d.getDate()}`
    return {
      date: dateStr,
      count: Math.round(30 + Math.sin(i * 0.25) * 15 + Math.random() * 10),
    }
  })
}

export function generateSparkline(): SparklineItem[] {
  return Array.from({ length: 7 }, (_, i) => ({
    date: `D-${7 - i}`,
    tokens: Math.round(8000 + Math.sin(i * 0.8) * 4000 + Math.random() * 3000),
  }))
}

/* ---- Pre-computed snapshots ---- */

export const TREND_7 = generateTrendData(7)
export const TREND_30 = generateTrendData(30)
export const TREND_90 = generateTrendData(90)

export const ACTIVE_7 = generateActiveTrend(7)
export const ACTIVE_30 = generateActiveTrend(30)
export const ACTIVE_90 = generateActiveTrend(90)

/* ---- Dashboard data ---- */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const userTokenRankData: any[] = [
  { userId: '1', username: 'zhangsan', tokenCount: 520000, percentage: 30.5 },
  { userId: '2', username: 'lisi', tokenCount: 310000, percentage: 18.2 },
  { userId: '3', username: 'wangwu', tokenCount: 180000, percentage: 10.6 },
  { userId: '4', username: 'zhaoliu', tokenCount: 145000, percentage: 8.5 },
  { userId: '5', username: 'sunqi', tokenCount: 98000, percentage: 5.7 },
  { userId: '6', username: 'zhouba', tokenCount: 87000, percentage: 5.1 },
  { userId: '7', username: 'wujiu', tokenCount: 65000, percentage: 3.8 },
  { userId: '8', username: 'zhengshi', tokenCount: 52000, percentage: 3.1 },
  { userId: '9', username: 'zhengshi2', tokenCount: 2000, percentage: 3.1 }
]

export const MOCK_ADMIN_DASHBOARD: AdminDashboardData = {
  overview: {
    activeUsers: 42,
    activeUsersDiff: 0.15,
    totalSessions: 500,
    newSessions: 30,
    weekTokens: 500000,
    todayTokens: 80000,
    enabledModels: 3,
    sparkline: generateSparkline(),
  },
  tokenTrend: TREND_30,
  modelUsage: [
    { modelName: 'DeepSeek V3', tokenCount: 245600, percentage: 45.2 },
    { modelName: 'Claude 3.5', tokenCount: 153400, percentage: 28.2 },
    { modelName: 'Qwen Max', tokenCount: 68700, percentage: 12.6 },
    { modelName: 'GLM-4 Plus', tokenCount: 45200, percentage: 8.3 },
    { modelName: 'GPT-4o', tokenCount: 31200, percentage: 5.7 },
  ],
  userTokenRank: userTokenRankData,
  activeTrend: ACTIVE_30,
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
}

export async function getAdminDashboard(refresh?: boolean): Promise<AdminDashboardData> {
  await (refresh ? heavyDelay() : listDelay())
  return {
    ...MOCK_ADMIN_DASHBOARD,
    overview: {
      ...MOCK_ADMIN_DASHBOARD.overview,
      sparkline: generateSparkline(),
    },
  }
}

export async function getAdminTokenTrend(
  days?: number,
  refresh?: boolean,
): Promise<{ items: TokenTrendItem[] }> {
  await (refresh ? heavyDelay() : listDelay())

  let trendMap: Record<number, TokenTrendItem[]>
  trendMap = { 7: TREND_7, 30: TREND_30, 90: TREND_90 }

  const items = days != null && trendMap[days]
    ? trendMap[days]
    : generateTrendData(days ?? 30)

  return { items }
}

export async function getAdminActiveUsers(
  days?: number,
  refresh?: boolean,
): Promise<{ items: ActiveTrendItem[] }> {
  await (refresh ? heavyDelay() : listDelay())

  let activeMap: Record<number, ActiveTrendItem[]>
  activeMap = { 7: ACTIVE_7, 30: ACTIVE_30, 90: ACTIVE_90 }

  const items = days != null && activeMap[days]
    ? activeMap[days]
    : generateActiveTrend(days ?? 30)

  return { items }
}
