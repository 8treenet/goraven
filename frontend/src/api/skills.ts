import http from './http'
import type {
  PaginatedResponse,
  SkillSimple,
  MarketSkill,
  MarketSkillDetail,
  UserSkill,
  UserSkillDetail,
  SkillCategory,
  InstallSkillResult,
  UserSkillStatus,
  RefreshSkillsResult,
  ShareSkill,
  ShareSkillResult,
  ShareSkillDetail,
} from './types'

/** GET /api/skills/simpleSkills */
export function getSimpleSkills() {
  return http.get<SkillSimple[]>('/skills/simpleSkills')
}

/** GET /api/skills/simpleSkills/byIds?ids=1&ids=2 */
export function getSimpleSkillsByIDs(ids: number[]) {
  return http.get<SkillSimple[]>('/skills/simpleSkills/byIds', { params: { ids }, paramsSerializer: { indexes: null } })
}

/** GET /api/skills/market */
export function getMarketSkills(params?: {
  search?: string
  categoryId?: number
  source?: string
  page?: number
  pageSize?: number
}) {
  return http.get<PaginatedResponse<MarketSkill>>('/skills/market', { params })
}

/** GET /api/skills/market/:id */
export function getMarketSkillDetail(id: number) {
  return http.get<MarketSkillDetail>(`/skills/market/${id}`)
}

/** GET /api/skills/user */
export function getUserSkills(params?: {
  search?: string
  categoryId?: number
  source?: string
  status?: number
  page?: number
  pageSize?: number
}) {
  return http.get<PaginatedResponse<UserSkill>>('/skills/user', { params })
}

/** GET /api/skills/user/:id */
export function getUserSkillDetail(id: number) {
  return http.get<UserSkillDetail>(`/skills/user/${id}`)
}

/** PUT /api/skills/user/:id */
export function updateUserSkill(id: number, req: { icon?: string; categoryId?: number }) {
  return http.put(`/skills/user/${id}`, req)
}

/** DELETE /api/skills/user/:id */
export function deleteUserSkill(id: number) {
  return http.delete(`/skills/user/${id}`)
}

/** POST /api/skills/user/refresh */
export function refreshSkills() {
  return http.post<RefreshSkillsResult>('/skills/user/refresh')
}

/** POST /api/skills/install */
export function installSkill(skillId: number) {
  return http.post<InstallSkillResult>('/skills/install', { skillId })
}

/** GET /api/skills/user/:id/status */
export function getUserSkillStatus(id: number) {
  return http.get<UserSkillStatus>(`/skills/user/${id}/status`)
}

/** PUT /api/skills/user/:id/retry */
export function retryInstall(id: number) {
  return http.put(`/skills/user/${id}/retry`)
}

/** PUT /api/skills/user/:id/alwaysOn */
export function toggleAlwaysOn(id: number, alwaysOn: number) {
  return http.put(`/skills/user/${id}/alwaysOn`, { alwaysOn })
}

/** GET /api/skills/shares */
export function getShareSkills(params?: {
  search?: string
  page?: number
  pageSize?: number
}) {
  return http.get<PaginatedResponse<ShareSkill>>('/skills/shares', { params })
}

/** POST /api/skills/shares */
export function shareSkillToTeam(userSkillId: number, note?: string) {
  return http.post<ShareSkillResult>('/skills/shares', { userSkillId, note })
}

/** POST /api/skills/shares/:id/install */
export function installShareSkill(shareId: number) {
  return http.post<InstallSkillResult>(`/skills/shares/${shareId}/install`)
}

/** DELETE /api/skills/shares/:id */
export function deleteShareSkill(shareId: number) {
  return http.delete(`/skills/shares/${shareId}`)
}

/** GET /api/skills/shares/:id */
export function getShareSkillDetail(shareId: number) {
  return http.get<ShareSkillDetail>(`/skills/shares/${shareId}`)
}

/** GET /api/skills/categories */
export function getSkillCategories() {
  return http.get<SkillCategory[]>('/skills/categories')
}
