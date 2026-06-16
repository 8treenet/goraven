import http from '../http'
import type { PaginatedResponse, AdminMcpItem } from '../types'

/** GET /api/admin/mcp */
export function getMCPs(params?: { search?: string; transport?: string; status?: number; page?: number; pageSize?: number }) {
  return http.get<PaginatedResponse<AdminMcpItem>>('/admin/mcp', { params })
}

/** GET /api/admin/mcp/:id */
export function getMCPDetail(id: number) {
  return http.get<AdminMcpItem>(`/admin/mcp/${id}`)
}

/** POST /api/admin/mcp */
export function createMCP(data: {
  name: string
  displayName?: string
  icon?: string
  description?: string
  transport: string
  httpUrl?: string
  httpHeader?: Record<string, string>
  httpProxyUrl?: string
  stdioType?: string
  stdioEnv?: Record<string, string>
  stdioArgs?: string[]
  remark?: string
}) {
  return http.post<AdminMcpItem>('/admin/mcp', data)
}

/** PUT /api/admin/mcp/:id */
export function updateMCP(
  id: number,
  data: {
    name?: string
    displayName?: string
    icon?: string
    description?: string
    httpUrl?: string
    httpHeader?: Record<string, string>
    httpProxyUrl?: string
    stdioType?: string
    stdioEnv?: Record<string, string>
    stdioArgs?: string[]
    status?: number
    remark?: string
  },
) {
  return http.put(`/admin/mcp/${id}`, data)
}

/** DELETE /api/admin/mcp/:id */
export function deleteMCP(id: number) {
  return http.delete(`/admin/mcp/${id}`)
}

/** PUT /api/admin/mcp/:id/status */
export function updateMCPStatus(id: number, status: number) {
  return http.put(`/admin/mcp/${id}/status`, { status })
}

/** GET /api/admin/mcp/recommend */
export function getRecommendMCPs() {
  return http.get<{
    list: Array<{
      name: string
      displayName: string
      icon: string
      description: string
      transport: string
      httpUrl: string
      httpHeader: string
      stdioType: string
      stdioArgs: string
      stdioEnv: string
      installed: boolean
      mcpId?: number
      mcpStatus?: number
    }>
  }>('/admin/mcp/recommend')
}

/** POST /api/admin/mcp/healthCheck */
export function checkMCPHealth() {
  return http.post('/admin/mcp/healthCheck')
}
