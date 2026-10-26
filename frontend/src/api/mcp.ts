import http from './http'
import type { McpInfo } from './types'

/** GET /api/mcp */
export function getMcpEndpoints() {
  return http.get<McpInfo[]>('/mcp')
}

/** GET /api/mcp/byIds?ids=1&ids=2 */
export function getMcpEndpointsByIDs(ids: number[]) {
  return http.get<McpInfo[]>('/mcp/byIds', { params: { ids }, paramsSerializer: { indexes: null } })
}
