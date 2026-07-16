import http from '../http'
import type { PaginatedResponse, AdminSharedProjectItem, AdminSharedProjectListReq } from '../types'

/** GET /api/admin/sharedProjects */
export function getSharedProjects(params?: AdminSharedProjectListReq) {
  return http.get<PaginatedResponse<AdminSharedProjectItem>>('/admin/sharedProjects', { params })
}

/** DELETE /api/admin/sharedProjects/:id */
export function unshareProject(id: number) {
  return http.delete(`/admin/sharedProjects/${id}`)
}
