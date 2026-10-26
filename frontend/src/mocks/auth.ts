/* ============================================
   Mock data — Auth / Install
   Extracted from features/auth/
   ============================================ */

import type {
  LoginRequest,
  LoginResponse,
  CheckDbRequest,
  CheckRedisRequest,
  InitRequest,
  InitResponse,
} from './types'
import { delay } from './delay'

/* ---------- Login ---------- */

export async function login(req: LoginRequest): Promise<LoginResponse> {
  await delay(400 + Math.random() * 400) // 400-800ms
  if (req.username !== 'admin' || req.password !== 'admin123') {
    throw new Error('账号或密码错误')
  }
  return { accessToken: 'rvn_mock_token' }
}

/* ---------- Install ---------- */

export async function checkDb(_req: CheckDbRequest): Promise<void> {
  await delay(800 + Math.random() * 600) // 800-1400ms
}

export async function checkRedis(_req: CheckRedisRequest): Promise<void> {
  await delay(600 + Math.random() * 400) // 600-1000ms
}

export async function initSystem(req: InitRequest): Promise<InitResponse> {
  await delay(1500)
  return { userId: '1', username: req.username }
}

export async function ping(): Promise<boolean> {
  await delay(2000 + Math.random() * 3000) // 2000-5000ms
  return true
}
