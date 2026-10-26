import type {
  PersonaListItem,
  PersonaDetail,
  CreatePersonaRequest,
  TemplateItem,
  TemplateDetail,
  TemplateCategoryItem,
  PersonaCategoryItem,
} from './types'
import { listDelay, itemDelay, mutationDelay } from './delay'

/* ============================================
   Local types for edit page
   ============================================ */

export interface PersonaModelItem {
  aiModelId: number
  providerName: string
  displayName: string
  modelName: string
  icon: string
}

export interface PersonaMcpItem {
  mcpId: number
  name: string
  icon: string
  description: string
}

export interface PersonaSkillItem {
  userSkillId: number
  name: string
  description: string
  icon: string
}

/* ============================================
   Internal state (mutable across calls)
   ============================================ */

const _initialPersonas: PersonaListItem[] = [
  { personaId: 1, name: '编程助手', icon: 'code', categoryName: '开发', modelName: 'DeepSeek V3', roleInfo: '你是一个专业的编程助手，擅长 Go、TypeScript 和 Python。请用简洁的代码示例回答问题，注意代码风格和最佳实践。遇到复杂问题时，先分析需求再给出方案。', mcpNames: ['filesystem', 'github', 'brave-search'], skillNames: ['code-review', 'git-helper', 'regex-builder'] },
  { personaId: 2, name: '翻译专家', icon: 'globe', categoryName: '语言', modelName: '默认模型', roleInfo: '你是一个专业翻译，精通中英日法互译。保持原文语气和风格，技术文档翻译保留术语一致性。', mcpNames: ['brave-search'], skillNames: [] },
  { personaId: 3, name: '数据分析师', icon: 'search', categoryName: '数据', modelName: 'Claude 3.5 Sonnet', roleInfo: '你是一个数据分析师，从原始数据中提取洞察，生成可视化报告。支持 CSV、JSON 和数据库查询。', mcpNames: ['redis', 'database'], skillNames: ['data-analyzer', 'csv-parser'] },
  { personaId: 4, name: '文档写手', icon: 'file-text', categoryName: '写作', modelName: '默认模型', roleInfo: '你是一个技术文档写手，撰写技术文档、PRD 和会议纪要。输出简洁，结构化，可直接交付。', mcpNames: [], skillNames: ['markdown-writer'] },
  { personaId: 5, name: '运维助手', icon: 'terminal', categoryName: '运维', modelName: '默认模型', roleInfo: '你是一个运维工程师，管理 Docker 容器和 K8s 集群，分析日志，处理 CI/CD 流水线故障。', mcpNames: ['docker', 'kubernetes', 'redis', 'stripe'], skillNames: ['docker-helper', 'log-analyzer'] },
  { personaId: 6, name: '通用助手', icon: 'bot', categoryName: '通用', modelName: '默认模型', roleInfo: '你是一个通用助手，回答各类问题，可调用基础工具完成日常任务。', mcpNames: ['filesystem'], skillNames: [] },
  { personaId: 7, name: 'SQL 专家', icon: 'terminal', categoryName: '数据', modelName: 'GPT-4o', roleInfo: '你是一个 SQL 专家，精通查询优化、表结构设计和数据库迁移。支持 PostgreSQL 和 MySQL。', mcpNames: ['database'], skillNames: ['sql-optimizer', 'data-analyzer'] },
  { personaId: 8, name: '安全审计', icon: 'search', categoryName: '运维', modelName: 'Claude 3.5 Sonnet', roleInfo: '你是一个安全审计专家，审查代码安全漏洞，检查配置合规性，生成安全审计报告。', mcpNames: ['filesystem', 'github'], skillNames: ['code-review', 'log-analyzer', 'regex-builder'] },
  { personaId: 9, name: '前端架构师', icon: 'code', categoryName: '开发', modelName: 'DeepSeek V3', roleInfo: '你是一个前端架构师，React 和 TypeScript 专家，专注组件架构和性能优化。', mcpNames: ['github', 'brave-search'], skillNames: ['code-review', 'git-helper'] },
  { personaId: 10, name: '英文润色', icon: 'globe', categoryName: '语言', modelName: '默认模型', roleInfo: '你是一个英文润色专家，改进语法、语气和流畅度。适用于邮件、论文和文档。', mcpNames: [], skillNames: [] },
  { personaId: 19, name: '法语翻译', icon: 'globe', categoryName: '语言', modelName: '默认模型', roleInfo: '你是一个专业法语翻译，覆盖商务、文学和技术领域。', mcpNames: [], skillNames: [] },
  { personaId: 20, name: '数据分析 V2', icon: 'search', categoryName: '数据', modelName: 'GPT-4o', roleInfo: '你是一个高级数据分析师，支持机器学习特征工程和统计建模。', mcpNames: ['database', 'redis'], skillNames: ['data-analyzer', 'csv-parser', 'chart-generator'] },
]

export const PERSONAS: PersonaListItem[] = [..._initialPersonas]

const _initialDetails: Record<number, PersonaDetail> = {
  1: {
    personaId: 1,
    name: '编程助手',
    icon: 'code',
    roleInfo: '你是一个专业的编程助手，擅长 Go、TypeScript 和 Python。请用简洁的代码示例回答问题，注意代码风格和最佳实践。遇到复杂问题时，先分析需求再给出方案。',
    categoryId: 2,
    categoryName: '开发',
    categoryIcon: 'code',
    mcpIds: [1, 2, 3],
    mcpNames: [
      { id: 1, name: 'filesystem', icon: 'file-text' },
      { id: 2, name: 'brave-search', icon: 'search' },
      { id: 3, name: 'github', icon: 'code' },
    ],
    skillIds: [1, 6, 10, 12],
    skillNames: [
      { id: 1, name: 'code-review', icon: 'code' },
      { id: 6, name: 'log-analyzer', icon: 'terminal' },
      { id: 10, name: 'git-helper', icon: 'code' },
      { id: 12, name: 'regex-builder', icon: 'code' },
    ],
    aiModelId: 1,
    modelName: 'DeepSeek - DeepSeek V3',
    modelIcon: '/logos/deepseek.svg',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
  2: {
    personaId: 2,
    name: '翻译专家',
    icon: 'globe',
    roleInfo: '你是一个专业翻译，精通中英日法互译。保持原文语气和风格，技术文档翻译保留术语一致性。',
    categoryId: 3,
    categoryName: '语言',
    categoryIcon: 'globe',
    mcpIds: [2],
    mcpNames: [
      { id: 2, name: 'brave-search', icon: 'search' },
    ],
    skillIds: [],
    skillNames: [],
    aiModelId: 0,
    modelName: '默认模型',
    modelIcon: '',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
  3: {
    personaId: 3,
    name: '数据分析师',
    icon: 'search',
    roleInfo: '你是一个数据分析师，从原始数据中提取洞察，生成可视化报告。支持 CSV、JSON 和数据库查询。',
    categoryId: 4,
    categoryName: '数据',
    categoryIcon: 'search',
    mcpIds: [5, 8],
    mcpNames: [
      { id: 5, name: 'redis', icon: 'terminal' },
      { id: 8, name: 'database', icon: 'database' },
    ],
    skillIds: [3, 9],
    skillNames: [
      { id: 3, name: 'data-analyzer', icon: 'search' },
      { id: 9, name: 'csv-parser', icon: 'file-text' },
    ],
    aiModelId: 2,
    modelName: 'Claude - Claude 3.5 Sonnet',
    modelIcon: '/logos/claude.svg',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
  4: {
    personaId: 4,
    name: '文档写手',
    icon: 'file-text',
    roleInfo: '你是一个技术文档写手，撰写技术文档、PRD 和会议纪要。输出简洁，结构化，可直接交付。',
    categoryId: 5,
    categoryName: '写作',
    categoryIcon: 'file-text',
    mcpIds: [],
    mcpNames: [],
    skillIds: [11],
    skillNames: [
      { id: 11, name: 'markdown-writer', icon: 'file-text' },
    ],
    aiModelId: 0,
    modelName: '默认模型',
    modelIcon: '',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
  5: {
    personaId: 5,
    name: '运维助手',
    icon: 'terminal',
    roleInfo: '你是一个运维工程师，管理 Docker 容器和 K8s 集群，分析日志，处理 CI/CD 流水线故障。',
    categoryId: 6,
    categoryName: '运维',
    categoryIcon: 'terminal',
    mcpIds: [5, 9, 10, 12],
    mcpNames: [
      { id: 5, name: 'redis', icon: 'terminal' },
      { id: 9, name: 'docker', icon: 'terminal' },
      { id: 10, name: 'kubernetes', icon: 'terminal' },
      { id: 12, name: 'stripe', icon: 'star' },
    ],
    skillIds: [2, 6],
    skillNames: [
      { id: 2, name: 'docker-helper', icon: 'wrench' },
      { id: 6, name: 'log-analyzer', icon: 'terminal' },
    ],
    aiModelId: 0,
    modelName: '默认模型',
    modelIcon: '',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
  6: {
    personaId: 6,
    name: '通用助手',
    icon: 'bot',
    roleInfo: '你是一个通用助手，回答各类问题，可调用基础工具完成日常任务。',
    categoryId: 1,
    categoryName: '通用',
    categoryIcon: 'bot',
    mcpIds: [1],
    mcpNames: [
      { id: 1, name: 'filesystem', icon: 'file-text' },
    ],
    skillIds: [],
    skillNames: [],
    aiModelId: 0,
    modelName: '默认模型',
    modelIcon: '',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
  7: {
    personaId: 7,
    name: 'SQL 专家',
    icon: 'terminal',
    roleInfo: '你是一个 SQL 专家，精通查询优化、表结构设计和数据库迁移。支持 PostgreSQL 和 MySQL。',
    categoryId: 4,
    categoryName: '数据',
    categoryIcon: 'search',
    mcpIds: [8],
    mcpNames: [
      { id: 8, name: 'database', icon: 'database' },
    ],
    skillIds: [3, 13],
    skillNames: [
      { id: 3, name: 'data-analyzer', icon: 'search' },
      { id: 13, name: 'sql-optimizer', icon: 'database' },
    ],
    aiModelId: 3,
    modelName: 'OpenAI - GPT-4o',
    modelIcon: '/logos/openai.svg',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
  8: {
    personaId: 8,
    name: '安全审计',
    icon: 'search',
    roleInfo: '你是一个安全审计专家，审查代码安全漏洞，检查配置合规性，生成安全审计报告。',
    categoryId: 6,
    categoryName: '运维',
    categoryIcon: 'terminal',
    mcpIds: [1, 3],
    mcpNames: [
      { id: 1, name: 'filesystem', icon: 'file-text' },
      { id: 3, name: 'github', icon: 'code' },
    ],
    skillIds: [1, 6, 12],
    skillNames: [
      { id: 1, name: 'code-review', icon: 'code' },
      { id: 6, name: 'log-analyzer', icon: 'terminal' },
      { id: 12, name: 'regex-builder', icon: 'code' },
    ],
    aiModelId: 2,
    modelName: 'Claude - Claude 3.5 Sonnet',
    modelIcon: '/logos/claude.svg',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
  9: {
    personaId: 9,
    name: '前端架构师',
    icon: 'code',
    roleInfo: '你是一个前端架构师，React 和 TypeScript 专家，专注组件架构和性能优化。',
    categoryId: 2,
    categoryName: '开发',
    categoryIcon: 'code',
    mcpIds: [2, 3],
    mcpNames: [
      { id: 2, name: 'brave-search', icon: 'search' },
      { id: 3, name: 'github', icon: 'code' },
    ],
    skillIds: [1, 10],
    skillNames: [
      { id: 1, name: 'code-review', icon: 'code' },
      { id: 10, name: 'git-helper', icon: 'code' },
    ],
    aiModelId: 1,
    modelName: 'DeepSeek - DeepSeek V3',
    modelIcon: '/logos/deepseek.svg',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
  10: {
    personaId: 10,
    name: '英文润色',
    icon: 'globe',
    roleInfo: '你是一个英文润色专家，改进语法、语气和流畅度。适用于邮件、论文和文档。',
    categoryId: 3,
    categoryName: '语言',
    categoryIcon: 'globe',
    mcpIds: [],
    mcpNames: [],
    skillIds: [],
    skillNames: [],
    aiModelId: 0,
    modelName: '默认模型',
    modelIcon: '',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
  19: {
    personaId: 19,
    name: '法语翻译',
    icon: 'globe',
    roleInfo: '你是一个专业法语翻译，覆盖商务、文学和技术领域。',
    categoryId: 3,
    categoryName: '语言',
    categoryIcon: 'globe',
    mcpIds: [],
    mcpNames: [],
    skillIds: [],
    skillNames: [],
    aiModelId: 0,
    modelName: '默认模型',
    modelIcon: '',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
  20: {
    personaId: 20,
    name: '数据分析 V2',
    icon: 'search',
    roleInfo: '你是一个高级数据分析师，支持机器学习特征工程和统计建模。',
    categoryId: 4,
    categoryName: '数据',
    categoryIcon: 'search',
    mcpIds: [5, 8],
    mcpNames: [
      { id: 5, name: 'redis', icon: 'terminal' },
      { id: 8, name: 'database', icon: 'database' },
    ],
    skillIds: [3, 9, 14],
    skillNames: [
      { id: 3, name: 'data-analyzer', icon: 'search' },
      { id: 9, name: 'csv-parser', icon: 'file-text' },
      { id: 14, name: 'chart-generator', icon: 'bar-chart' },
    ],
    aiModelId: 3,
    modelName: 'OpenAI - GPT-4o',
    modelIcon: '/logos/openai.svg',
    created: '2025-01-15T08:00:00Z',
    updated: '2025-06-01T10:30:00Z',
  },
}

export const PERSONA_DETAILS: Record<number, PersonaDetail> = { ..._initialDetails }

/* ============================================
   Templates (moved from persona-templates.ts)
   ============================================ */

const _fullTemplates: TemplateDetail[] = [
  { templateId: 1, name: '通用助手', icon: 'bot', categoryId: 1, categoryName: '通用', categoryIcon: 'bot', description: '日常问答、信息查询的通用 AI 助手', roleInfo: '你是一个有帮助的AI助手。你可以回答各种问题，提供建议，帮助用户解决日常问题。请保持回答简洁、准确、友好。如果你不确定答案，请诚实地告诉用户。' },
  { templateId: 2, name: 'Python 开发专家', icon: 'code', categoryId: 2, categoryName: '编程开发', categoryIcon: 'code', description: 'Python 代码编写、调试和架构设计', roleInfo: '你是一位资深的 Python 开发专家。你擅长编写高质量的 Python 代码，熟悉 PEP 8 规范，精通 Django、Flask、FastAPI 等主流框架。你能帮助用户进行代码审查、性能优化、架构设计和调试。请给出清晰、可运行的代码示例，并解释关键设计决策。' },
  { templateId: 3, name: '前端工程师', icon: 'zap', categoryId: 2, categoryName: '编程开发', categoryIcon: 'code', description: 'React、TypeScript 前端开发', roleInfo: '你是一位经验丰富的前端工程师。你精通 React、TypeScript、Next.js、Tailwind CSS 等现代前端技术栈。你能帮助用户解决组件设计、状态管理、性能优化、响应式布局等问题。请提供可维护、可测试的代码方案。' },
  { templateId: 4, name: 'Go 后端开发', icon: 'code', categoryId: 2, categoryName: '编程开发', categoryIcon: 'code', description: 'Go 语言后端服务开发', roleInfo: '你是一位 Go 语言后端开发专家。你熟悉 Go 标准库和常用第三方库，擅长微服务架构、API 设计、数据库优化和并发编程。请提供简洁、高效、符合 Go 惯例的代码方案，并注意错误处理和性能考量。' },
  { templateId: 5, name: 'DevOps 工程师', icon: 'terminal', categoryId: 2, categoryName: '编程开发', categoryIcon: 'code', description: 'CI/CD、容器化和基础设施即代码', roleInfo: '你是一位 DevOps 工程师。你精通 Docker、Kubernetes、Terraform、GitHub Actions 等工具。你能帮助用户设计 CI/CD 流水线、容器化应用、管理云基础设施。请提供安全、可扩展、易于维护的解决方案。' },
  { templateId: 6, name: '全栈工程师', icon: 'wrench', categoryId: 2, categoryName: '编程开发', categoryIcon: 'code', description: '前后端全栈开发', roleInfo: '你是一位全栈工程师，精通前后端技术。你能从数据库设计到 UI 实现提供端到端的解决方案。请根据用户的需求选择最合适的技术方案，并解释技术选型的理由。' },
  { templateId: 7, name: '中英翻译', icon: 'languages', categoryId: 3, categoryName: '翻译语言', categoryIcon: 'languages', description: '中英文互译，保持语境和风格', roleInfo: '你是一位专业的翻译人员，精通中文和英文。你不仅能进行字面翻译，还能根据上下文保持原文的语气、风格和文化适配。请提供准确的翻译，并在必要时附上翻译说明。' },
  { templateId: 8, name: '多语言翻译', icon: 'languages', categoryId: 3, categoryName: '翻译语言', categoryIcon: 'languages', description: '支持中英日韩等多语言互译', roleInfo: '你是一位多语言翻译专家。你精通中文、英文、日文、韩文等多种语言。你能帮助用户进行多语言翻译、本地化适配和跨文化沟通。请确保翻译准确、自然，符合目标语言的表达习惯。' },
  { templateId: 9, name: '文案撰写', icon: 'pen-line', categoryId: 4, categoryName: '写作创作', categoryIcon: 'pen-line', description: '营销文案、广告语和品牌故事创作', roleInfo: '你是一位专业的文案撰稿人。你擅长撰写各种类型的文案，包括营销文案、广告语、品牌故事、产品描述等。请根据目标受众和品牌调性，创作有吸引力、有说服力的文案。' },
  { templateId: 10, name: '技术文档', icon: 'file-text', categoryId: 4, categoryName: '写作创作', categoryIcon: 'pen-line', description: 'API 文档、技术说明和教程撰写', roleInfo: '你是一位技术文档撰写专家。你擅长编写清晰、准确、易于理解的技术文档，包括 API 文档、使用指南、教程和故障排除文档。请使用简洁的语言，提供代码示例，并注意文档的结构和可读性。' },
  { templateId: 11, name: '创意写作', icon: 'pen-line', categoryId: 4, categoryName: '写作创作', categoryIcon: 'pen-line', description: '故事创作、剧本和诗歌', roleInfo: '你是一位创意写作者。你擅长创作各种类型的文学作品，包括短篇小说、剧本、诗歌等。请根据用户的需求发挥想象力，创作有深度、有感染力的作品。' },
  { templateId: 12, name: '数据分析师', icon: 'bar-chart-3', categoryId: 5, categoryName: '数据分析', categoryIcon: 'bar-chart-3', description: '数据处理、可视化和商业分析', roleInfo: '你是一位数据分析师。你擅长使用 Python（pandas、numpy、matplotlib）和 SQL 进行数据分析。你能帮助用户清洗数据、探索性分析、构建可视化报表和提取业务洞察。请提供可复现的分析代码和清晰的结论。' },
  { templateId: 13, name: '数据科学家', icon: 'bar-chart-3', categoryId: 5, categoryName: '数据分析', categoryIcon: 'bar-chart-3', description: '机器学习建模和统计推断', roleInfo: '你是一位数据科学家。你精通机器学习算法、统计建模和实验设计。你能帮助用户进行特征工程、模型选择、超参数调优和结果解释。请给出严谨的分析方法，并注意模型的解释性和可部署性。' },
  { templateId: 14, name: '学习导师', icon: 'graduation-cap', categoryId: 6, categoryName: '学习教育', categoryIcon: 'graduation-cap', description: '知识问答、概念解释和学习规划', roleInfo: '你是一位耐心、知识渊博的学习导师。你善于将复杂的概念用简单易懂的方式解释清楚。你能帮助用户理解各种学科的知识，制定学习计划，并提供学习资源建议。请用循序渐进的方式讲解，鼓励用户思考和提问。' },
  { templateId: 15, name: '学术研究', icon: 'graduation-cap', categoryId: 6, categoryName: '学习教育', categoryIcon: 'graduation-cap', description: '论文写作、文献综述和研究方法', roleInfo: '你是一位学术研究助手。你熟悉学术论文的写作规范、文献综述方法和研究设计。你能帮助用户梳理研究思路、撰写论文、润色文字和检查引用格式。请保持学术严谨性，提供有理有据的建议。' },
  { templateId: 16, name: '金融顾问', icon: 'calculator', categoryId: 7, categoryName: '金融分析', categoryIcon: 'calculator', description: '投资分析、风险评估和财务规划', roleInfo: '你是一位金融顾问。你精通投资分析、风险评估、资产配置和财务规划。你能帮助用户分析市场趋势、评估投资机会、制定财务目标。请注意：你提供的是信息和分析，不构成投资建议。用户在做出任何投资决策前应咨询持牌专业人士。' },
  { templateId: 17, name: '日常问答', icon: 'message-circle', categoryId: 1, categoryName: '通用', categoryIcon: 'bot', description: '生活常识、百科知识问答', roleInfo: '你是一个知识丰富的问答助手。你可以回答各种领域的问题，从科学技术到人文历史，从生活常识到专业领域。请提供准确、客观的信息，并注明信息来源（如有）。' },
  { templateId: 18, name: '客服助手', icon: 'heart', categoryId: 1, categoryName: '通用', categoryIcon: 'bot', description: '客户服务、问题解答和投诉处理', roleInfo: '你是一位专业的客服助手。你能礼貌、耐心地处理用户的问题和投诉。请始终以用户为中心，积极解决问题，保持专业和友善的态度。如果遇到无法解决的问题，请明确告知用户并提供替代方案。' },
  { templateId: 19, name: 'AI 研究顾问', icon: 'brain', categoryId: 6, categoryName: '学习教育', categoryIcon: 'graduation-cap', description: 'AI 技术研究、论文解读和趋势分析', roleInfo: '你是一位 AI 研究顾问。你精通机器学习、深度学习、自然语言处理、计算机视觉等 AI 领域的前沿技术。你能帮助用户解读学术论文、分析技术趋势、评估模型架构和训练策略。请提供有深度的技术见解，并引用相关研究。' },
  { templateId: 20, name: '国际资讯', icon: 'globe', categoryId: 7, categoryName: '金融分析', categoryIcon: 'calculator', description: '国际新闻摘要、跨文化沟通和时政分析', roleInfo: '你是一位国际资讯分析员。你关注全球时事、经济动态和跨文化议题。你能帮助用户理解国际事件的背景、各方立场和潜在影响。请保持客观、中立，提供多角度分析，并明确区分事实与观点。' },
]

// TemplateDetail[] (with roleInfo) — used by PersonaEditPage template selector
export const TEMPLATES: TemplateDetail[] = _fullTemplates

export const TEMPLATE_CATEGORIES: TemplateCategoryItem[] = [
  { categoryId: 1, name: '通用', icon: 'bot' },
  { categoryId: 2, name: '编程开发', icon: 'code' },
  { categoryId: 3, name: '翻译语言', icon: 'languages' },
  { categoryId: 4, name: '写作创作', icon: 'pen-line' },
  { categoryId: 5, name: '数据分析', icon: 'bar-chart-3' },
  { categoryId: 6, name: '学习教育', icon: 'graduation-cap' },
  { categoryId: 7, name: '金融分析', icon: 'calculator' },
]

/* ============================================
   Persona categories
   ============================================ */

export const PERSONA_CATEGORIES: PersonaCategoryItem[] = [
  { categoryId: 1, name: '通用', icon: 'star' },
  { categoryId: 2, name: '开发', icon: 'code' },
  { categoryId: 3, name: '语言', icon: 'globe' },
  { categoryId: 4, name: '数据', icon: 'search' },
  { categoryId: 5, name: '写作', icon: 'file-text' },
  { categoryId: 6, name: '运维', icon: 'terminal' },
]

/* ============================================
   Models
   ============================================ */

export const PERSONA_MODELS: PersonaModelItem[] = [
  { aiModelId: 0, providerName: '默认模型', displayName: '默认模型', modelName: '', icon: '' },
  { aiModelId: 1, providerName: 'DeepSeek', displayName: 'DeepSeek V3', modelName: 'deepseek-chat', icon: '/logos/deepseek.svg' },
  { aiModelId: 2, providerName: 'Claude', displayName: 'Claude 3.5 Sonnet', modelName: 'claude-3.5-sonnet', icon: '/logos/claude.svg' },
  { aiModelId: 3, providerName: 'OpenAI', displayName: 'GPT-4o', modelName: 'gpt-4o', icon: '/logos/openai.svg' },
  { aiModelId: 4, providerName: 'Qwen', displayName: 'Qwen Max', modelName: 'qwen-max', icon: '/logos/bailian.svg' },
  { aiModelId: 5, providerName: 'Gemini', displayName: 'Gemini 2.0 Flash', modelName: 'gemini-2.0-flash', icon: '' },
]

/* ============================================
   MCP endpoints
   ============================================ */

export const PERSONA_MCPS: PersonaMcpItem[] = [
  { mcpId: 1, name: 'filesystem', icon: 'file-text', description: '文件系统读写与管理操作' },
  { mcpId: 2, name: 'brave-search', icon: 'search', description: 'Brave 网页搜索引擎' },
  { mcpId: 3, name: 'github', icon: 'code', description: 'GitHub 仓库、PR、Issue API 交互' },
  { mcpId: 4, name: 'postgres', icon: 'terminal', description: 'PostgreSQL 数据库查询与管理' },
  { mcpId: 5, name: 'redis', icon: 'terminal', description: 'Redis 缓存读写与监控' },
  { mcpId: 6, name: 'slack', icon: 'bot', description: 'Slack 消息发送与频道管理' },
  { mcpId: 7, name: 'notion', icon: 'file-text', description: 'Notion 页面与数据库操作' },
  { mcpId: 8, name: 'jira', icon: 'wrench', description: 'Jira 工单查询与状态更新' },
  { mcpId: 9, name: 'docker', icon: 'terminal', description: 'Docker 容器与镜像管理' },
  { mcpId: 10, name: 'kubernetes', icon: 'terminal', description: 'K8s 集群 Pod、Service 操作' },
  { mcpId: 11, name: 'aws-s3', icon: 'file-text', description: 'AWS S3 对象存储操作' },
  { mcpId: 12, name: 'stripe', icon: 'star', description: 'Stripe 支付与账单查询' },
  { mcpId: 13, name: 'sendgrid', icon: 'globe', description: 'SendGrid 邮件发送服务' },
  { mcpId: 14, name: 'twilio', icon: 'globe', description: 'Twilio 短信与语音服务' },
  { mcpId: 15, name: 'google-drive', icon: 'file-text', description: 'Google Drive 文件上传下载' },
  { mcpId: 16, name: 'mysql', icon: 'terminal', description: 'MySQL 数据库查询与管理' },
]

/* ============================================
   Skills
   ============================================ */

export const PERSONA_SKILLS: PersonaSkillItem[] = [
  { userSkillId: 1, name: 'code-review', description: '代码审查与安全漏洞检测', icon: 'code' },
  { userSkillId: 2, name: 'docker-helper', description: 'Docker 容器生命周期管理', icon: 'wrench' },
  { userSkillId: 3, name: 'data-analyzer', description: 'CSV/JSON 数据导入与图表生成', icon: 'search' },
  { userSkillId: 4, name: 'pdf-parser', description: 'PDF 文档解析与内容提取', icon: 'file-text' },
  { userSkillId: 5, name: 'image-optimizer', description: '图片压缩与格式转换', icon: 'star' },
  { userSkillId: 6, name: 'log-analyzer', description: '日志聚合分析与异常检测', icon: 'terminal' },
  { userSkillId: 7, name: 'api-tester', description: 'REST/GraphQL API 接口测试', icon: 'globe' },
  { userSkillId: 8, name: 'sql-formatter', description: 'SQL 语句格式化和优化建议', icon: 'terminal' },
  { userSkillId: 9, name: 'markdown-lint', description: 'Markdown 语法检查与自动修复', icon: 'file-text' },
  { userSkillId: 10, name: 'git-helper', description: 'Git 工作流辅助与提交信息生成', icon: 'code' },
  { userSkillId: 11, name: 'i18n-translator', description: '多语言 i18n 文件翻译', icon: 'globe' },
  { userSkillId: 12, name: 'regex-builder', description: '正则表达式可视化构建', icon: 'code' },
  { userSkillId: 13, name: 'openapi-gen', description: 'OpenAPI/Swagger 文档生成', icon: 'file-text' },
  { userSkillId: 14, name: 'db-migration', description: '数据库迁移脚本生成与校验', icon: 'terminal' },
]

/* ============================================
   ID counter for new personas
   ============================================ */

let _nextPersonaId = (() => {
  let max = 0
  for (const p of PERSONAS) {
    if (p.personaId > max) max = p.personaId
  }
  return max + 1
})()

/* ============================================
   Resolve helpers
   ============================================ */

function resolveModelName(aiModelId: number): string {
  const m = PERSONA_MODELS.find(x => x.aiModelId === aiModelId)
  if (!m || m.aiModelId === 0) return '默认模型'
  return `${m.providerName} - ${m.displayName}`
}

function resolveCategoryName(categoryId: number): string {
  const c = PERSONA_CATEGORIES.find(x => x.categoryId === categoryId)
  return c ? c.name : '未知'
}

function resolveCategoryIcon(categoryId: number): string {
  const c = PERSONA_CATEGORIES.find(x => x.categoryId === categoryId)
  return c ? c.icon : 'bot'
}

function resolveMcpIdsToNames(ids: number[]): { id: number; name: string; icon: string }[] {
  return ids.map(id => {
    const m = PERSONA_MCPS.find(x => x.mcpId === id)
    return m ? { id: m.mcpId, name: m.name, icon: m.icon } : { id, name: `mcp-${id}`, icon: 'wrench' }
  })
}

function resolveSkillIdsToNames(ids: number[]): { id: number; name: string; icon: string }[] {
  return ids.map(id => {
    const s = PERSONA_SKILLS.find(x => x.userSkillId === id)
    return s ? { id: s.userSkillId, name: s.name, icon: s.icon } : { id, name: `skill-${id}`, icon: 'wrench' }
  })
}

/* ============================================
   Async mock functions
   ============================================ */

export async function getPersonas(): Promise<PersonaListItem[]> {
  await listDelay()
  return [...PERSONAS]
}

export async function getPersonaDetail(id: number): Promise<PersonaDetail | undefined> {
  await itemDelay()
  return PERSONA_DETAILS[id] ? { ...PERSONA_DETAILS[id] } : undefined
}

export async function createPersona(req: CreatePersonaRequest): Promise<PersonaListItem> {
  await mutationDelay()

  const id = _nextPersonaId++
  const modelName = resolveModelName(req.aiModelId ?? 0)
  const categoryName = resolveCategoryName(req.categoryId)
  const mcpIds = req.mcpIds ?? []
  const skillIds = req.skillIds ?? []

  const listItem: PersonaListItem = {
    personaId: id,
    name: req.name,
    icon: req.icon ?? 'bot',
    categoryName,
    modelName,
    roleInfo: req.roleInfo,
    mcpNames: mcpIds.map(mid => PERSONA_MCPS.find(x => x.mcpId === mid)?.name ?? `mcp-${mid}`),
    skillNames: skillIds.map(sid => PERSONA_SKILLS.find(x => x.userSkillId === sid)?.name ?? `skill-${sid}`),
  }

  const detail: PersonaDetail = {
    personaId: id,
    name: req.name,
    icon: req.icon ?? 'bot',
    roleInfo: req.roleInfo,
    categoryId: req.categoryId,
    categoryName,
    categoryIcon: resolveCategoryIcon(req.categoryId),
    mcpIds,
    mcpNames: resolveMcpIdsToNames(mcpIds),
    skillIds,
    skillNames: resolveSkillIdsToNames(skillIds),
    aiModelId: req.aiModelId ?? 0,
    modelName,
    modelIcon: PERSONA_MODELS.find(x => x.aiModelId === (req.aiModelId ?? 0))?.icon ?? '',
    created: new Date().toISOString(),
    updated: new Date().toISOString(),
  }

  PERSONAS.push(listItem)
  PERSONA_DETAILS[id] = detail

  return listItem
}

export async function updatePersona(id: number, req: Partial<CreatePersonaRequest>): Promise<void> {
  await mutationDelay()

  const listIdx = PERSONAS.findIndex(p => p.personaId === id)
  const detail = PERSONA_DETAILS[id]

  if (!detail && listIdx < 0) return

  // Resolve updated values, falling back to existing detail values
  const newCategoryId = req.categoryId ?? detail?.categoryId ?? 0
  const newAiModelId = req.aiModelId ?? detail?.aiModelId ?? 0
  const newMcpIds = req.mcpIds ?? detail?.mcpIds ?? []
  const newSkillIds = req.skillIds ?? detail?.skillIds ?? []
  const categoryName = resolveCategoryName(newCategoryId)
  const modelName = resolveModelName(newAiModelId)

  // Update list item
  if (listIdx >= 0) {
    PERSONAS[listIdx] = {
      ...PERSONAS[listIdx],
      name: req.name ?? PERSONAS[listIdx].name,
      icon: req.icon ?? PERSONAS[listIdx].icon,
      categoryName,
      modelName,
      roleInfo: req.roleInfo ?? PERSONAS[listIdx].roleInfo,
      mcpNames: newMcpIds.map(mid => PERSONA_MCPS.find(x => x.mcpId === mid)?.name ?? `mcp-${mid}`),
      skillNames: newSkillIds.map(sid => PERSONA_SKILLS.find(x => x.userSkillId === sid)?.name ?? `skill-${sid}`),
    }
  }

  // Update detail
  if (detail) {
    PERSONA_DETAILS[id] = {
      ...detail,
      name: req.name ?? detail.name,
      icon: req.icon ?? detail.icon,
      roleInfo: req.roleInfo ?? detail.roleInfo,
      categoryId: newCategoryId,
      categoryName,
      categoryIcon: resolveCategoryIcon(newCategoryId),
      mcpIds: newMcpIds,
      mcpNames: resolveMcpIdsToNames(newMcpIds),
      skillIds: newSkillIds,
      skillNames: resolveSkillIdsToNames(newSkillIds),
      aiModelId: newAiModelId,
      modelName,
      modelIcon: PERSONA_MODELS.find(x => x.aiModelId === newAiModelId)?.icon ?? '',
      updated: new Date().toISOString(),
    }
  }
}

export async function deletePersona(id: number): Promise<void> {
  await mutationDelay()

  const listIdx = PERSONAS.findIndex(p => p.personaId === id)
  if (listIdx >= 0) {
    PERSONAS.splice(listIdx, 1)
  }
  delete PERSONA_DETAILS[id]
}

export async function getTemplates(): Promise<TemplateItem[]> {
  await listDelay()
  return TEMPLATES.map(t => ({ ...t }))
}

export async function getTemplateDetail(id: number): Promise<TemplateDetail | undefined> {
  await itemDelay()
  const t = _fullTemplates.find(x => x.templateId === id)
  return t ? { ...t } : undefined
}

export async function getTemplateCategories(): Promise<TemplateCategoryItem[]> {
  await listDelay()
  return TEMPLATE_CATEGORIES.map(c => ({ ...c }))
}

export async function getPersonaCategories(): Promise<PersonaCategoryItem[]> {
  await listDelay()
  return PERSONA_CATEGORIES.map(c => ({ ...c }))
}

export async function getModels(): Promise<PersonaModelItem[]> {
  await itemDelay()
  return PERSONA_MODELS.map(m => ({ ...m }))
}

export async function getMcps(): Promise<PersonaMcpItem[]> {
  await itemDelay()
  return PERSONA_MCPS.map(m => ({ ...m }))
}

export async function getSkills(): Promise<PersonaSkillItem[]> {
  await itemDelay()
  return PERSONA_SKILLS.map(s => ({ ...s }))
}
