/* ============================================
   Mock data — User / Profile
   Extracted from features/profile/ProfilePage.tsx
   ============================================ */

import type { UserInfo, UpdateProfileRequest, ChangePasswordRequest } from './types'
import { itemDelay, mutationDelay } from './delay'

/* ---------- Mock user data ---------- */

export const MOCK_USER: UserInfo = {
  userId: '1',
  username: 'admin',
  nickname: 'Admin',
  email: 'admin@example.com',
  avatar: '',
  role: 1,
  status: 1,
  created: '2024-01-15T08:00:00Z',
}

/* ---------- Async mock functions ---------- */

export async function getProfile(): Promise<UserInfo> {
  await itemDelay()
  return { ...MOCK_USER }
}

export async function updateProfile(req: UpdateProfileRequest): Promise<UserInfo> {
  await mutationDelay()
  if (req.nickname !== undefined) {
    MOCK_USER.nickname = req.nickname
  }
  if (req.email !== undefined) {
    MOCK_USER.email = req.email
  }
  if (req.avatar !== undefined) {
    MOCK_USER.avatar = req.avatar
  }
  return { ...MOCK_USER }
}

export async function changePassword(_req: ChangePasswordRequest): Promise<void> {
  await mutationDelay()
  if (!_req.currentPassword) {
    throw new Error('当前密码不能为空')
  }
}

export async function logout(): Promise<void> {
  await itemDelay()
}
