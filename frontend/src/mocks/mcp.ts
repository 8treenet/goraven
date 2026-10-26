/* ============================================
   Mock data — MCP Endpoints
   GET /api/mcp
   ============================================ */

import type { McpEndpoint } from './types'
import { listDelay } from './delay'

/* ---------- Mock data ---------- */

export const MOCK_MCP_ENDPOINTS: McpEndpoint[] = [
  { mcpId: 1, name: 'filesystem', displayName: '文件系统', icon: 'file-text', description: '文件系统读写操作' },
  { mcpId: 2, name: 'brave-search', displayName: 'Brave 搜索', icon: 'search', description: 'Brave 网页搜索' },
  { mcpId: 3, name: 'github', displayName: 'GitHub', icon: 'code', description: 'GitHub API 交互' },
  { mcpId: 4, name: 'postgres', displayName: 'PostgreSQL', icon: 'database', description: 'PostgreSQL 数据库查询' },
  { mcpId: 5, name: 'redis', displayName: 'Redis', icon: 'terminal', description: 'Redis 缓存操作' },
  { mcpId: 6, name: 'docker', displayName: 'Docker', icon: 'terminal', description: 'Docker 容器管理' },
  { mcpId: 7, name: 'slack', displayName: 'Slack', icon: 'bot', description: 'Slack 消息推送' },
]

/* ---------- Async function ---------- */

export async function getMcpEndpoints(): Promise<McpEndpoint[]> {
  await listDelay()
  return [...MOCK_MCP_ENDPOINTS]
}
