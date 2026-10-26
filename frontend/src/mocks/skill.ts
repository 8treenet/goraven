/* ============================================
   Mock data — Skills (non-admin)
   My skills, market, categories
   ============================================ */

import type {
  SkillCategory,
  MarketSkill,
  MarketSkillDetail,
  UserSkill,
  UserSkillDetail,
  SimpleSkill,
  PaginatedResponse,
} from './types'
import { listDelay, itemDelay, mutationDelay, heavyDelay } from './delay'

/* ---------- Helper ---------- */

export function generateSkillMd(name: string, description: string): string {
  return `---
name: ${name}
description: ${description}
version: 1.0.0
tools:
  - name: main
    description: Main function
    parameters:
      type: object
      properties:
        input:
          type: string
          description: Input parameter
---

# ${name}

${description}

## Usage

This skill provides automated operations.

## Configuration

No additional configuration required.
`
}

/* ---------- Categories ---------- */

const MOCK_SKILL_CATEGORIES: SkillCategory[] = [
  { categoryId: 1, name: '开发工具', icon: 'code' },
  { categoryId: 2, name: '数据处理', icon: 'search' },
  { categoryId: 3, name: '文档工具', icon: 'file-text' },
  { categoryId: 4, name: '运维工具', icon: 'terminal' },
]

/* ---------- Installed skills ---------- */

let MOCK_INSTALLED_SKILLS: UserSkill[] = [
  { userSkillId: 1, skillName: 'code-review', description: '代码审查工具', icon: 'code', marketSkillId: 1, categoryId: 1, categoryName: '开发工具', source: 'market', installStatus: 2, created: '2025-03-01T10:00:00Z', updated: '2025-06-01T08:30:00Z' },
  { userSkillId: 2, skillName: 'pdf-reader', description: 'PDF 文档解析工具', icon: 'file-text', marketSkillId: 2, categoryId: 3, categoryName: '文档工具', source: 'market', installStatus: 2, created: '2025-03-02T10:00:00Z', updated: '2025-05-28T08:30:00Z' },
  { userSkillId: 3, skillName: 'excel-analyzer', description: 'Excel 数据处理工具', icon: 'file-text', marketSkillId: 3, categoryId: 2, categoryName: '数据处理', source: 'market', installStatus: 2, created: '2025-03-03T10:00:00Z', updated: '2025-06-01T08:30:00Z' },
  { userSkillId: 4, skillName: 'api-doc-gen', description: 'API 文档自动生成', icon: 'book', marketSkillId: 4, categoryId: 3, categoryName: '文档工具', source: 'market', installStatus: 1, created: '2025-06-01T10:00:00Z', updated: '2025-06-01T10:00:00Z' },
  { userSkillId: 5, skillName: 'db-migration', description: '数据库迁移助手', icon: 'database', marketSkillId: 5, categoryId: 1, categoryName: '开发工具', source: 'market', installStatus: 3, installError: 'pip dependency conflict', created: '2025-05-01T10:00:00Z', updated: '2025-05-15T08:30:00Z' },
  { userSkillId: 6, skillName: 'unit-test-gen', description: '单元测试生成工具', icon: 'wrench', marketSkillId: 6, categoryId: 1, categoryName: '开发工具', source: 'market', installStatus: 2, created: '2025-04-15T10:00:00Z', updated: '2025-05-20T08:30:00Z' },
]

/* ---------- Market skills ---------- */

const MOCK_MARKET_SKILLS: MarketSkill[] = [
  { skillId: 1, name: 'code-review', description: '代码审查工具 — 自动分析代码问题和最佳实践', icon: 'code', source: 'clawhub', categoryId: 1, categoryName: '开发工具', installedCount: 2340, userInstalled: true, updated: '2025-06-01T10:00:00Z' },
  { skillId: 2, name: 'pdf-reader', description: 'PDF 文档解析 — 提取文字、图片和表格', icon: 'file-text', source: 'clawhub', categoryId: 3, categoryName: '文档工具', installedCount: 1890, userInstalled: true, updated: '2025-05-28T10:00:00Z' },
  { skillId: 3, name: 'excel-analyzer', description: 'Excel 数据处理 — 统计分析和可视化', icon: 'file-text', source: 'clawhub', categoryId: 2, categoryName: '数据处理', installedCount: 1560, userInstalled: true, updated: '2025-06-01T10:00:00Z' },
  { skillId: 4, name: 'api-doc-gen', description: 'API 文档自动生成 — 从代码生成 OpenAPI 规范', icon: 'book', source: 'clawhub', categoryId: 3, categoryName: '文档工具', installedCount: 980, userInstalled: true, updated: '2025-05-25T10:00:00Z' },
  { skillId: 5, name: 'db-migration', description: '数据库迁移助手 — 生成和管理迁移脚本', icon: 'database', source: 'clawhub', categoryId: 1, categoryName: '开发工具', installedCount: 750, userInstalled: true, updated: '2025-05-20T10:00:00Z' },
  { skillId: 6, name: 'unit-test-gen', description: '单元测试生成 — 自动生成测试用例和 mock', icon: 'wrench', source: 'clawhub', categoryId: 1, categoryName: '开发工具', installedCount: 1340, userInstalled: true, updated: '2025-06-01T10:00:00Z' },
  { skillId: 7, name: 'docker-deploy', description: 'Docker 部署助手 — 自动生成 Dockerfile 和 compose 文件', icon: 'terminal', source: 'clawhub', categoryId: 1, categoryName: '开发工具', installedCount: 890, userInstalled: false, updated: '2025-05-15T10:00:00Z' },
  { skillId: 8, name: 'log-analyzer', description: '日志分析工具 — 解析和可视化各类日志', icon: 'terminal', source: 'clawhub', categoryId: 4, categoryName: '运维工具', installedCount: 620, userInstalled: false, updated: '2025-05-10T10:00:00Z' },
]

/* ---------- Async functions ---------- */

export async function getSimpleSkills(): Promise<SimpleSkill[]> {
  await listDelay()
  return MOCK_INSTALLED_SKILLS.filter((s) => s.installStatus === 2).map((s) => ({
    userSkillId: s.userSkillId,
    skillName: s.skillName,
    description: s.description,
    icon: s.icon,
    source: s.source,
    categoryId: s.categoryId,
    categoryName: s.categoryName,
  }))
}

export async function getMarketSkills(params?: {
  search?: string
  categoryId?: number
  source?: string
  page?: number
  pageSize?: number
}): Promise<PaginatedResponse<MarketSkill>> {
  await listDelay()
  let list = [...MOCK_MARKET_SKILLS]
  if (params?.search) {
    const q = params.search.toLowerCase()
    list = list.filter((s) => s.name.toLowerCase().includes(q) || s.description.toLowerCase().includes(q))
  }
  if (params?.categoryId) list = list.filter((s) => s.categoryId === params.categoryId)
  const page = params?.page ?? 1
  const pageSize = params?.pageSize ?? 10
  const totalCount = list.length
  const totalPage = Math.ceil(totalCount / pageSize)
  const start = (page - 1) * pageSize
  return { list: list.slice(start, start + pageSize), totalPage, totalCount, page, pageSize }
}

export async function getMarketSkillDetail(id: number): Promise<MarketSkillDetail> {
  await itemDelay()
  const skill = MOCK_MARKET_SKILLS.find((s) => s.skillId === id)
  if (!skill) throw new Error(`Skill ${id} not found`)
  return { ...skill, content: generateSkillMd(skill.name, skill.description) }
}

export async function getUserSkills(params?: {
  search?: string
  categoryId?: number
  source?: string
  status?: number
  page?: number
  pageSize?: number
}): Promise<PaginatedResponse<UserSkill>> {
  await listDelay()
  let list = [...MOCK_INSTALLED_SKILLS]
  if (params?.search) {
    const q = params.search.toLowerCase()
    list = list.filter((s) => s.skillName.toLowerCase().includes(q))
  }
  if (params?.categoryId) list = list.filter((s) => s.categoryId === params.categoryId)
  if (params?.status !== undefined) list = list.filter((s) => s.installStatus === params.status)
  const page = params?.page ?? 1
  const pageSize = params?.pageSize ?? 10
  const totalCount = list.length
  const totalPage = Math.ceil(totalCount / pageSize)
  const start = (page - 1) * pageSize
  return { list: list.slice(start, start + pageSize), totalPage, totalCount, page, pageSize }
}

export async function getUserSkillDetail(id: number): Promise<UserSkillDetail> {
  await itemDelay()
  const skill = MOCK_INSTALLED_SKILLS.find((s) => s.userSkillId === id)
  if (!skill) throw new Error(`UserSkill ${id} not found`)
  return { ...skill, content: generateSkillMd(skill.skillName, skill.description) }
}

export async function updateUserSkill(id: number, req: { icon?: string; categoryId?: number }): Promise<void> {
  await mutationDelay()
  const skill = MOCK_INSTALLED_SKILLS.find((s) => s.userSkillId === id)
  if (!skill) throw new Error(`UserSkill ${id} not found`)
  if (req.icon !== undefined) skill.icon = req.icon
  if (req.categoryId !== undefined) {
    skill.categoryId = req.categoryId
    const cat = MOCK_SKILL_CATEGORIES.find((c) => c.categoryId === req.categoryId)
    if (cat) skill.categoryName = cat.name
  }
}

export async function deleteUserSkill(id: number): Promise<void> {
  await mutationDelay()
  MOCK_INSTALLED_SKILLS = MOCK_INSTALLED_SKILLS.filter((s) => s.userSkillId !== id)
}

export async function refreshSkills(): Promise<{ added: number; removed: number }> {
  await heavyDelay()
  return { added: Math.random() > 0.5 ? 1 : 0, removed: 0 }
}

export async function installSkill(skillId: number): Promise<{ userSkillId: number }> {
  await heavyDelay()
  const marketSkill = MOCK_MARKET_SKILLS.find((s) => s.skillId === skillId)
  if (!marketSkill) throw new Error(`MarketSkill ${skillId} not found`)
  const newId = MOCK_INSTALLED_SKILLS.length + 1
  MOCK_INSTALLED_SKILLS.push({
    userSkillId: newId,
    skillName: marketSkill.name,
    description: marketSkill.description,
    icon: marketSkill.icon,
    marketSkillId: skillId,
    categoryId: marketSkill.categoryId,
    categoryName: marketSkill.categoryName,
    source: 'market',
    installStatus: 2,
    created: new Date().toISOString(),
    updated: new Date().toISOString(),
  })
  marketSkill.userInstalled = true
  marketSkill.installedCount++
  return { userSkillId: newId }
}

export async function retryInstall(userSkillId: number): Promise<void> {
  await heavyDelay()
  const skill = MOCK_INSTALLED_SKILLS.find((s) => s.userSkillId === userSkillId)
  if (!skill) throw new Error(`UserSkill ${userSkillId} not found`)
  skill.installStatus = 2
  skill.installError = undefined
}

export async function getSkillCategories(): Promise<SkillCategory[]> {
  await listDelay()
  return [...MOCK_SKILL_CATEGORIES]
}
