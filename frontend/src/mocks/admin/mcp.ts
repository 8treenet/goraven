/* ============================================
   Mock data — Admin MCP
   Extracted from features/admin/mcp/AdminMcpPage.tsx
   ============================================ */

import type { AdminMcpItem, McpRecommendItem, PaginatedResponse } from '../types'
import { listDelay, itemDelay, mutationDelay, heavyDelay } from '../delay'

export const PAGE_SIZE = 10

/* ============================================
   In‑memory store
   ============================================ */

let mcpStore: AdminMcpItem[] = []
let storeInitialized = false

function ensureStore(): AdminMcpItem[] {
  if (!storeInitialized) {
    mcpStore = generateMockMcps()
    storeInitialized = true
  }
  return mcpStore
}

function nextId(): number {
  const store = ensureStore()
  const ids = store.map((m) => m.mcpId)
  return Math.max(0, ...ids) + 1
}

/* ============================================
   Mock data generators
   ============================================ */

function generateMockMcps(): AdminMcpItem[] {
  const now = Date.now()
  return [
    {
      mcpId: 1,
      name: 'filesystem',
      displayName: '文件系统',
      icon: 'folder',
      description: '访问本地文件系统，读写文件',
      transport: 'Stdio',
      httpUrl: '',
      httpHeader: '',
      httpProxyUrl: '',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '["@modelcontextprotocol/server-filesystem","/tmp"]',
      status: 1,
      healthLatency: 45,
      healthCheckedAt: new Date(now - 60000).toISOString(),
      remark: '',
      created: '2025-06-01T08:00:00Z',
      updated: '2025-06-10T14:30:00Z',
    },
    {
      mcpId: 2,
      name: 'github',
      displayName: 'GitHub',
      icon: 'git-branch',
      description: '管理 GitHub 仓库、Issues、PR',
      transport: 'Stdio',
      httpUrl: '',
      httpHeader: '',
      httpProxyUrl: '',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '["@modelcontextprotocol/server-github"]',
      status: 1,
      healthLatency: 120,
      healthCheckedAt: new Date(now - 120000).toISOString(),
      remark: '',
      created: '2025-06-02T09:00:00Z',
      updated: '2025-06-09T11:00:00Z',
    },
    {
      mcpId: 3,
      name: 'postgres',
      displayName: 'PostgreSQL',
      icon: 'database',
      description: 'PostgreSQL 数据库查询和分析',
      transport: 'Stdio',
      httpUrl: '',
      httpHeader: '',
      httpProxyUrl: '',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '["@modelcontextprotocol/server-postgres","postgresql://localhost:5432/mydb"]',
      status: 1,
      healthLatency: 80,
      healthCheckedAt: new Date(now - 300000).toISOString(),
      remark: '',
      created: '2025-06-03T10:00:00Z',
      updated: '2025-06-08T16:00:00Z',
    },
    {
      mcpId: 4,
      name: 'slack',
      displayName: 'Slack',
      icon: 'message-square',
      description: '发送和读取 Slack 消息',
      transport: 'Stdio',
      httpUrl: '',
      httpHeader: '',
      httpProxyUrl: '',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '["@modelcontextprotocol/server-slack"]',
      status: 0,
      healthLatency: 0,
      healthCheckedAt: '',
      remark: '',
      created: '2025-06-04T08:00:00Z',
      updated: '2025-06-04T08:00:00Z',
    },
    {
      mcpId: 5,
      name: 'brave-search',
      displayName: 'Brave 搜索',
      icon: 'search',
      description: '通过 Brave Search API 搜索网页',
      transport: 'SSE',
      httpUrl: 'https://brave-mcp.example.com/sse',
      httpHeader: '{"Authorization":"Bearer sk-brave-xxx"}',
      httpProxyUrl: '',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '',
      status: 1,
      healthLatency: 210,
      healthCheckedAt: new Date(now - 90000).toISOString(),
      remark: '',
      created: '2025-06-05T10:00:00Z',
      updated: '2025-06-07T09:00:00Z',
    },
    {
      mcpId: 6,
      name: 'time',
      displayName: '时间服务',
      icon: 'clock',
      description: '获取当前时间和时区转换',
      transport: 'Stdio',
      httpUrl: '',
      httpHeader: '',
      httpProxyUrl: '',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '["@modelcontextprotocol/server-time"]',
      status: 1,
      healthLatency: 35,
      healthCheckedAt: new Date(now - 50000).toISOString(),
      remark: '',
      created: '2025-06-06T08:00:00Z',
      updated: '2025-06-06T08:00:00Z',
    },
    {
      mcpId: 7,
      name: 'puppeteer',
      displayName: 'Puppeteer',
      icon: 'globe',
      description: '无头浏览器，截图和网页抓取',
      transport: 'Stdio',
      httpUrl: '',
      httpHeader: '',
      httpProxyUrl: '',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '["@modelcontextprotocol/server-puppeteer"]',
      status: 1,
      healthLatency: 340,
      healthCheckedAt: new Date(now - 180000).toISOString(),
      remark: '',
      created: '2025-06-07T08:00:00Z',
      updated: '2025-06-07T08:00:00Z',
    },
    {
      mcpId: 8,
      name: 'memory',
      displayName: '记忆存储',
      icon: 'brain',
      description: '持久化记忆和图谱知识',
      transport: 'Stdio',
      httpUrl: '',
      httpHeader: '',
      httpProxyUrl: '',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '["@modelcontextprotocol/server-memory"]',
      status: 1,
      healthLatency: 55,
      healthCheckedAt: new Date(now - 200000).toISOString(),
      remark: '',
      created: '2025-06-08T08:00:00Z',
      updated: '2025-06-08T08:00:00Z',
    },
    {
      mcpId: 9,
      name: 'fetch',
      displayName: 'Web Fetch',
      icon: 'download',
      description: '抓取网页内容并转换为 Markdown',
      transport: 'StreamableHttp',
      httpUrl: 'https://fetch-mcp.example.com/mcp',
      httpHeader: '{"Authorization":"Bearer sk-xxxxx"}',
      httpProxyUrl: '',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '',
      status: 1,
      healthLatency: 150,
      healthCheckedAt: new Date(now - 70000).toISOString(),
      remark: '',
      created: '2025-06-09T08:00:00Z',
      updated: '2025-06-09T08:00:00Z',
    },
    {
      mcpId: 10,
      name: 'cloudflare',
      displayName: 'Cloudflare',
      icon: 'cloud',
      description: '管理 Cloudflare Workers 和 KV',
      transport: 'SSE',
      httpUrl: 'https://cf-mcp.example.com/sse',
      httpHeader: '{"CF-API-Token":"cf-xxx"}',
      httpProxyUrl: '',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '',
      status: 0,
      healthLatency: 0,
      healthCheckedAt: new Date(now - 3600000).toISOString(),
      remark: '',
      created: '2025-06-10T08:00:00Z',
      updated: '2025-06-10T15:00:00Z',
    },
    {
      mcpId: 11,
      name: 'everart',
      displayName: 'EverArt 图片生成',
      icon: 'image',
      description: '通过 EverArt API 生成 AI 图片',
      transport: 'SSE',
      httpUrl: 'https://everart-mcp.example.com/sse',
      httpHeader: '',
      httpProxyUrl: 'http://proxy.internal:8080',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '',
      status: 1,
      healthLatency: 280,
      healthCheckedAt: new Date(now - 40000).toISOString(),
      remark: '',
      created: '2025-06-11T08:00:00Z',
      updated: '2025-06-11T08:00:00Z',
    },
    {
      mcpId: 12,
      name: 'sequential-thinking',
      displayName: '顺序思考',
      icon: 'lightbulb',
      description: '通过顺序思考增强推理能力',
      transport: 'Stdio',
      httpUrl: '',
      httpHeader: '',
      httpProxyUrl: '',
      stdioType: 'npx',
      stdioEnv: '',
      stdioArgs: '["@modelcontextprotocol/server-sequential-thinking"]',
      status: 1,
      healthLatency: 60,
      healthCheckedAt: new Date(now - 150000).toISOString(),
      remark: '',
      created: '2025-06-12T08:00:00Z',
      updated: '2025-06-12T08:00:00Z',
    },
  ]
}

export function generateRecommendMcps(): McpRecommendItem[] {
  return [
    {
      name: 'filesystem',
      displayName: '文件系统',
      icon: 'folder',
      description: '访问本地文件系统',
      transport: 'Stdio',
      httpUrl: '',
      stdioType: 'npx',
      stdioArgs: '["@modelcontextprotocol/server-filesystem","/tmp"]',
      installed: true,
      mcpId: 1,
      mcpStatus: 1,
    },
    {
      name: 'github',
      displayName: 'GitHub',
      icon: 'git-branch',
      description: '管理 GitHub 仓库和 PR',
      transport: 'Stdio',
      httpUrl: '',
      stdioType: 'npx',
      stdioArgs: '["@modelcontextprotocol/server-github"]',
      installed: true,
      mcpId: 2,
      mcpStatus: 1,
    },
    {
      name: 'postgres',
      displayName: 'PostgreSQL',
      icon: 'database',
      description: '数据库查询和分析',
      transport: 'Stdio',
      httpUrl: '',
      stdioType: 'npx',
      stdioArgs: '["@modelcontextprotocol/server-postgres","postgresql://localhost:5432/mydb"]',
      installed: true,
      mcpId: 3,
      mcpStatus: 1,
    },
    {
      name: 'brave-search',
      displayName: 'Brave 搜索',
      icon: 'search',
      description: '网页搜索',
      transport: 'SSE',
      httpUrl: 'https://brave-mcp.example.com/sse',
      stdioType: 'npx',
      stdioArgs: '[]',
      installed: true,
      mcpId: 5,
      mcpStatus: 1,
    },
    {
      name: 'slack',
      displayName: 'Slack',
      icon: 'message-square',
      description: '发送和读取 Slack 消息',
      transport: 'Stdio',
      httpUrl: '',
      stdioType: 'npx',
      stdioArgs: '["@modelcontextprotocol/server-slack"]',
      installed: true,
      mcpId: 4,
      mcpStatus: 0,
    },
    {
      name: 'puppeteer',
      displayName: 'Puppeteer',
      icon: 'globe',
      description: '无头浏览器，网页截图和抓取',
      transport: 'Stdio',
      httpUrl: '',
      stdioType: 'npx',
      stdioArgs: '["@modelcontextprotocol/server-puppeteer"]',
      installed: true,
      mcpId: 7,
      mcpStatus: 1,
    },
    {
      name: 'memory',
      displayName: '记忆存储',
      icon: 'brain',
      description: '持久化记忆和图谱知识',
      transport: 'Stdio',
      httpUrl: '',
      stdioType: 'npx',
      stdioArgs: '["@modelcontextprotocol/server-memory"]',
      installed: true,
      mcpId: 8,
      mcpStatus: 1,
    },
    {
      name: 'fetch',
      displayName: 'Web Fetch',
      icon: 'download',
      description: '抓取网页内容',
      transport: 'StreamableHttp',
      httpUrl: 'https://fetch-mcp.example.com/mcp',
      stdioType: 'npx',
      stdioArgs: '[]',
      installed: true,
      mcpId: 9,
      mcpStatus: 1,
    },
    {
      name: 'everart',
      displayName: 'EverArt',
      icon: 'image',
      description: 'AI 图片生成',
      transport: 'SSE',
      httpUrl: 'https://everart-mcp.example.com/sse',
      stdioType: 'npx',
      stdioArgs: '[]',
      installed: true,
      mcpId: 11,
      mcpStatus: 1,
    },
    {
      name: 'exa',
      displayName: 'Exa 搜索',
      icon: 'search',
      description: '互联网搜索',
      transport: 'SSE',
      httpUrl: 'https://exa-mcp.example.com/sse',
      stdioType: 'npx',
      stdioArgs: '[]',
      installed: false,
      mcpId: 0,
      mcpStatus: 0,
    },
    {
      name: 'sqlite',
      displayName: 'SQLite',
      icon: 'database',
      description: 'SQLite 数据库操作',
      transport: 'Stdio',
      httpUrl: '',
      stdioType: 'uvx',
      stdioArgs: '["mcp-server-sqlite"]',
      installed: false,
      mcpId: 0,
      mcpStatus: 0,
    },
    {
      name: 'weather',
      displayName: '天气查询',
      icon: 'cloud',
      description: '获取天气信息',
      transport: 'Stdio',
      httpUrl: '',
      stdioType: 'npx',
      stdioArgs: '["@h1deya/mcp-server-weather"]',
      installed: false,
      mcpId: 0,
      mcpStatus: 0,
    },
  ]
}

/* ============================================
   Async mock functions
   ============================================ */

export async function getMcps(params?: {
  page?: number
  pageSize?: number
  search?: string
  transport?: string
}): Promise<PaginatedResponse<AdminMcpItem>> {
  await listDelay()

  const store = ensureStore()
  let filtered = [...store]

  if (params?.search) {
    const q = params.search.toLowerCase()
    filtered = filtered.filter(
      (m) =>
        m.name.toLowerCase().includes(q) ||
        m.displayName.toLowerCase().includes(q),
    )
  }

  if (params?.transport && params.transport !== 'all') {
    filtered = filtered.filter((m) => m.transport === params.transport)
  }

  const page = params?.page ?? 1
  const pageSize = params?.pageSize ?? PAGE_SIZE
  const totalCount = filtered.length
  const totalPage = Math.max(1, Math.ceil(totalCount / pageSize))
  const start = (page - 1) * pageSize
  const list = filtered.slice(start, start + pageSize)

  return { list, totalPage, totalCount, page, pageSize }
}

export async function getMcpDetail(id: number): Promise<AdminMcpItem> {
  await itemDelay()

  const store = ensureStore()
  const item = store.find((m) => m.mcpId === id)
  if (!item) {
    throw new Error('MCP 端点不存在')
  }

  return { ...item }
}

export async function createMcp(req: {
  name: string
  displayName: string
  icon?: string
  description?: string
  transport: string
  httpUrl?: string
  httpHeader?: string
  httpProxyUrl?: string
  stdioType?: string
  stdioEnv?: string
  stdioArgs?: string
  remark?: string
}): Promise<AdminMcpItem> {
  await mutationDelay()

  const store = ensureStore()
  const now = new Date().toISOString()
  const newItem: AdminMcpItem = {
    mcpId: nextId(),
    name: req.name,
    displayName: req.displayName || req.name,
    icon: req.icon ?? 'plug',
    description: req.description ?? '',
    transport: req.transport,
    httpUrl: req.httpUrl ?? '',
    httpHeader: req.httpHeader ?? '',
    httpProxyUrl: req.httpProxyUrl ?? '',
    stdioType: req.stdioType ?? 'npx',
    stdioEnv: req.stdioEnv ?? '',
    stdioArgs: req.stdioArgs ?? '',
    status: 1,
    healthLatency: Math.floor(Math.random() * 300) + 20,
    healthCheckedAt: now,
    remark: req.remark ?? '',
    created: now,
    updated: now,
  }
  store.unshift(newItem)
  return { ...newItem }
}

export async function updateMcp(
  id: number,
  req: {
    displayName?: string
    icon?: string
    description?: string
    httpUrl?: string
    httpHeader?: string
    httpProxyUrl?: string
    stdioType?: string
    stdioEnv?: string
    stdioArgs?: string
    remark?: string
    status?: number
  },
): Promise<AdminMcpItem> {
  await mutationDelay()

  const store = ensureStore()
  const index = store.findIndex((m) => m.mcpId === id)
  if (index === -1) {
    throw new Error('MCP 端点不存在')
  }

  const item = store[index]
  const updated: AdminMcpItem = {
    ...item,
    displayName: req.displayName ?? item.displayName,
    icon: req.icon ?? item.icon,
    description: req.description ?? item.description,
    httpUrl: req.httpUrl ?? item.httpUrl,
    httpHeader: req.httpHeader ?? item.httpHeader,
    httpProxyUrl: req.httpProxyUrl ?? item.httpProxyUrl,
    stdioType: req.stdioType ?? item.stdioType,
    stdioEnv: req.stdioEnv ?? item.stdioEnv,
    stdioArgs: req.stdioArgs ?? item.stdioArgs,
    remark: req.remark ?? item.remark,
    status: req.status ?? item.status,
    updated: new Date().toISOString(),
  }
  store[index] = updated
  return { ...updated }
}

export async function toggleMcpStatus(
  id: number,
  status: number,
): Promise<AdminMcpItem> {
  await mutationDelay()

  const store = ensureStore()
  const index = store.findIndex((m) => m.mcpId === id)
  if (index === -1) {
    throw new Error('MCP 端点不存在')
  }

  store[index] = {
    ...store[index],
    status,
    updated: new Date().toISOString(),
  }
  return { ...store[index] }
}

export async function deleteMcp(id: number): Promise<void> {
  await mutationDelay()

  const store = ensureStore()
  const index = store.findIndex((m) => m.mcpId === id)
  if (index === -1) {
    throw new Error('MCP 端点不存在')
  }
  store.splice(index, 1)
}

export async function getRecommendMcps(): Promise<McpRecommendItem[]> {
  await listDelay()
  return generateRecommendMcps()
}

export async function healthCheck(): Promise<{ status: string }> {
  await heavyDelay()
  return { status: 'checking' }
}
