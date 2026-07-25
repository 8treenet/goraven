/* ============================================
   Mock data — Admin Skills
   Extracted from features/admin/skills/AdminSkillsPage.tsx
   ============================================ */

import type {
  SystemSkillItem,
  SystemSkillDetail,
  AdminMarketSkillItem,
  ClawHubSkillItem,
  AdminSkillCategoryItem,
  PaginatedResponse,
} from '../types'
import { listDelay, itemDelay, mutationDelay, heavyDelay } from '../delay'

/* ============================================
   Local types
   ============================================ */

type InstallStatus = 1 | 2 | 3
type SourceType = 'clawhub' | 'custom_upload'
type ClawHubSort = 'newest' | 'updated' | 'downloads' | 'stars' | 'installs' | 'trending'

interface InstalledSkillRecord {
  recordId: number
  userId: string
  skillId: number
  skillName: string
  categoryName: string
  categoryIcon: string
  source: SourceType
  installStatus: InstallStatus
  reason: string
  created: string
}

interface MarketSkillFull {
  marketSkillId: number
  name: string
  description: string
  icon: string
  source: SourceType
  sourceUrl: string
  categoryId: number
  categoryName: string
  categoryIcon: string
  installedCount: number
  status: number
  sortOrder: number
  remark: string
  content: string
  created: string
  updated: string
}

interface SystemSkillFull {
  systemSkillId: number
  name: string
  description: string
  icon: string
  content: string
  status: number
  created: string
  updated: string
}

interface ClawHubItemFull {
  slug: string
  displayName: string
  summary: string
  version: string
  score: number
  downloads: number
  installs: number
  stars: number
  updatedAt: string
  content: string
}

/* ============================================
   YAML template
   ============================================ */

export const GLOBAL_TEMPLATE = `---
name: goraven-
description:
---

`

/* ============================================
   In‑memory stores
   ============================================ */

let systemSkillStore: SystemSkillFull[] = []
let marketSkillStore: MarketSkillFull[] = []
let installedSkillStore: InstalledSkillRecord[] = []
let categoryStore: AdminSkillCategoryItem[] = []
let clawHubStore: ClawHubItemFull[] = []
let storesInitialized = false

function ensureStores() {
  if (!storesInitialized) {
    systemSkillStore = generateSystemSkillFulls()
    marketSkillStore = generateMarketSkillFulls()
    installedSkillStore = generateInstalledSkillRecords()
    categoryStore = generateSkillCategories()
    clawHubStore = generateClawHubItemFulls()
    storesInitialized = true
  }
}

/* ============================================
   ID helpers
   ============================================ */

function nextSystemSkillId(): number {
  const ids = systemSkillStore.map((s) => s.systemSkillId)
  return Math.max(10, ...ids) + 1
}

function nextMarketSkillId(): number {
  const ids = marketSkillStore.map((s) => s.marketSkillId)
  return Math.max(200, ...ids) + 1
}

function nextCategoryId(): number {
  const ids = categoryStore.map((c) => c.categoryId)
  return Math.max(20, ...ids) + 1
}

function toSystemSkillItem(full: SystemSkillFull): SystemSkillItem {
  return {
    systemSkillId: full.systemSkillId,
    name: full.name,
    description: full.description,
    icon: full.icon,
    status: full.status,
    created: full.created,
    updated: full.updated,
  }
}

function toAdminMarketSkillItem(full: MarketSkillFull): AdminMarketSkillItem {
  return {
    marketSkillId: full.marketSkillId,
    name: full.name,
    description: full.description,
    icon: full.icon,
    source: full.source,
    categoryId: full.categoryId,
    categoryName: full.categoryName,
    status: full.status,
    sortOrder: full.sortOrder,
    installedCount: full.installedCount,
    remark: full.remark,
    created: full.created,
    updated: full.updated,
  }
}

function toClawHubSkillItem(full: ClawHubItemFull): ClawHubSkillItem {
  return {
    slug: full.slug,
    name: full.displayName,
    description: full.summary,
    icon: 'zap',
    source: 'clawhub',
    score: full.score,
    downloads: full.downloads,
    installs: full.installs,
    stars: full.stars,
  }
}

/* ============================================
   Data generators
   ============================================ */

function generateSystemSkillFulls(): SystemSkillFull[] {
  return [
    {
      systemSkillId: 1,
      name: 'goraven-code-review',
      description: '审查代码变更，指出风险、边界条件和可维护性问题',
      icon: 'code',
      status: 0,
      created: '2026-05-30T10:30:00Z',
      updated: '2026-05-30T10:30:00Z',
      content: `---
name: goraven-code-review
description: 审查代码变更，指出风险、边界条件和可维护性问题
---

你是严谨的代码审查助手。优先指出真实风险，避免风格化建议。`,
    },
    {
      systemSkillId: 2,
      name: 'goraven-db-migration',
      description: '规划数据库迁移步骤，检查兼容性和回滚策略',
      icon: 'database',
      status: 0,
      created: '2026-05-28T15:12:00Z',
      updated: '2026-05-28T15:12:00Z',
      content: `---
name: goraven-db-migration
description: 规划数据库迁移步骤，检查兼容性和回滚策略
---

处理数据库迁移时，先说明影响范围，再给出执行和回滚步骤。`,
    },
    {
      systemSkillId: 3,
      name: 'goraven-finance-audit',
      description: '辅助财务数据核对，强调来源、口径和异常解释',
      icon: 'calculator',
      status: 1,
      created: '2026-05-21T09:48:00Z',
      updated: '2026-05-21T09:48:00Z',
      content: `---
name: goraven-finance-audit
description: 辅助财务数据核对，强调来源、口径和异常解释
---

所有财务结论必须说明数据口径，不能编造缺失字段。`,
    },
    {
      systemSkillId: 4,
      name: 'goraven-incident-response',
      description: '将故障信息整理为排查路径、影响评估和复盘条目',
      icon: 'alert-triangle',
      status: 0,
      created: '2026-05-18T18:20:00Z',
      updated: '2026-05-18T18:20:00Z',
      content: `---
name: goraven-incident-response
description: 将故障信息整理为排查路径、影响评估和复盘条目
---

面对故障时先稳定范围，再定位原因，不直接跳到修复建议。`,
    },
  ]
}

export function generateSystemSkills(): SystemSkillItem[] {
  return [
    {
      systemSkillId: 1,
      name: 'goraven-code-review',
      description: '审查代码变更，指出风险、边界条件和可维护性问题',
      icon: 'code',
      status: 0,
      created: '2026-05-30T10:30:00Z',
      updated: '2026-05-30T10:30:00Z',
    },
    {
      systemSkillId: 2,
      name: 'goraven-db-migration',
      description: '规划数据库迁移步骤，检查兼容性和回滚策略',
      icon: 'database',
      status: 0,
      created: '2026-05-28T15:12:00Z',
      updated: '2026-05-28T15:12:00Z',
    },
    {
      systemSkillId: 3,
      name: 'goraven-finance-audit',
      description: '辅助财务数据核对，强调来源、口径和异常解释',
      icon: 'calculator',
      status: 1,
      created: '2026-05-21T09:48:00Z',
      updated: '2026-05-21T09:48:00Z',
    },
    {
      systemSkillId: 4,
      name: 'goraven-incident-response',
      description: '将故障信息整理为排查路径、影响评估和复盘条目',
      icon: 'alert-triangle',
      status: 0,
      created: '2026-05-18T18:20:00Z',
      updated: '2026-05-18T18:20:00Z',
    },
  ]
}

export function generateSkillCategories(): AdminSkillCategoryItem[] {
  return [
    {
      categoryId: 1,
      name: '通用',
      icon: 'sparkles',
      isDefault: 1,
      created: '2026-05-30T09:00:00Z',
      updated: '2026-05-30T09:00:00Z',
    },
    {
      categoryId: 2,
      name: '编程开发',
      icon: 'code',
      isDefault: 0,
      created: '2026-05-29T11:00:00Z',
      updated: '2026-05-29T11:00:00Z',
    },
    {
      categoryId: 3,
      name: '数据与AI',
      icon: 'brain',
      isDefault: 0,
      created: '2026-05-28T11:00:00Z',
      updated: '2026-05-28T11:00:00Z',
    },
    {
      categoryId: 4,
      name: '运维部署',
      icon: 'terminal',
      isDefault: 0,
      created: '2026-05-27T11:00:00Z',
      updated: '2026-05-27T11:00:00Z',
    },
    {
      categoryId: 5,
      name: '商业效率',
      icon: 'briefcase',
      isDefault: 0,
      created: '2026-05-26T11:00:00Z',
      updated: '2026-05-26T11:00:00Z',
    },
  ]
}

function generateMarketSkillFulls(): MarketSkillFull[] {
  return [
    {
      marketSkillId: 101,
      name: 'pull-request-writer',
      description: '根据提交记录生成结构化 PR 描述和测试说明',
      icon: 'git-branch',
      source: 'clawhub',
      sourceUrl: 'https://clawhub.dev/skills/pull-request-writer',
      categoryId: 2,
      categoryName: '编程开发',
      categoryIcon: 'code',
      installedCount: 12,
      status: 1,
      sortOrder: 10,
      remark: '团队开发常用，保持上架',
      created: '2026-05-30T12:15:00Z',
      updated: '2026-05-30T12:15:00Z',
      content: `---
name: pull-request-writer
description: 根据提交记录生成结构化 PR 描述和测试说明
---

Write concise pull request summaries with risks and tests.`,
    },
    {
      marketSkillId: 102,
      name: 'sql-explain-assistant',
      description: '分析 SQL 执行计划，给出索引和查询改写建议',
      icon: 'database',
      source: 'clawhub',
      sourceUrl: 'https://clawhub.dev/skills/sql-explain-assistant',
      categoryId: 3,
      categoryName: '数据与AI',
      categoryIcon: 'brain',
      installedCount: 8,
      status: 1,
      sortOrder: 20,
      remark: '',
      created: '2026-05-27T08:30:00Z',
      updated: '2026-05-27T08:30:00Z',
      content: `---
name: sql-explain-assistant
description: 分析 SQL 执行计划，给出索引和查询改写建议
---

Inspect query plans carefully before proposing indexes.`,
    },
    {
      marketSkillId: 103,
      name: 'deploy-checklist',
      description: '生成发布前检查清单，覆盖配置、迁移和回滚',
      icon: 'rocket',
      source: 'custom_upload',
      sourceUrl: '',
      categoryId: 4,
      categoryName: '运维部署',
      categoryIcon: 'terminal',
      installedCount: 5,
      status: 0,
      sortOrder: 30,
      remark: '等待新版压缩包验证',
      created: '2026-05-22T16:40:00Z',
      updated: '2026-05-22T16:40:00Z',
      content: `---
name: deploy-checklist
description: 生成发布前检查清单，覆盖配置、迁移和回滚
---

Produce deployment checklists with clear ownership.`,
    },
    {
      marketSkillId: 104,
      name: 'meeting-brief',
      description: '把会议材料压缩为议题、决策点和后续动作',
      icon: 'notebook',
      source: 'custom_upload',
      sourceUrl: '',
      categoryId: 5,
      categoryName: '商业效率',
      categoryIcon: 'briefcase',
      installedCount: 0,
      status: 1,
      sortOrder: 40,
      remark: '',
      created: '2026-05-20T10:20:00Z',
      updated: '2026-05-20T10:20:00Z',
      content: `---
name: meeting-brief
description: 把会议材料压缩为议题、决策点和后续动作
---

Summarize meetings into decisions and action items.`,
    },
    {
      marketSkillId: 105,
      name: 'risk-ledger',
      description: '维护项目风险台账，按概率、影响和责任人归类',
      icon: 'shield',
      source: 'clawhub',
      sourceUrl: 'https://clawhub.dev/skills/risk-ledger',
      categoryId: 1,
      categoryName: '通用',
      categoryIcon: 'sparkles',
      installedCount: 3,
      status: 1,
      sortOrder: 50,
      remark: '',
      created: '2026-05-18T14:10:00Z',
      updated: '2026-05-18T14:10:00Z',
      content: `---
name: risk-ledger
description: 维护项目风险台账，按概率、影响和责任人归类
---

Track risk with owner, probability, impact, mitigation.`,
    },
  ]
}

export function generateMarketSkills(): AdminMarketSkillItem[] {
  return [
    {
      marketSkillId: 101,
      name: 'pull-request-writer',
      description: '根据提交记录生成结构化 PR 描述和测试说明',
      icon: 'git-branch',
      source: 'clawhub',
      categoryId: 2,
      categoryName: '编程开发',
      status: 1,
      sortOrder: 10,
      installedCount: 12,
      remark: '团队开发常用，保持上架',
      created: '2026-05-30T12:15:00Z',
      updated: '2026-05-30T12:15:00Z',
    },
    {
      marketSkillId: 102,
      name: 'sql-explain-assistant',
      description: '分析 SQL 执行计划，给出索引和查询改写建议',
      icon: 'database',
      source: 'clawhub',
      categoryId: 3,
      categoryName: '数据与AI',
      status: 1,
      sortOrder: 20,
      installedCount: 8,
      remark: '',
      created: '2026-05-27T08:30:00Z',
      updated: '2026-05-27T08:30:00Z',
    },
    {
      marketSkillId: 103,
      name: 'deploy-checklist',
      description: '生成发布前检查清单，覆盖配置、迁移和回滚',
      icon: 'rocket',
      source: 'custom_upload',
      categoryId: 4,
      categoryName: '运维部署',
      status: 0,
      sortOrder: 30,
      installedCount: 5,
      remark: '等待新版压缩包验证',
      created: '2026-05-22T16:40:00Z',
      updated: '2026-05-22T16:40:00Z',
    },
    {
      marketSkillId: 104,
      name: 'meeting-brief',
      description: '把会议材料压缩为议题、决策点和后续动作',
      icon: 'notebook',
      source: 'custom_upload',
      categoryId: 5,
      categoryName: '商业效率',
      status: 1,
      sortOrder: 40,
      installedCount: 0,
      remark: '',
      created: '2026-05-20T10:20:00Z',
      updated: '2026-05-20T10:20:00Z',
    },
    {
      marketSkillId: 105,
      name: 'risk-ledger',
      description: '维护项目风险台账，按概率、影响和责任人归类',
      icon: 'shield',
      source: 'clawhub',
      categoryId: 1,
      categoryName: '通用',
      status: 1,
      sortOrder: 50,
      installedCount: 3,
      remark: '',
      created: '2026-05-18T14:10:00Z',
      updated: '2026-05-18T14:10:00Z',
    },
  ]
}

function generateInstalledSkillRecords(): InstalledSkillRecord[] {
  return [
    {
      recordId: 1,
      userId: 'usr_8f3a91d21c',
      skillId: 101,
      skillName: 'pull-request-writer',
      categoryName: '编程开发',
      categoryIcon: 'code',
      source: 'clawhub',
      installStatus: 2,
      reason: '',
      created: '2026-05-31T09:12:00Z',
    },
    {
      recordId: 2,
      userId: 'usr_19cba5e87a',
      skillId: 102,
      skillName: 'sql-explain-assistant',
      categoryName: '数据与AI',
      categoryIcon: 'brain',
      source: 'clawhub',
      installStatus: 2,
      reason: '',
      created: '2026-05-30T13:24:00Z',
    },
    {
      recordId: 3,
      userId: 'usr_a73e4c50ff',
      skillId: 103,
      skillName: 'deploy-checklist',
      categoryName: '运维部署',
      categoryIcon: 'terminal',
      source: 'custom_upload',
      installStatus: 3,
      reason: '安装脚本退出码 1',
      created: '2026-05-29T17:42:00Z',
    },
    {
      recordId: 4,
      userId: 'usr_8f3a91d21c',
      skillId: 104,
      skillName: 'meeting-brief',
      categoryName: '商业效率',
      categoryIcon: 'briefcase',
      source: 'custom_upload',
      installStatus: 1,
      reason: '',
      created: '2026-05-29T10:18:00Z',
    },
    {
      recordId: 5,
      userId: 'usr_044b72aa19',
      skillId: 101,
      skillName: 'pull-request-writer',
      categoryName: '编程开发',
      categoryIcon: 'code',
      source: 'clawhub',
      installStatus: 2,
      reason: '',
      created: '2026-05-28T08:36:00Z',
    },
    {
      recordId: 6,
      userId: 'usr_9025cf7d33',
      skillId: 105,
      skillName: 'risk-ledger',
      categoryName: '通用',
      categoryIcon: 'sparkles',
      source: 'clawhub',
      installStatus: 2,
      reason: '',
      created: '2026-05-27T11:06:00Z',
    },
  ]
}

export function generateInstalledSkills(): InstalledSkillRecord[] {
  return [
    {
      recordId: 1,
      userId: 'usr_8f3a91d21c',
      skillId: 101,
      skillName: 'pull-request-writer',
      categoryName: '编程开发',
      categoryIcon: 'code',
      source: 'clawhub',
      installStatus: 2,
      reason: '',
      created: '2026-05-31T09:12:00Z',
    },
    {
      recordId: 2,
      userId: 'usr_19cba5e87a',
      skillId: 102,
      skillName: 'sql-explain-assistant',
      categoryName: '数据与AI',
      categoryIcon: 'brain',
      source: 'clawhub',
      installStatus: 2,
      reason: '',
      created: '2026-05-30T13:24:00Z',
    },
    {
      recordId: 3,
      userId: 'usr_a73e4c50ff',
      skillId: 103,
      skillName: 'deploy-checklist',
      categoryName: '运维部署',
      categoryIcon: 'terminal',
      source: 'custom_upload',
      installStatus: 3,
      reason: '安装脚本退出码 1',
      created: '2026-05-29T17:42:00Z',
    },
    {
      recordId: 4,
      userId: 'usr_8f3a91d21c',
      skillId: 104,
      skillName: 'meeting-brief',
      categoryName: '商业效率',
      categoryIcon: 'briefcase',
      source: 'custom_upload',
      installStatus: 1,
      reason: '',
      created: '2026-05-29T10:18:00Z',
    },
    {
      recordId: 5,
      userId: 'usr_044b72aa19',
      skillId: 101,
      skillName: 'pull-request-writer',
      categoryName: '编程开发',
      categoryIcon: 'code',
      source: 'clawhub',
      installStatus: 2,
      reason: '',
      created: '2026-05-28T08:36:00Z',
    },
    {
      recordId: 6,
      userId: 'usr_9025cf7d33',
      skillId: 105,
      skillName: 'risk-ledger',
      categoryName: '通用',
      categoryIcon: 'sparkles',
      source: 'clawhub',
      installStatus: 2,
      reason: '',
      created: '2026-05-27T11:06:00Z',
    },
  ]
}

function generateClawHubItemFulls(): ClawHubItemFull[] {
  return [
    {
      slug: 'api-contract-auditor',
      displayName: 'API Contract Auditor',
      summary: '检查接口文档、请求示例和返回字段是否一致',
      version: '1.3.0',
      score: 0.96,
      downloads: 1284,
      installs: 742,
      stars: 91,
      updatedAt: '2026-05-31T12:00:00Z',
      content: `---
name: api-contract-auditor
description: 检查接口文档、请求示例和返回字段是否一致
---

Review API contracts for consistency and missing edge cases.`,
    },
    {
      slug: 'incident-postmortem',
      displayName: 'Incident Postmortem',
      summary: '把故障时间线整理为复盘报告和改进项',
      version: '2.0.1',
      score: 0.91,
      downloads: 986,
      installs: 668,
      stars: 76,
      updatedAt: '2026-05-29T09:00:00Z',
      content: `---
name: incident-postmortem
description: 把故障时间线整理为复盘报告和改进项
---

Create postmortems from incident timelines.`,
    },
    {
      slug: 'front-end-a11y-check',
      displayName: 'Frontend A11y Check',
      summary: '审查前端界面的键盘访问、语义和对比度问题',
      version: '0.8.5',
      score: 0.88,
      downloads: 612,
      installs: 331,
      stars: 54,
      updatedAt: '2026-05-26T14:20:00Z',
      content: `---
name: front-end-a11y-check
description: 审查前端界面的键盘访问、语义和对比度问题
---

Check UI changes against accessibility basics.`,
    },
    {
      slug: 'release-note-builder',
      displayName: 'Release Note Builder',
      summary: '根据 commits 和 issue 列表生成发布说明',
      version: '1.0.2',
      score: 0.86,
      downloads: 420,
      installs: 218,
      stars: 32,
      updatedAt: '2026-06-01T16:30:00Z',
      content: `---
name: release-note-builder
description: 根据 commits 和 issue 列表生成发布说明
---

Build release notes from commits, issues, and breaking changes.`,
    },
    {
      slug: 'log-pattern-miner',
      displayName: 'Log Pattern Miner',
      summary: '从日志片段中提取重复模式、异常峰值和排查线索',
      version: '0.9.1',
      score: 0.94,
      downloads: 5340,
      installs: 1180,
      stars: 143,
      updatedAt: '2026-05-18T08:00:00Z',
      content: `---
name: log-pattern-miner
description: 从日志片段中提取重复模式、异常峰值和排查线索
---

Find recurring log patterns and incident clues.`,
    },
    {
      slug: 'spreadsheet-clean-room',
      displayName: 'Spreadsheet Clean Room',
      summary: '清洗表格字段，整理缺失值、异常值和口径说明',
      version: '1.6.4',
      score: 0.82,
      downloads: 2104,
      installs: 1940,
      stars: 61,
      updatedAt: '2026-04-30T10:15:00Z',
      content: `---
name: spreadsheet-clean-room
description: 清洗表格字段，整理缺失值、异常值和口径说明
---

Clean spreadsheet data with explicit assumptions.`,
    },
    {
      slug: 'diagram-to-task-list',
      displayName: 'Diagram To Task List',
      summary: '把架构图或流程图描述拆成可执行任务列表',
      version: '0.4.0',
      score: 0.79,
      downloads: 388,
      installs: 116,
      stars: 118,
      updatedAt: '2026-05-24T19:45:00Z',
      content: `---
name: diagram-to-task-list
description: 把架构图或流程图描述拆成可执行任务列表
---

Convert diagrams into ordered implementation tasks.`,
    },
    {
      slug: 'prompt-regression-tester',
      displayName: 'Prompt Regression Tester',
      summary: '为提示词改动生成回归样例和期望输出检查点',
      version: '0.2.0',
      score: 0.93,
      downloads: 240,
      installs: 82,
      stars: 27,
      updatedAt: '2026-06-02T08:10:00Z',
      content: `---
name: prompt-regression-tester
description: 为提示词改动生成回归样例和期望输出检查点
---

Create regression cases for prompt changes.`,
    },
  ]
}

export function generateClawHubItems(): ClawHubSkillItem[] {
  return [
    {
      slug: 'api-contract-auditor',
      name: 'API Contract Auditor',
      description: '检查接口文档、请求示例和返回字段是否一致',
      icon: 'zap',
      source: 'clawhub',
      score: 0.96,
      downloads: 1284,
      installs: 742,
      stars: 91,
    },
    {
      slug: 'incident-postmortem',
      name: 'Incident Postmortem',
      description: '把故障时间线整理为复盘报告和改进项',
      icon: 'zap',
      source: 'clawhub',
      score: 0.91,
      downloads: 986,
      installs: 668,
      stars: 76,
    },
    {
      slug: 'front-end-a11y-check',
      name: 'Frontend A11y Check',
      description: '审查前端界面的键盘访问、语义和对比度问题',
      icon: 'zap',
      source: 'clawhub',
      score: 0.88,
      downloads: 612,
      installs: 331,
      stars: 54,
    },
    {
      slug: 'release-note-builder',
      name: 'Release Note Builder',
      description: '根据 commits 和 issue 列表生成发布说明',
      icon: 'zap',
      source: 'clawhub',
      score: 0.86,
      downloads: 420,
      installs: 218,
      stars: 32,
    },
    {
      slug: 'log-pattern-miner',
      name: 'Log Pattern Miner',
      description: '从日志片段中提取重复模式、异常峰值和排查线索',
      icon: 'zap',
      source: 'clawhub',
      score: 0.94,
      downloads: 5340,
      installs: 1180,
      stars: 143,
    },
    {
      slug: 'spreadsheet-clean-room',
      name: 'Spreadsheet Clean Room',
      description: '清洗表格字段，整理缺失值、异常值和口径说明',
      icon: 'zap',
      source: 'clawhub',
      score: 0.82,
      downloads: 2104,
      installs: 1940,
      stars: 61,
    },
    {
      slug: 'diagram-to-task-list',
      name: 'Diagram To Task List',
      description: '把架构图或流程图描述拆成可执行任务列表',
      icon: 'zap',
      source: 'clawhub',
      score: 0.79,
      downloads: 388,
      installs: 116,
      stars: 118,
    },
    {
      slug: 'prompt-regression-tester',
      name: 'Prompt Regression Tester',
      description: '为提示词改动生成回归样例和期望输出检查点',
      icon: 'zap',
      source: 'clawhub',
      score: 0.93,
      downloads: 240,
      installs: 82,
      stars: 27,
    },
  ]
}

/* ============================================
   Async mock functions — System skills
   ============================================ */

export async function getSystemSkills(params?: {
  page?: number
  pageSize?: number
  search?: string
  status?: string
}): Promise<PaginatedResponse<SystemSkillItem>> {
  await listDelay()
  ensureStores()

  let filtered = systemSkillStore.map(toSystemSkillItem)

  if (params?.search) {
    const q = params.search.toLowerCase()
    filtered = filtered.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        s.description.toLowerCase().includes(q),
    )
  }

  if (params?.status && params.status !== 'all') {
    const statusNum = Number(params.status)
    filtered = filtered.filter((s) => s.status === statusNum)
  }

  const page = params?.page ?? 1
  const pageSize = params?.pageSize ?? 20
  const totalCount = filtered.length
  const totalPage = Math.max(1, Math.ceil(totalCount / pageSize))
  const start = (page - 1) * pageSize
  const list = filtered.slice(start, start + pageSize)

  return { list, totalPage, totalCount, page, pageSize }
}

export async function getSystemSkillDetail(
  id: number,
): Promise<SystemSkillDetail> {
  await itemDelay()
  ensureStores()

  const full = systemSkillStore.find((s) => s.systemSkillId === id)
  if (!full) {
    throw new Error('全局技能不存在')
  }

  return {
    ...toSystemSkillItem(full),
    content: full.content,
  }
}

export async function createSystemSkill(content: string): Promise<SystemSkillDetail> {
  await mutationDelay()
  ensureStores()

  const parsed = parseSkillContent(content)
  if (parsed.errors.length > 0) {
    throw new Error(parsed.errors[0])
  }

  const now = new Date().toISOString()
  const full: SystemSkillFull = {
    systemSkillId: nextSystemSkillId(),
    name: parsed.name,
    description: parsed.description,
    icon: 'bot',
    content,
    status: 0,
    created: now,
    updated: now,
  }
  systemSkillStore.unshift(full)

  return { ...toSystemSkillItem(full), content }
}

export async function updateSystemSkill(
  id: number,
  content: string,
): Promise<SystemSkillDetail> {
  await mutationDelay()
  ensureStores()

  const parsed = parseSkillContent(content)
  if (parsed.errors.length > 0) {
    throw new Error(parsed.errors[0])
  }

  const index = systemSkillStore.findIndex((s) => s.systemSkillId === id)
  if (index === -1) {
    throw new Error('全局技能不存在')
  }

  systemSkillStore[index] = {
    ...systemSkillStore[index],
    name: parsed.name,
    description: parsed.description,
    content,
    updated: new Date().toISOString(),
  }

  return {
    ...toSystemSkillItem(systemSkillStore[index]),
    content,
  }
}

export async function toggleSystemSkillStatus(
  id: number,
  status: number,
): Promise<SystemSkillItem> {
  await mutationDelay()
  ensureStores()

  const index = systemSkillStore.findIndex((s) => s.systemSkillId === id)
  if (index === -1) {
    throw new Error('全局技能不存在')
  }

  systemSkillStore[index] = {
    ...systemSkillStore[index],
    status,
    updated: new Date().toISOString(),
  }

  return toSystemSkillItem(systemSkillStore[index])
}

export async function deleteSystemSkill(id: number): Promise<void> {
  await mutationDelay()
  ensureStores()

  const index = systemSkillStore.findIndex((s) => s.systemSkillId === id)
  if (index === -1) {
    throw new Error('全局技能不存在')
  }
  systemSkillStore.splice(index, 1)
}

/* ============================================
   Async mock functions — Market skills
   ============================================ */

export async function getMarketSkills(params?: {
  page?: number
  pageSize?: number
  search?: string
  source?: string
  status?: string
}): Promise<PaginatedResponse<AdminMarketSkillItem>> {
  await listDelay()
  ensureStores()

  let filtered = marketSkillStore.map(toAdminMarketSkillItem)

  if (params?.search) {
    const q = params.search.toLowerCase()
    filtered = filtered.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        s.description.toLowerCase().includes(q),
    )
  }

  if (params?.source && params.source !== 'all') {
    filtered = filtered.filter((s) => s.source === params.source)
  }

  if (params?.status && params.status !== 'all') {
    const statusNum = Number(params.status)
    filtered = filtered.filter((s) => s.status === statusNum)
  }

  const page = params?.page ?? 1
  const pageSize = params?.pageSize ?? 20
  const totalCount = filtered.length
  const totalPage = Math.max(1, Math.ceil(totalCount / pageSize))
  const start = (page - 1) * pageSize
  const list = filtered.slice(start, start + pageSize)

  return { list, totalPage, totalCount, page, pageSize }
}

export async function getMarketSkillDetail(
  id: number,
): Promise<MarketSkillFull> {
  await itemDelay()
  ensureStores()

  const full = marketSkillStore.find((s) => s.marketSkillId === id)
  if (!full) {
    throw new Error('市场技能不存在')
  }

  return { ...full }
}

export async function updateMarketSkill(
  id: number,
  req: {
    icon?: string
    categoryId?: number
    sortOrder?: number
    remark?: string
  },
): Promise<AdminMarketSkillItem> {
  await mutationDelay()
  ensureStores()

  const index = marketSkillStore.findIndex((s) => s.marketSkillId === id)
  if (index === -1) {
    throw new Error('市场技能不存在')
  }

  const skill = marketSkillStore[index]
  let categoryName = skill.categoryName
  let categoryIcon = skill.categoryIcon

  if (req.categoryId && req.categoryId !== skill.categoryId) {
    const cat = categoryStore.find((c) => c.categoryId === req.categoryId)
    if (cat) {
      categoryName = cat.name
      categoryIcon = cat.icon
    }
  }

  marketSkillStore[index] = {
    ...skill,
    icon: req.icon ?? skill.icon,
    categoryId: req.categoryId ?? skill.categoryId,
    categoryName,
    categoryIcon,
    sortOrder: req.sortOrder ?? skill.sortOrder,
    remark: req.remark ?? skill.remark,
    updated: new Date().toISOString(),
  }

  return toAdminMarketSkillItem(marketSkillStore[index])
}

export async function toggleMarketSkillStatus(
  id: number,
  status: number,
): Promise<AdminMarketSkillItem> {
  await mutationDelay()
  ensureStores()

  const index = marketSkillStore.findIndex((s) => s.marketSkillId === id)
  if (index === -1) {
    throw new Error('市场技能不存在')
  }

  marketSkillStore[index] = {
    ...marketSkillStore[index],
    status,
    updated: new Date().toISOString(),
  }

  return toAdminMarketSkillItem(marketSkillStore[index])
}

export async function deleteMarketSkill(
  id: number,
  cascade?: boolean,
): Promise<void> {
  await mutationDelay()
  ensureStores()

  const index = marketSkillStore.findIndex((s) => s.marketSkillId === id)
  if (index === -1) {
    throw new Error('市场技能不存在')
  }

  marketSkillStore.splice(index, 1)

  if (cascade) {
    installedSkillStore = installedSkillStore.filter((r) => r.skillId !== id)
  }
}

export async function getMarketSkillUsers(
  id: number,
): Promise<InstalledSkillRecord[]> {
  await listDelay()
  ensureStores()

  return installedSkillStore.filter((r) => r.skillId === id)
}

export async function publishMarketSkill(
  uploadId: string,
  icon: string,
  categoryId: number,
): Promise<AdminMarketSkillItem> {
  await heavyDelay()
  ensureStores()

  const category =
    categoryStore.find((c) => c.categoryId === categoryId) ??
    categoryStore[0]
  const now = new Date().toISOString()
  const id = nextMarketSkillId()

  const full: MarketSkillFull = {
    marketSkillId: id,
    name: `custom-skill-${id}`,
    description: '从自定义 zip 发布的技能，已解析 SKILL.md 元数据',
    icon,
    source: 'custom_upload',
    sourceUrl: '',
    categoryId: category.categoryId,
    categoryName: category.name,
    categoryIcon: category.icon,
    installedCount: 0,
    status: 1,
    sortOrder: 70,
    remark: `mock uploadId: ${uploadId}`,
    content: `---\nname: custom-skill-${id}\ndescription: 从自定义 zip 发布的技能\n---\n\nMock skill content.`,
    created: now,
    updated: now,
  }
  marketSkillStore.unshift(full)

  return toAdminMarketSkillItem(full)
}

export async function importFromClawhub(
  slug: string,
  icon: string,
  categoryId: number,
): Promise<AdminMarketSkillItem> {
  await heavyDelay()
  ensureStores()

  const clawHubItem = clawHubStore.find((i) => i.slug === slug)
  if (!clawHubItem) {
    throw new Error('ClawHub 技能不存在')
  }

  const category =
    categoryStore.find((c) => c.categoryId === categoryId) ??
    categoryStore[0]
  const now = new Date().toISOString()
  const id = nextMarketSkillId()

  const full: MarketSkillFull = {
    marketSkillId: id,
    name: clawHubItem.slug,
    description: clawHubItem.summary,
    icon,
    source: 'clawhub',
    sourceUrl: `https://clawhub.dev/skills/${clawHubItem.slug}`,
    categoryId: category.categoryId,
    categoryName: category.name,
    categoryIcon: category.icon,
    installedCount: 0,
    status: 1,
    sortOrder: 60,
    remark: '',
    content: clawHubItem.content,
    created: now,
    updated: now,
  }
  marketSkillStore.unshift(full)

  return toAdminMarketSkillItem(full)
}

/* ============================================
   Async mock functions — Categories
   ============================================ */

export async function getSkillCategories(): Promise<AdminSkillCategoryItem[]> {
  await listDelay()
  ensureStores()
  return [...categoryStore]
}

export async function getAllSkillCategories(): Promise<AdminSkillCategoryItem[]> {
  await listDelay()
  ensureStores()
  return [...categoryStore]
}

export async function getSkillCategoryDetail(
  id: number,
): Promise<AdminSkillCategoryItem> {
  await itemDelay()
  ensureStores()

  const category = categoryStore.find((c) => c.categoryId === id)
  if (!category) {
    throw new Error('分类不存在')
  }

  return { ...category }
}

export async function createSkillCategory(
  name: string,
  icon: string,
): Promise<AdminSkillCategoryItem> {
  await mutationDelay()
  ensureStores()

  if (!name.trim()) {
    throw new Error('分类名称不能为空')
  }

  const now = new Date().toISOString()
  const category: AdminSkillCategoryItem = {
    categoryId: nextCategoryId(),
    name: name.trim(),
    icon,
    isDefault: 0,
    created: now,
    updated: now,
  }
  categoryStore.push(category)

  return { ...category }
}

export async function updateSkillCategory(
  id: number,
  req: { name?: string; icon?: string },
): Promise<AdminSkillCategoryItem> {
  await mutationDelay()
  ensureStores()

  const index = categoryStore.findIndex((c) => c.categoryId === id)
  if (index === -1) {
    throw new Error('分类不存在')
  }

  categoryStore[index] = {
    ...categoryStore[index],
    name: req.name ?? categoryStore[index].name,
    icon: req.icon ?? categoryStore[index].icon,
    updated: new Date().toISOString(),
  }

  // Also update market skills that use this category
  const cat = categoryStore[index]
  marketSkillStore = marketSkillStore.map((s) =>
    s.categoryId === id
      ? {
          ...s,
          categoryName: cat.name,
          categoryIcon: cat.icon,
        }
      : s,
  )

  return { ...categoryStore[index] }
}

export async function deleteSkillCategory(id: number): Promise<void> {
  await mutationDelay()
  ensureStores()

  const index = categoryStore.findIndex((c) => c.categoryId === id)
  if (index === -1) {
    throw new Error('分类不存在')
  }

  const defaultCategory = categoryStore.find((c) => c.isDefault === 1)
  if (defaultCategory) {
    categoryStore.splice(index, 1)
    marketSkillStore = marketSkillStore.map((s) =>
      s.categoryId === id
        ? {
            ...s,
            categoryId: defaultCategory.categoryId,
            categoryName: defaultCategory.name,
            categoryIcon: defaultCategory.icon,
          }
        : s,
    )
  }
}

/* ============================================
   Async mock functions — ClawHub
   ============================================ */

export async function searchClawhub(
  q: string,
  limit?: number,
): Promise<ClawHubSkillItem[]> {
  await listDelay()
  ensureStores()

  if (!q.trim()) {
    return []
  }

  const query = q.trim().toLowerCase()
  let results = clawHubStore
    .filter(
      (item) =>
        item.displayName.toLowerCase().includes(query) ||
        item.summary.toLowerCase().includes(query),
    )
    .map(toClawHubSkillItem)

  if (limit && limit > 0) {
    results = results.slice(0, limit)
  }

  return results
}

export async function exploreClawhub(
  sort: ClawHubSort,
): Promise<ClawHubSkillItem[]> {
  await listDelay()
  ensureStores()

  const sorted = [...clawHubStore].sort((a, b) => {
    switch (sort) {
      case 'newest':
      case 'updated':
        return (
          new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
        )
      case 'downloads':
        return b.downloads - a.downloads
      case 'stars':
        return b.stars - a.stars
      case 'installs':
        return b.installs - a.installs
      case 'trending':
      default:
        return b.score - a.score
    }
  })

  return sorted.map(toClawHubSkillItem)
}

export async function getClawhubSkillDetail(
  slug: string,
): Promise<ClawHubItemFull> {
  await itemDelay()
  ensureStores()

  const item = clawHubStore.find((i) => i.slug === slug)
  if (!item) {
    throw new Error('ClawHub 技能不存在')
  }

  return { ...item }
}

/* ============================================
   Helper — SKILL.md content parser
   ============================================ */

function parseSkillContent(content: string): {
  name: string
  description: string
  errors: string[]
} {
  const errors: string[] = []
  const match = content.match(/^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/)

  if (!match) {
    return {
      name: '',
      description: '',
      errors: ['YAML frontmatter 必须以单独一行 --- 开始和结束'],
    }
  }

  const frontmatter = match[1]
  const name =
    frontmatter
      .match(/^name:\s*(.+)$/m)
      ?.[1]
      ?.trim() ?? ''
  const description =
    frontmatter
      .match(/^description:\s*(.+)$/m)
      ?.[1]
      ?.trim() ?? ''

  if (!name) errors.push('缺少 name')
  if (!description) errors.push('缺少 description')
  if (name && !name.startsWith('goraven-'))
    errors.push('name 必须以 goraven- 开头')
  if (name && !/^goraven-[a-z0-9][a-z0-9-]*$/.test(name))
    errors.push('name 只能使用小写英文、数字和连字符')

  return { name, description, errors }
}
