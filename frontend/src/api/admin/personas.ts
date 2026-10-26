import http from '../http'
import type { PaginatedResponse, AdminPersonaTemplateItem, AdminPersonaCategoryItem } from '../types'

/* ---------- Persona templates ---------- */

/** GET /api/admin/personaTemplates */
export function getPersonaTemplates(params?: { search?: string; categoryId?: number; page?: number; pageSize?: number }) {
  return http.get<PaginatedResponse<AdminPersonaTemplateItem>>('/admin/personaTemplates', { params })
}

/** GET /api/admin/personaTemplates/:id */
export function getPersonaTemplateDetail(id: number) {
  return http.get<AdminPersonaTemplateItem & { roleInfo: string }>(`/admin/personaTemplates/${id}`)
}

/** POST /api/admin/personaTemplates */
export function createPersonaTemplate(data: { name: string; icon?: string; description?: string; roleInfo: string; categoryId: number; sortOrder?: number }) {
  return http.post<{ status: string }>('/admin/personaTemplates', data)
}

/** PUT /api/admin/personaTemplates/:id */
export function updatePersonaTemplate(id: number, data: { name?: string; icon?: string; description?: string; roleInfo?: string; categoryId?: number; sortOrder?: number }) {
  return http.put<{ status: string }>(`/admin/personaTemplates/${id}`, data)
}

/** DELETE /api/admin/personaTemplates/:id */
export function deletePersonaTemplate(id: number) {
  return http.delete<{ status: string }>(`/admin/personaTemplates/${id}`)
}

/* ---------- Persona categories ---------- */

/** GET /api/admin/personaCategories */
export function getPersonaCategories(params?: { search?: string; page?: number; pageSize?: number }) {
  return http.get<PaginatedResponse<AdminPersonaCategoryItem>>('/admin/personaCategories', { params })
}

/** GET /api/admin/personaCategories/all */
export function getAllPersonaCategories() {
  return http.get<{ list: AdminPersonaCategoryItem[] }>('/admin/personaCategories/all')
}

/** GET /api/admin/personaCategories/:id */
export function getPersonaCategoryDetail(id: number) {
  return http.get<AdminPersonaCategoryItem>(`/admin/personaCategories/${id}`)
}

/** POST /api/admin/personaCategories */
export function createPersonaCategory(data: { name: string; icon?: string }) {
  return http.post<{ status: string }>('/admin/personaCategories', data)
}

/** PUT /api/admin/personaCategories/:id */
export function updatePersonaCategory(id: number, data: { name?: string; icon?: string }) {
  return http.put<{ status: string }>(`/admin/personaCategories/${id}`, data)
}

/** DELETE /api/admin/personaCategories/:id */
export function deletePersonaCategory(id: number) {
  return http.delete<{ status: string }>(`/admin/personaCategories/${id}`)
}
