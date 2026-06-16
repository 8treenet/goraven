import http from './http'
import type { PaginatedResponse, SessionSimple, SessionDetail, ShareInfo, CreateShareRequest } from './types'

/** GET /api/sessions */
export function getSessions(params?: { page?: number; pageSize?: number }) {
  return http.get<PaginatedResponse<SessionSimple>>('/sessions', { params })
}

/** GET /api/sessions/:id */
export function getSessionDetail(sessionId: string) {
  return http.get<SessionDetail>(`/sessions/${sessionId}`)
}

/** PUT /api/sessions/:id */
export function updateSession(sessionId: string, data: { title?: string; personaId?: number; isArchived?: number }) {
  return http.put(`/sessions/${sessionId}`, data)
}

/** DELETE /api/sessions/:id */
export function deleteSession(sessionId: string) {
  return http.delete(`/sessions/${sessionId}`)
}

/** POST /api/sessions/:id/share */
export function createShare(sessionId: string, data: CreateShareRequest) {
  return http.post<ShareInfo>(`/sessions/${sessionId}/share`, data)
}

/** GET /api/sessions/:id/share */
export function getShare(sessionId: string) {
  return http.get<ShareInfo>(`/sessions/${sessionId}/share`)
}

/** DELETE /api/sessions/:id/share */
export function deleteShare(sessionId: string) {
  return http.delete(`/sessions/${sessionId}/share`)
}

/** GET /api/sessions/my-shares */
export function getMyShares(params?: { page?: number; pageSize?: number }) {
  return http.get<PaginatedResponse<ShareInfo>>('/sessions/my-shares', { params })
}
