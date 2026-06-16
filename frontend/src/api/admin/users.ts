import http from '../http'
import type { PaginatedResponse, AdminUserItem } from '../types'

/** GET /api/admin/users */
export function getUsers(params?: { search?: string; role?: number; page?: number; pageSize?: number }) {
  return http.get<PaginatedResponse<AdminUserItem>>('/admin/users', { params })
}

/** POST /api/admin/users */
export function createUser(data: { username: string; password: string; nickname?: string; role?: number }) {
  return http.post<AdminUserItem>('/admin/users', data)
}

/** POST /api/admin/users/batch */
export function batchGetUsers(data: { userIds: string[] }) {
  return http.post<AdminUserItem[]>('/admin/users/batch', data)
}

/** GET /api/admin/users/:userId */
export function getUserDetail(userId: string) {
  return http.get<AdminUserItem>(`/admin/users/${userId}`)
}

/** PUT /api/admin/users/:userId */
export function updateUser(userId: string, data: { nickname?: string; email?: string; role?: number; status?: number }) {
  return http.put(`/admin/users/${userId}`, data)
}

/** PUT /api/admin/users/:userId/reset-password */
export function resetPassword(userId: string, data: { password: string }) {
  return http.put(`/admin/users/${userId}/reset-password`, data)
}

/** DELETE /api/admin/users/:userId */
export function deleteUser(userId: string) {
  return http.delete(`/admin/users/${userId}`)
}
