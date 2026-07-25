/* ============================================
   Admin Users — mock data and handlers
   ============================================ */

import type { AdminUserItem, PaginatedResponse } from '../types'
import { listDelay, itemDelay, mutationDelay } from '../delay'

/* ============================================
   Constants
   ============================================ */

export const SUPER_ADMIN_ID = 'u-000'
export const PAGE_SIZE = 20

/* ============================================
   Generate mock users (42 users: u-000 – u-041)
   ============================================ */

function generateMockUsers(): AdminUserItem[] {
  const names = [
    { username: 'admin', nickname: '超级管理员', role: 1 as number, status: 1 as number },
    { username: 'alice', nickname: 'Alice Wang', role: 0 as number, status: 1 as number },
    { username: 'bob', nickname: 'Bob Chen', role: 0 as number, status: 1 as number },
    { username: 'carol', nickname: 'Carol Li', role: 0 as number, status: 0 as number },
    { username: 'dave', nickname: 'Dave Zhang', role: 0 as number, status: 1 as number },
    { username: 'eve', nickname: 'Eve Liu', role: 1 as number, status: 1 as number },
    { username: 'frank', nickname: 'Frank Wu', role: 0 as number, status: 0 as number },
    { username: 'grace', nickname: 'Grace Zhao', role: 0 as number, status: 1 as number },
    { username: 'henry', nickname: 'Henry Sun', role: 0 as number, status: 1 as number },
    { username: 'iris', nickname: 'Iris Yang', role: 0 as number, status: 1 as number },
    { username: 'jack', nickname: 'Jack Huang', role: 0 as number, status: 0 as number },
    { username: 'kate', nickname: 'Kate Xu', role: 0 as number, status: 1 as number },
    { username: 'leo', nickname: 'Leo Zhou', role: 0 as number, status: 1 as number },
    { username: 'mia', nickname: 'Mia Deng', role: 0 as number, status: 1 as number },
    { username: 'nick', nickname: 'Nick Feng', role: 0 as number, status: 0 as number },
    { username: 'olivia', nickname: 'Olivia Guo', role: 0 as number, status: 1 as number },
    { username: 'peter', nickname: 'Peter Hu', role: 0 as number, status: 1 as number },
    { username: 'queen', nickname: 'Queen Jin', role: 0 as number, status: 1 as number },
    { username: 'ryan', nickname: 'Ryan Jiang', role: 1 as number, status: 1 as number },
    { username: 'sara', nickname: 'Sara Pan', role: 0 as number, status: 0 as number },
    { username: 'tom', nickname: 'Tom He', role: 0 as number, status: 1 as number },
    { username: 'uma', nickname: 'Uma Qin', role: 0 as number, status: 1 as number },
    { username: 'vince', nickname: 'Vince Long', role: 0 as number, status: 1 as number },
    { username: 'wendy', nickname: 'Wendy Duan', role: 0 as number, status: 0 as number },
    { username: 'xavier', nickname: 'Xavier Wan', role: 0 as number, status: 1 as number },
    { username: 'yuki', nickname: 'Yuki Tan', role: 0 as number, status: 1 as number },
    { username: 'zack', nickname: 'Zack Bai', role: 0 as number, status: 1 as number },
    { username: 'devops', nickname: 'DevOps Bot', role: 0 as number, status: 0 as number },
    { username: 'tester1', nickname: '测试一号', role: 0 as number, status: 1 as number },
    { username: 'tester2', nickname: '测试二号', role: 0 as number, status: 1 as number },
    { username: 'aiengineer', nickname: 'AI Engineer', role: 0 as number, status: 1 as number },
    { username: 'mlops', nickname: 'ML Ops', role: 0 as number, status: 1 as number },
    { username: 'datasci', nickname: '数据科学家', role: 0 as number, status: 0 as number },
    { username: 'backend', nickname: '后端工程师', role: 0 as number, status: 1 as number },
    { username: 'frontend', nickname: '前端工程师', role: 0 as number, status: 1 as number },
    { username: 'pm_wang', nickname: 'Wang PM', role: 0 as number, status: 1 as number },
    { username: 'designer', nickname: '设计师', role: 0 as number, status: 1 as number },
    { username: 'intern', nickname: '实习生', role: 0 as number, status: 0 as number },
    { username: 'guest1', nickname: '访客A', role: 0 as number, status: 1 as number },
    { username: 'guest2', nickname: '访客B', role: 0 as number, status: 1 as number },
    { username: 'auditor', nickname: '审计员', role: 1 as number, status: 1 as number },
    { username: 'sysop', nickname: '系统运维', role: 1 as number, status: 1 as number },
  ]

  const avatarColors = ['4A90D9', 'E07040', '50B86C', 'C050C0', 'D9A040', '40B0C0']
  return names.map((n, i) => ({
    userId: `u-${String(i).padStart(3, '0')}`,
    username: n.username,
    nickname: n.nickname,
    email: `${n.username}@goraven.local`,
    avatar:
      i % 3 === 0
        ? `https://api.dicebear.com/9.x/initials/svg?seed=${n.username}&backgroundColor=${avatarColors[i % avatarColors.length]}`
        : '',
    role: n.role,
    status: n.status,
    dailyTokenLimit: i % 5 === 0 ? 10 : 0,
    sessionCount: n.role === 1 ? Math.floor(Math.random() * 50) + 5 : Math.floor(Math.random() * 30),
    lastActiveTime:
      i % 7 === 0
        ? ''
        : new Date(Date.now() - Math.random() * 30 * 24 * 3600 * 1000).toISOString(),
    created: new Date(2025, 0, 1 + i * 3).toISOString(),
  }))
}

/* ============================================
   Mutable state — shared across all handlers
   ============================================ */

let users: AdminUserItem[] = generateMockUsers()

/* ============================================
   Request types
   ============================================ */

export interface GetUsersParams {
  search?: string
  role?: string
  page: number
  pageSize: number
}

export interface CreateUserRequest {
  username: string
  password: string
  nickname: string
  role: number
}

export interface UpdateUserRequest {
  nickname: string
  email: string
  role: number
  status: number
  dailyTokenLimit: number
}

/* ============================================
   Helpers
   ============================================ */

function paginate<T>(list: T[], page: number, pageSize: number): PaginatedResponse<T> {
  const totalCount = list.length
  const totalPage = Math.max(1, Math.ceil(totalCount / pageSize))
  const safePage = Math.min(page, totalPage)
  const start = (safePage - 1) * pageSize
  const paged = list.slice(start, start + pageSize)

  return {
    list: paged,
    totalPage,
    totalCount,
    page: safePage,
    pageSize,
  }
}

/* ============================================
   Handlers
   ============================================ */

/** Reset to initial mock data (useful for dev reload) */
export function resetUsers(): void {
  users = generateMockUsers()
}

/** List users with optional search, role filter, and pagination */
export async function getUsers(params: GetUsersParams): Promise<PaginatedResponse<AdminUserItem>> {
  await listDelay()

  let filtered = users

  if (params.search?.trim()) {
    const q = params.search.trim().toLowerCase()
    filtered = filtered.filter((u) => u.username.toLowerCase().includes(q))
  }

  if (params.role && params.role !== 'all') {
    filtered = filtered.filter((u) => u.role === Number(params.role))
  }

  return paginate(filtered, params.page, params.pageSize)
}

/** Get single user detail */
export async function getUserDetail(userId: string): Promise<AdminUserItem> {
  await itemDelay()

  const user = users.find((u) => u.userId === userId)
  if (!user) {
    throw new Error(`User not found: ${userId}`)
  }

  return { ...user }
}

/** Create a new user */
export async function createUser(req: CreateUserRequest): Promise<AdminUserItem> {
  await mutationDelay()

  const newUser: AdminUserItem = {
    userId: `u-${String(users.length).padStart(3, '0')}`,
    username: req.username,
    nickname: req.nickname,
    email: `${req.username}@goraven.local`,
    avatar: '',
    role: req.role,
    status: 1,
    dailyTokenLimit: 0,
    sessionCount: 0,
    lastActiveTime: '',
    created: new Date().toISOString(),
  }

  users = [newUser, ...users]
  return { ...newUser }
}

/** Update an existing user */
export async function updateUser(userId: string, req: UpdateUserRequest): Promise<AdminUserItem> {
  await mutationDelay()

  const idx = users.findIndex((u) => u.userId === userId)
  if (idx === -1) {
    throw new Error(`User not found: ${userId}`)
  }

  users = users.map((u) =>
    u.userId === userId
      ? { ...u, nickname: req.nickname, email: req.email, role: req.role, status: req.status, dailyTokenLimit: req.dailyTokenLimit }
      : u,
  )

  return { ...users[idx] }
}

/** Reset a user's password (mock — does not actually store the password) */
export async function resetPassword(_userId: string, _password: string): Promise<void> {
  await mutationDelay()
  // no-op: mock does not persist passwords
}

/** Delete a user */
export async function deleteUser(userId: string): Promise<void> {
  await mutationDelay()

  const idx = users.findIndex((u) => u.userId === userId)
  if (idx === -1) {
    throw new Error(`User not found: ${userId}`)
  }

  users = users.filter((u) => u.userId !== userId)
}

/** Batch get users by ID */
export async function batchGetUsers(userIds: string[]): Promise<AdminUserItem[]> {
  await listDelay()

  const idSet = new Set(userIds)
  return users.filter((u) => idSet.has(u.userId)).map((u) => ({ ...u }))
}
