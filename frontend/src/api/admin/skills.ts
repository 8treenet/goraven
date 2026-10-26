import http from '../http'
import type {
  AdminMarketSkillDetail,
  AdminMarketSkillItem,
  AdminMarketSkillSource,
  AdminMarketSkillUserItem,
  AdminSkillCategoryDetail,
  AdminSkillCategoryItem,
  AdminSkillCategorySimple,
  AdminSkillStatus,
  AdminSystemSkillDetail,
  AdminSystemSkillItem,
  ClawHubExploreResponse,
  ClawHubSearchResponse,
  ClawHubSkillDetail,
  PaginatedResponse,
} from '../types'

/* ---------- System skills ---------- */

/** GET /api/admin/systemSkills */
export function getSystemSkills(params?: { search?: string; status?: AdminSkillStatus; page?: number; pageSize?: number }) {
  return http.get<PaginatedResponse<AdminSystemSkillItem>>('/admin/systemSkills', { params })
}

/** GET /api/admin/systemSkills/:id */
export function getSystemSkillDetail(id: number) {
  return http.get<AdminSystemSkillDetail>(`/admin/systemSkills/${id}`)
}

/** POST /api/admin/systemSkills */
export function createSystemSkill(data: { content: string }) {
  return http.post<{ status: string }>('/admin/systemSkills', data)
}

/** PUT /api/admin/systemSkills/:id */
export function updateSystemSkill(id: number, data: { content?: string }) {
  return http.put<{ status: string }>(`/admin/systemSkills/${id}`, data)
}

/** PUT /api/admin/systemSkills/:id/status */
export function updateSystemSkillStatus(id: number, status: AdminSkillStatus) {
  return http.put<{ status: string }>(`/admin/systemSkills/${id}/status`, { status })
}

/** DELETE /api/admin/systemSkills/:id */
export function deleteSystemSkill(id: number) {
  return http.delete<{ status: string }>(`/admin/systemSkills/${id}`)
}

/* ---------- Market skills ---------- */

/** GET /api/admin/marketSkills */
export function getMarketSkills(params?: {
  search?: string
  source?: AdminMarketSkillSource
  status?: AdminSkillStatus
  page?: number
  pageSize?: number
}) {
  return http.get<PaginatedResponse<AdminMarketSkillItem>>('/admin/marketSkills', { params })
}

/** GET /api/admin/marketSkills/:id */
export function getMarketSkillDetail(id: number) {
  return http.get<AdminMarketSkillDetail>(`/admin/marketSkills/${id}`)
}

/** PUT /api/admin/marketSkills/:id */
export function updateMarketSkill(
  id: number,
  data: { icon?: string; categoryId?: number; sortOrder?: number; remark?: string },
) {
  return http.put<{ status: string }>(`/admin/marketSkills/${id}`, data)
}

/** PUT /api/admin/marketSkills/:id/status */
export function updateMarketSkillStatus(id: number, status: AdminSkillStatus) {
  return http.put<{ status: string }>(`/admin/marketSkills/${id}/status`, { status })
}

/** DELETE /api/admin/marketSkills/:id */
export function deleteMarketSkill(id: number, cascade?: boolean) {
  return http.delete<{ status: string }>(`/admin/marketSkills/${id}`, {
    params: cascade === undefined ? undefined : { cascade },
  })
}

/** GET /api/admin/marketSkills/:id/users */
export function getMarketSkillUsers(id: number, params?: { page?: number; pageSize?: number }) {
  return http.get<PaginatedResponse<AdminMarketSkillUserItem>>(`/admin/marketSkills/${id}/users`, { params })
}

/** POST /api/admin/marketSkills/publish */
export function publishMarketSkill(data: { uploadId: string; icon?: string; categoryId: number }) {
  return http.post<{ status: string }>('/admin/marketSkills/publish', data)
}

/** POST /api/admin/marketSkills/import */
export function importClawHubSkill(data: { slug: string; icon?: string; categoryId: number }) {
  return http.post<{ status: string }>('/admin/marketSkills/import', data)
}

/* ---------- ClawHub ---------- */

/** GET /api/admin/clawhub/search */
export function searchClawHub(params: { q?: string; limit?: number }) {
  return http.get<ClawHubSearchResponse>('/admin/clawhub/search', { params })
}

/** GET /api/admin/clawhub/explore */
export function exploreClawHub(params?: { sort?: 'newest' | 'updated' | 'downloads' | 'stars' | 'installs' | 'trending' }) {
  return http.get<ClawHubExploreResponse>('/admin/clawhub/explore', { params })
}

/** GET /api/admin/clawhub/skills/:slug */
export function getClawHubSkillDetail(slug: string) {
  return http.get<ClawHubSkillDetail>(`/admin/clawhub/skills/${encodeURIComponent(slug)}`)
}

/* ---------- Skill categories ---------- */

/** GET /api/admin/skillCategories */
export function getSkillCategories(params?: { search?: string; page?: number; pageSize?: number }) {
  return http.get<PaginatedResponse<AdminSkillCategoryItem>>('/admin/skillCategories', { params })
}

/** GET /api/admin/skillCategories/all */
export function getAllSkillCategories() {
  return http.get<{ list: AdminSkillCategorySimple[] }>('/admin/skillCategories/all')
}

/** GET /api/admin/skillCategories/:id */
export function getSkillCategoryDetail(id: number) {
  return http.get<AdminSkillCategoryDetail>(`/admin/skillCategories/${id}`)
}

/** POST /api/admin/skillCategories */
export function createSkillCategory(data: { name: string; icon?: string }) {
  return http.post<{ status: string }>('/admin/skillCategories', data)
}

/** PUT /api/admin/skillCategories/:id */
export function updateSkillCategory(id: number, data: { name?: string; icon?: string }) {
  return http.put<{ status: string }>(`/admin/skillCategories/${id}`, data)
}

/** DELETE /api/admin/skillCategories/:id */
export function deleteSkillCategory(id: number) {
  return http.delete<{ status: string }>(`/admin/skillCategories/${id}`)
}
