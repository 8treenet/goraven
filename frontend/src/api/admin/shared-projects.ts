import http from '../http'
import type { PaginatedResponse, AdminTeamProjectItem, AdminTeamProjectListReq } from '../types'

/** GET /api/admin/sharedProjects */
export function getTeamProjects(params?: AdminTeamProjectListReq) {
  return http.get<PaginatedResponse<AdminTeamProjectItem>>('/admin/sharedProjects', { params })
}

/** PUT /api/admin/sharedProjects/:id */
export function updateTeamProject(id: number, data: { description: string }) {
  return http.put(`/admin/sharedProjects/${id}`, data)
}

/** DELETE /api/admin/sharedProjects/:id */
export function deleteTeamProject(id: number) {
  return http.delete(`/admin/sharedProjects/${id}`)
}
