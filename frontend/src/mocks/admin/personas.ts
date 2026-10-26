/* ============================================
   Admin Persona Templates — mock data
   Aligned with AdminPersonaTemplatesPage.tsx
   ============================================ */

import type {
  AdminPersonaTemplateItem,
  AdminPersonaCategoryItem,
  PaginatedResponse,
} from '../types'
import { listDelay, itemDelay, mutationDelay } from '../delay'

/* ============================================
   Constants
   ============================================ */

export const PAGE_SIZE = 15

/* ============================================
   In-memory store
   ============================================ */

let templates: AdminPersonaTemplateItem[] = []
let categories: AdminPersonaCategoryItem[] = []
let nextTemplateId = 0
let nextCategoryId = 0

/* ============================================
   Generate mock categories (12 categories)
   ============================================ */

export function generateMockTemplateCategories(): AdminPersonaCategoryItem[] {
  const presets: AdminPersonaCategoryItem[] = [
    { categoryId: 1, name: '通用', icon: 'bot', isDefault: 1, templateCount: 3, created: '2025-01-01T00:00:00Z', updated: '2025-01-01T00:00:00Z' },
    { categoryId: 2, name: '编程开发', icon: 'code', isDefault: 0, templateCount: 5, created: '2025-01-01T00:00:00Z', updated: '2025-06-15T10:30:00Z' },
    { categoryId: 3, name: '翻译语言', icon: 'languages', isDefault: 0, templateCount: 2, created: '2025-01-01T00:00:00Z', updated: '2025-01-01T00:00:00Z' },
    { categoryId: 4, name: '写作创作', icon: 'pen-line', isDefault: 0, templateCount: 3, created: '2025-01-01T00:00:00Z', updated: '2025-05-20T14:00:00Z' },
    { categoryId: 5, name: '数据分析', icon: 'bar-chart-3', isDefault: 0, templateCount: 2, created: '2025-01-01T00:00:00Z', updated: '2025-01-01T00:00:00Z' },
    { categoryId: 6, name: '学习教育', icon: 'graduation-cap', isDefault: 0, templateCount: 3, created: '2025-01-01T00:00:00Z', updated: '2025-06-25T09:00:00Z' },
    { categoryId: 7, name: '金融分析', icon: 'calculator', isDefault: 0, templateCount: 2, created: '2025-03-10T08:00:00Z', updated: '2025-06-22T14:00:00Z' },
    { categoryId: 8, name: '法律咨询', icon: 'shield', isDefault: 0, templateCount: 0, created: '2025-04-15T16:00:00Z', updated: '2025-04-15T16:00:00Z' },
    { categoryId: 9, name: '创意设计', icon: 'palette', isDefault: 0, templateCount: 0, created: '2025-05-01T09:00:00Z', updated: '2025-05-01T09:00:00Z' },
    { categoryId: 10, name: '健康医疗', icon: 'heart', isDefault: 0, templateCount: 0, created: '2025-05-10T10:00:00Z', updated: '2025-05-10T10:00:00Z' },
    { categoryId: 11, name: '运维工程', icon: 'terminal', isDefault: 0, templateCount: 0, created: '2025-06-01T11:00:00Z', updated: '2025-06-01T11:00:00Z' },
    { categoryId: 12, name: '知识库', icon: 'book-open', isDefault: 0, templateCount: 0, created: '2025-06-10T12:00:00Z', updated: '2025-06-10T12:00:00Z' },
  ]
  return presets
}

/* ============================================
   Generate mock templates (20 templates)
   ============================================ */

export function generateMockTemplates(
  cats?: AdminPersonaCategoryItem[],
): AdminPersonaTemplateItem[] {
  const catList = cats ?? generateMockTemplateCategories()
  const getCategory = (id: number) => catList.find((c) => c.categoryId === id)!

  const raw = [
    {
      templateId: 1, name: '通用助手', icon: 'bot',
      description: '日常问答、信息查询的通用 AI 助手',
      roleInfo: '你是一个有帮助的AI助手。你可以回答各种问题，提供建议，帮助用户解决日常问题。请保持回答简洁、准确、友好。如果你不确定答案，请诚实地告诉用户。',
      categoryId: 1, usageCount: 256, sortOrder: 0,
      created: '2025-01-15T08:00:00Z', updated: '2025-06-20T10:00:00Z',
    },
    {
      templateId: 2, name: 'Python 开发专家', icon: 'code',
      description: 'Python 代码编写、调试和架构设计',
      roleInfo: '你是一位资深的 Python 开发专家。你擅长编写高质量的 Python 代码，熟悉 PEP 8 规范，精通 Django、Flask、FastAPI 等主流框架。你能帮助用户进行代码审查、性能优化、架构设计和调试。请给出清晰、可运行的代码示例，并解释关键设计决策。',
      categoryId: 2, usageCount: 189, sortOrder: 0,
      created: '2025-01-20T09:00:00Z', updated: '2025-06-18T14:00:00Z',
    },
    {
      templateId: 3, name: '前端工程师', icon: 'zap',
      description: 'React、TypeScript 前端开发',
      roleInfo: '你是一位经验丰富的前端工程师。你精通 React、TypeScript、Next.js、Tailwind CSS 等现代前端技术栈。你能帮助用户解决组件设计、状态管理、性能优化、响应式布局等问题。请提供可维护、可测试的代码方案。',
      categoryId: 2, usageCount: 145, sortOrder: 1,
      created: '2025-01-25T10:00:00Z', updated: '2025-06-10T09:00:00Z',
    },
    {
      templateId: 4, name: 'Go 后端开发', icon: 'code',
      description: 'Go 语言后端服务开发',
      roleInfo: '你是一位 Go 语言后端开发专家。你熟悉 Go 标准库和常用第三方库，擅长微服务架构、API 设计、数据库优化和并发编程。请提供简洁、高效、符合 Go 惯例的代码方案，并注意错误处理和性能考量。',
      categoryId: 2, usageCount: 98, sortOrder: 2,
      created: '2025-02-01T11:00:00Z', updated: '2025-05-30T16:00:00Z',
    },
    {
      templateId: 5, name: 'DevOps 工程师', icon: 'terminal',
      description: 'CI/CD、容器化和基础设施即代码',
      roleInfo: '你是一位 DevOps 工程师。你精通 Docker、Kubernetes、Terraform、GitHub Actions 等工具。你能帮助用户设计 CI/CD 流水线、容器化应用、管理云基础设施。请提供安全、可扩展、易于维护的解决方案。',
      categoryId: 2, usageCount: 67, sortOrder: 3,
      created: '2025-02-10T12:00:00Z', updated: '2025-04-20T08:00:00Z',
    },
    {
      templateId: 6, name: '全栈工程师', icon: 'wrench',
      description: '前后端全栈开发',
      roleInfo: '你是一位全栈工程师，精通前后端技术。你能从数据库设计到 UI 实现提供端到端的解决方案。请根据用户的需求选择最合适的技术方案，并解释技术选型的理由。',
      categoryId: 2, usageCount: 112, sortOrder: 4,
      created: '2025-03-01T13:00:00Z', updated: '2025-06-01T11:00:00Z',
    },
    {
      templateId: 7, name: '中英翻译', icon: 'languages',
      description: '中英文互译，保持语境和风格',
      roleInfo: '你是一位专业的翻译人员，精通中文和英文。你不仅能进行字面翻译，还能根据上下文保持原文的语气、风格和文化适配。请提供准确的翻译，并在必要时附上翻译说明。',
      categoryId: 3, usageCount: 320, sortOrder: 0,
      created: '2025-01-15T08:00:00Z', updated: '2025-06-22T15:00:00Z',
    },
    {
      templateId: 8, name: '多语言翻译', icon: 'languages',
      description: '支持中英日韩等多语言互译',
      roleInfo: '你是一位多语言翻译专家。你精通中文、英文、日文、韩文等多种语言。你能帮助用户进行多语言翻译、本地化适配和跨文化沟通。请确保翻译准确、自然，符合目标语言的表达习惯。',
      categoryId: 3, usageCount: 178, sortOrder: 1,
      created: '2025-02-20T09:00:00Z', updated: '2025-05-15T10:00:00Z',
    },
    {
      templateId: 9, name: '文案撰写', icon: 'pen-line',
      description: '营销文案、广告语和品牌故事创作',
      roleInfo: '你是一位专业的文案撰稿人。你擅长撰写各种类型的文案，包括营销文案、广告语、品牌故事、产品描述等。请根据目标受众和品牌调性，创作有吸引力、有说服力的文案。',
      categoryId: 4, usageCount: 89, sortOrder: 0,
      created: '2025-02-15T10:00:00Z', updated: '2025-06-05T12:00:00Z',
    },
    {
      templateId: 10, name: '技术文档', icon: 'file-text',
      description: 'API 文档、技术说明和教程撰写',
      roleInfo: '你是一位技术文档撰写专家。你擅长编写清晰、准确、易于理解的技术文档，包括 API 文档、使用指南、教程和故障排除文档。请使用简洁的语言，提供代码示例，并注意文档的结构和可读性。',
      categoryId: 4, usageCount: 56, sortOrder: 1,
      created: '2025-03-10T11:00:00Z', updated: '2025-05-25T14:00:00Z',
    },
    {
      templateId: 11, name: '创意写作', icon: 'pen-line',
      description: '故事创作、剧本和诗歌',
      roleInfo: '你是一位创意写作者。你擅长创作各种类型的文学作品，包括短篇小说、剧本、诗歌等。请根据用户的需求发挥想象力，创作有深度、有感染力的作品。',
      categoryId: 4, usageCount: 43, sortOrder: 2,
      created: '2025-04-01T12:00:00Z', updated: '2025-04-01T12:00:00Z',
    },
    {
      templateId: 12, name: '数据分析师', icon: 'bar-chart-3',
      description: '数据处理、可视化和商业分析',
      roleInfo: '你是一位数据分析师。你擅长使用 Python（pandas、numpy、matplotlib）和 SQL 进行数据分析。你能帮助用户清洗数据、探索性分析、构建可视化报表和提取业务洞察。请提供可复现的分析代码和清晰的结论。',
      categoryId: 5, usageCount: 134, sortOrder: 0,
      created: '2025-02-05T09:00:00Z', updated: '2025-06-12T16:00:00Z',
    },
    {
      templateId: 13, name: '数据科学家', icon: 'bar-chart-3',
      description: '机器学习建模和统计推断',
      roleInfo: '你是一位数据科学家。你精通机器学习算法、统计建模和实验设计。你能帮助用户进行特征工程、模型选择、超参数调优和结果解释。请给出严谨的分析方法，并注意模型的解释性和可部署性。',
      categoryId: 5, usageCount: 78, sortOrder: 1,
      created: '2025-03-20T10:00:00Z', updated: '2025-05-10T08:00:00Z',
    },
    {
      templateId: 14, name: '学习导师', icon: 'graduation-cap',
      description: '知识问答、概念解释和学习规划',
      roleInfo: '你是一位耐心、知识渊博的学习导师。你善于将复杂的概念用简单易懂的方式解释清楚。你能帮助用户理解各种学科的知识，制定学习计划，并提供学习资源建议。请用循序渐进的方式讲解，鼓励用户思考和提问。',
      categoryId: 6, usageCount: 201, sortOrder: 0,
      created: '2025-01-10T08:00:00Z', updated: '2025-06-15T09:00:00Z',
    },
    {
      templateId: 15, name: '学术研究', icon: 'graduation-cap',
      description: '论文写作、文献综述和研究方法',
      roleInfo: '你是一位学术研究助手。你熟悉学术论文的写作规范、文献综述方法和研究设计。你能帮助用户梳理研究思路、撰写论文、润色文字和检查引用格式。请保持学术严谨性，提供有理有据的建议。',
      categoryId: 6, usageCount: 92, sortOrder: 1,
      created: '2025-03-15T11:00:00Z', updated: '2025-05-28T13:00:00Z',
    },
    {
      templateId: 16, name: '金融顾问', icon: 'calculator',
      description: '投资分析、风险评估和财务规划',
      roleInfo: '你是一位金融顾问。你精通投资分析、风险评估、资产配置和财务规划。你能帮助用户分析市场趋势、评估投资机会、制定财务目标。请注意：你提供的是信息和分析，不构成投资建议。用户在做出任何投资决策前应咨询持牌专业人士。',
      categoryId: 7, usageCount: 45, sortOrder: 0,
      created: '2025-04-10T14:00:00Z', updated: '2025-06-08T10:00:00Z',
    },
    {
      templateId: 17, name: '日常问答', icon: 'message-circle',
      description: '生活常识、百科知识问答',
      roleInfo: '你是一个知识丰富的问答助手。你可以回答各种领域的问题，从科学技术到人文历史，从生活常识到专业领域。请提供准确、客观的信息，并注明信息来源（如有）。',
      categoryId: 1, usageCount: 410, sortOrder: 1,
      created: '2025-01-15T08:00:00Z', updated: '2025-06-22T08:00:00Z',
    },
    {
      templateId: 18, name: '客服助手', icon: 'heart',
      description: '客户服务、问题解答和投诉处理',
      roleInfo: '你是一位专业的客服助手。你能礼貌、耐心地处理用户的问题和投诉。请始终以用户为中心，积极解决问题，保持专业和友善的态度。如果遇到无法解决的问题，请明确告知用户并提供替代方案。',
      categoryId: 1, usageCount: 167, sortOrder: 2,
      created: '2025-02-28T09:00:00Z', updated: '2025-06-20T15:00:00Z',
    },
    {
      templateId: 19, name: 'AI 研究顾问', icon: 'brain',
      description: 'AI 技术研究、论文解读和趋势分析',
      roleInfo: '你是一位 AI 研究顾问。你精通机器学习、深度学习、自然语言处理、计算机视觉等 AI 领域的前沿技术。你能帮助用户解读学术论文、分析技术趋势、评估模型架构和训练策略。请提供有深度的技术见解，并引用相关研究。',
      categoryId: 6, usageCount: 34, sortOrder: 2,
      created: '2025-05-15T10:00:00Z', updated: '2025-06-25T09:00:00Z',
    },
    {
      templateId: 20, name: '国际资讯', icon: 'globe',
      description: '国际新闻摘要、跨文化沟通和时政分析',
      roleInfo: '你是一位国际资讯分析员。你关注全球时事、经济动态和跨文化议题。你能帮助用户理解国际事件的背景、各方立场和潜在影响。请保持客观、中立，提供多角度分析，并明确区分事实与观点。',
      categoryId: 7, usageCount: 28, sortOrder: 1,
      created: '2025-06-01T08:00:00Z', updated: '2025-06-22T14:00:00Z',
    },
  ]

  return raw.map((t) => {
    const cat = getCategory(t.categoryId)
    return { ...t, categoryName: cat.name, categoryIcon: cat.icon }
  })
}

/* ============================================
   Initialize
   ============================================ */

function ensureInit(): void {
  if (categories.length === 0) {
    categories = generateMockTemplateCategories()
    nextCategoryId = categories.reduce((max, c) => Math.max(max, c.categoryId), 0) + 1
  }
  if (templates.length === 0) {
    templates = generateMockTemplates(categories)
    nextTemplateId = templates.reduce((max, t) => Math.max(max, t.templateId), 0) + 1
  }
}

/* ============================================
   Templates — CRUD
   ============================================ */

export interface GetTemplatesParams {
  page?: number
  pageSize?: number
  search?: string
  categoryId?: number
}

export async function getTemplates(
  params: GetTemplatesParams = {},
): Promise<PaginatedResponse<AdminPersonaTemplateItem>> {
  await listDelay()
  ensureInit()

  const page = params.page ?? 1
  const pageSize = params.pageSize ?? PAGE_SIZE

  let filtered = [...templates]

  if (params.search?.trim()) {
    const q = params.search.trim().toLowerCase()
    filtered = filtered.filter((t) => t.name.toLowerCase().includes(q))
  }

  if (params.categoryId != null) {
    filtered = filtered.filter((t) => t.categoryId === params.categoryId)
  }

  const totalCount = filtered.length
  const totalPage = Math.max(1, Math.ceil(totalCount / pageSize))
  const safePage = Math.min(page, totalPage)
  const start = (safePage - 1) * pageSize
  const list = filtered.slice(start, start + pageSize)

  return { list, totalPage, totalCount, page: safePage, pageSize }
}

export async function getTemplateDetail(
  id: number,
): Promise<AdminPersonaTemplateItem> {
  await itemDelay()
  ensureInit()

  const found = templates.find((t) => t.templateId === id)
  if (!found) throw new Error(`Template ${id} not found`)
  return { ...found }
}

export interface CreateTemplateRequest {
  name: string
  icon: string
  description: string
  roleInfo: string
  categoryId: number
  sortOrder: number
}

export async function createTemplate(
  req: CreateTemplateRequest,
): Promise<AdminPersonaTemplateItem> {
  await mutationDelay()
  ensureInit()

  const cat = categories.find((c) => c.categoryId === req.categoryId)
  if (!cat) throw new Error(`Category ${req.categoryId} not found`)

  const now = new Date().toISOString()
  const created: AdminPersonaTemplateItem = {
    templateId: nextTemplateId++,
    name: req.name,
    icon: req.icon,
    description: req.description,
    roleInfo: req.roleInfo,
    categoryId: req.categoryId,
    categoryName: cat.name,
    categoryIcon: cat.icon,
    usageCount: 0,
    sortOrder: req.sortOrder,
    created: now,
    updated: now,
  }

  templates = [created, ...templates]

  // Update category count
  categories = categories.map((c) =>
    c.categoryId === req.categoryId
      ? { ...c, templateCount: c.templateCount + 1 }
      : c,
  )

  return created
}

export interface UpdateTemplateRequest {
  name?: string
  icon?: string
  description?: string
  roleInfo?: string
  categoryId?: number
  sortOrder?: number
}

export async function updateTemplate(
  id: number,
  req: UpdateTemplateRequest,
): Promise<AdminPersonaTemplateItem> {
  await mutationDelay()
  ensureInit()

  const idx = templates.findIndex((t) => t.templateId === id)
  if (idx === -1) throw new Error(`Template ${id} not found`)

  const existing = templates[idx]
  const oldCategoryId = existing.categoryId
  const newCategoryId = req.categoryId ?? oldCategoryId
  const cat = categories.find((c) => c.categoryId === newCategoryId)
  if (!cat) throw new Error(`Category ${newCategoryId} not found`)

  const updated: AdminPersonaTemplateItem = {
    ...existing,
    name: req.name ?? existing.name,
    icon: req.icon ?? existing.icon,
    description: req.description ?? existing.description,
    roleInfo: req.roleInfo ?? existing.roleInfo,
    categoryId: newCategoryId,
    categoryName: cat.name,
    categoryIcon: cat.icon,
    sortOrder: req.sortOrder ?? existing.sortOrder,
    updated: new Date().toISOString(),
  }

  templates = templates.map((t) => (t.templateId === id ? updated : t))

  // Update category counts if changed
  if (oldCategoryId !== newCategoryId) {
    categories = categories.map((c) => {
      if (c.categoryId === oldCategoryId) return { ...c, templateCount: Math.max(0, c.templateCount - 1) }
      if (c.categoryId === newCategoryId) return { ...c, templateCount: c.templateCount + 1 }
      return c
    })
  }

  return updated
}

export async function deleteTemplate(id: number): Promise<void> {
  await mutationDelay()
  ensureInit()

  const target = templates.find((t) => t.templateId === id)
  if (!target) throw new Error(`Template ${id} not found`)

  templates = templates.filter((t) => t.templateId !== id)

  // Decrease category count
  categories = categories.map((c) =>
    c.categoryId === target.categoryId
      ? { ...c, templateCount: Math.max(0, c.templateCount - 1) }
      : c,
  )
}

/* ============================================
   Categories — CRUD
   ============================================ */

export interface GetCategoriesParams {
  page?: number
  pageSize?: number
}

export async function getCategories(
  params: GetCategoriesParams = {},
): Promise<PaginatedResponse<AdminPersonaCategoryItem>> {
  await listDelay()
  ensureInit()

  const page = params.page ?? 1
  const pageSize = params.pageSize ?? PAGE_SIZE

  const sorted = [...categories].sort((a, b) => a.name.localeCompare(b.name, 'zh'))
  const totalCount = sorted.length
  const totalPage = Math.max(1, Math.ceil(totalCount / pageSize))
  const safePage = Math.min(page, totalPage)
  const start = (safePage - 1) * pageSize
  const list = sorted.slice(start, start + pageSize)

  return { list, totalPage, totalCount, page: safePage, pageSize }
}

export async function getAllCategories(): Promise<AdminPersonaCategoryItem[]> {
  await listDelay()
  ensureInit()

  return [...categories].sort((a, b) => a.name.localeCompare(b.name, 'zh'))
}

export async function getCategoryDetail(
  id: number,
): Promise<AdminPersonaCategoryItem> {
  await itemDelay()
  ensureInit()

  const found = categories.find((c) => c.categoryId === id)
  if (!found) throw new Error(`Category ${id} not found`)
  return { ...found }
}

export async function createCategory(
  name: string,
  icon: string,
): Promise<AdminPersonaCategoryItem> {
  await mutationDelay()
  ensureInit()

  const now = new Date().toISOString()
  const created: AdminPersonaCategoryItem = {
    categoryId: nextCategoryId++,
    name,
    icon,
    isDefault: 0,
    templateCount: 0,
    created: now,
    updated: now,
  }

  categories = [...categories, created]
  return created
}

export interface UpdateCategoryRequest {
  name?: string
  icon?: string
}

export async function updateCategory(
  id: number,
  req: UpdateCategoryRequest,
): Promise<AdminPersonaCategoryItem> {
  await mutationDelay()
  ensureInit()

  const idx = categories.findIndex((c) => c.categoryId === id)
  if (idx === -1) throw new Error(`Category ${id} not found`)

  const existing = categories[idx]
  const updated: AdminPersonaCategoryItem = {
    ...existing,
    name: req.name ?? existing.name,
    icon: req.icon ?? existing.icon,
    updated: new Date().toISOString(),
  }

  // Update category and sync templates that reference it
  categories = categories.map((c) => (c.categoryId === id ? updated : c))

  if (req.name !== undefined || req.icon !== undefined) {
    templates = templates.map((t) =>
      t.categoryId === id
        ? { ...t, categoryName: updated.name, categoryIcon: updated.icon }
        : t,
    )
  }

  return updated
}

export async function deleteCategory(id: number): Promise<void> {
  await mutationDelay()
  ensureInit()

  const target = categories.find((c) => c.categoryId === id)
  if (!target) throw new Error(`Category ${id} not found`)

  const defaultCat = categories.find((c) => c.isDefault === 1)!

  // Move templates to default category
  templates = templates.map((t) =>
    t.categoryId === id
      ? {
          ...t,
          categoryId: defaultCat.categoryId,
          categoryName: defaultCat.name,
          categoryIcon: defaultCat.icon,
        }
      : t,
  )

  // Update template counts: add moved count to default, remove deleted category
  categories = categories
    .filter((c) => c.categoryId !== id)
    .map((c) =>
      c.categoryId === defaultCat.categoryId
        ? { ...c, templateCount: c.templateCount + target.templateCount }
        : c,
    )
}
