import http from './http'
import type { User, CaptchaRsp, CheckDbRequest, CheckRedisRequest, InitRequest, InitResponse } from './types'

/** POST /api/user/login */
export function login(data: { username: string; password: string; captchaAnswer?: number }) {
  return http.post<{ accessToken: string }>('/user/login', data)
}

/** GET /api/user/captcha?username=xxx */
export function getCaptcha(username: string) {
  return http.get<CaptchaRsp>('/user/captcha', { params: { username } })
}

/** GET /api/user/ */
export function getCurrentUser() {
  return http.get<User>('/user/')
}

/** PUT /api/user/profile */
export function updateProfile(data: { nickname?: string; email?: string; avatar?: string }) {
  return http.put('/user/profile', data)
}

/** PUT /api/user/password */
export function changePassword(data: { currentPassword: string; newPassword: string }) {
  return http.put('/user/password', data)
}

/** POST /api/user/logout */
export function logout() {
  return http.post('/user/logout')
}


/** POST /api/install/check-db */
export function checkDb(data: CheckDbRequest) {
  return http.post('/install/check-db', data)
}

/** POST /api/install/check-redis */
export function checkRedis(data: CheckRedisRequest) {
  return http.post('/install/check-redis', data)
}

/** POST /api/install/init */
export function initSystem(data: InitRequest) {
  return http.post<InitResponse>('/install/init', data)
}
