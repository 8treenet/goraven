import http from '../http'
import type { PaginatedResponse } from '../types'
import type { AdminModelItem, AdminModelDetail } from '../types'

/** GET /api/admin/models */
export function getModels(params?: { search?: string; providerId?: string; page?: number; pageSize?: number }) {
  return http.get<PaginatedResponse<AdminModelItem>>('/admin/models', { params })
}

/** POST /api/admin/models */
export function createModel(data: {
  providerDisplayName: string
  displayName?: string
  providerId: string
  modelName: string
  icon?: string
  apiKey?: string
  baseUrl?: string
  proxyUrl?: string
  conversationHeaderKey?: string
  contextLen?: number
  extraFields?: string
  isDefault?: number
  isFlash?: number
  isVisual?: number
  remark?: string
}) {
  return http.post<AdminModelItem>('/admin/models', data)
}

/** PUT /api/admin/models/:id */
export function updateModel(
  id: number,
  data: {
    providerDisplayName?: string
    displayName?: string
    modelName?: string
    icon?: string
    apiKey?: string
    baseUrl?: string
    proxyUrl?: string
    conversationHeaderKey?: string
    contextLen?: number
    extraFields?: string
    isDefault?: number
    isFlash?: number
    isVisual?: number
    status?: number
    remark?: string
  },
) {
  return http.put(`/admin/models/${id}`, data)
}

/** DELETE /api/admin/models/:id */
export function deleteModel(id: number) {
  return http.delete(`/admin/models/${id}`)
}

/** GET /api/admin/models/:id */
export function getModelDetail(id: number) {
  return http.get<AdminModelDetail>(`/admin/models/${id}`)
}

/** PUT /api/admin/models/:id/status */
export function updateModelStatus(id: number, status: number) {
  return http.put(`/admin/models/${id}/status`, { status })
}

/** PUT /api/admin/models/:id/default */
export function setDefaultModel(id: number) {
  return http.put(`/admin/models/${id}/default`)
}

/** PUT /api/admin/models/:id/flash */
export function setFlashModel(id: number) {
  return http.put(`/admin/models/${id}/flash`)
}

/** PUT /api/admin/models/:id/visual */
export function setVisualModel(id: number) {
  return http.put(`/admin/models/${id}/visual`)
}

/** GET /api/admin/models/:id/members — 查询模型可见成员 */
export function getModelMembers(id: number) {
  return http.get<{ memberIds: string[] }>(`/admin/models/${id}/members`)
}

/** PUT /api/admin/models/:id/members — 编辑模型成员 */
export function updateModelMembers(id: number, addUserIds: string[], removeUserIds: string[]) {
  return http.put(`/admin/models/${id}/members`, { addUserIds, removeUserIds })
}

/** PUT /api/admin/models/:id/access — 设置模型访问权限 */
export function updateModelAccess(id: number, access: number) {
  return http.put(`/admin/models/${id}/access`, { access })
}
