-- ============================================================
-- Mock Data: Persona Templates, User Personas, Persona Tools,
--            Skill Market, User Skills
-- 表: persona_template, user_persona, persona_tool,
--     skill_market, user_skill
-- 数据库: SQLite
-- 依赖: 01_users_and_auth.sql (user 表)
-- 所有用户数据均归属 @user1 = '90a431bee756432492c134f510bad949'
-- ============================================================

-- ============================================================
-- persona_template 表 (角色模板)
-- 12个模板，覆盖：通用、编程、财务、写作、教育、数据分析
-- categoryId: 1=通用 2=编程开发 3=翻译语言 4=写作创作 5=数据分析 6=学习教育
-- ============================================================

INSERT OR IGNORE INTO persona_template (templateId, name, description, icon, roleInfo, categoryId, usageCount, sortOrder, deleted, created, updated) VALUES

-- 通用
(1, '通用助手', '全能的 AI 对话助手，适合日常问答和任务辅助', 'bot', '你是一个全能的 AI 助手，能够帮助用户解决各种问题。回答应清晰、准确、有条理。如果问题超出你的能力范围，诚实地告知用户。', 1, 328, 1, 0, datetime('now'), datetime('now')),

-- 编程开发
(2, 'Go 后端专家', '精通 Go 语言的微服务架构师，擅长 Freedom 框架和 GORM', 'code', '你是一位资深 Go 后端开发专家。

技术栈：
- Go 1.22+
- Freedom 框架（依赖注入、MVC）
- GORM ORM
- Redis、MySQL、PostgreSQL
- Docker、Kubernetes

编码原则：
- 遵循 Go 惯用写法
- 优先使用标准库
- 错误处理不遗漏', 2, 189, 2, 0, datetime('now'), datetime('now')),

(3, '前端开发', 'React + TypeScript 前端开发专家', 'palette', '你是一位前端开发专家。

技术栈：
- React 18 + TypeScript
- Vite 构建
- Tailwind CSS v4
- shadcn/ui 组件库
- Zustand 状态管理

设计原则：
- 组件单一职责
- 保持可访问性 (a11y)', 2, 145, 3, 0, datetime('now'), datetime('now')),

-- 翻译语言
(4, '中英翻译', '专业中英文双向翻译', 'languages', '你是一位专业中英翻译。

要求：
- 保持原文的语气和风格
- 技术文档保留术语
- 文学作品注重文采
- Markdown 格式保持不变', 3, 210, 1, 0, datetime('now'), datetime('now')),

-- 写作创作
(5, '技术文档写作', '技术博客、API 文档、README 写作专家', 'pen-line', '你是一位技术文档写作专家。

文档类型：
- README 文档
- API 接口文档
- 技术博客文章
- 变更日志 (CHANGELOG)

风格：简洁明了，代码示例优先，遵循 Google Technical Writing 规范', 4, 134, 1, 0, datetime('now'), datetime('now')),

(6, '创意文案', '市场营销文案和创意写作', 'sparkles', '你是一位创意文案写手。

擅长：
- 产品落地页文案
- 社交媒体推文
- 品牌故事
- 广告标语

风格可调：严谨专业（B2B）/ 轻松活泼（B2C）/ 极简克制（Apple 风）', 4, 88, 2, 0, datetime('now'), datetime('now')),

-- 数据分析
(7, 'SQL 数据分析', 'SQL 查询、数据建模、性能优化专家', 'bar-chart-3', '你是一位 SQL 数据分析专家。

技能：
- 复杂查询（窗口函数、CTE、子查询）
- 数据库设计（ER 图、范式化）
- 性能调优（索引策略、查询计划解读）
- 数据清洗和 ETL', 5, 67, 1, 0, datetime('now'), datetime('now')),

(8, '财务分析', '财务报表分析、估值建模、全面预算管理', 'calculator', '你是一位财务分析专家。

擅长：
- 财务报表分析（利润表、资产负债表、现金流量表）
- DCF 估值模型与可比公司分析
- 比率分析（ROE、ROA、PE、PB、杜邦分析）
- 全面预算编制与执行分析
- 税务筹划（个人所得税、增值税、企业所得税）

注意：区分事实和观点，标注数据来源，不提供投资建议', 5, 45, 2, 0, datetime('now'), datetime('now')),

-- 学习教育
(9, '编程导师', '面向初学者的编程教学导师', 'graduation-cap', '你是一位耐心的编程导师，面向零基础或初学者。

教学原则：
- 用生活类比解释抽象概念
- 先给可运行的最小示例
- 循序渐进增加复杂度
- 鼓励动手实践
- 苏格拉底式提问引导思考', 6, 112, 1, 0, datetime('now'), datetime('now')),

(10, '教案设计师', '帮助教师设计教案、习题和教学活动', 'book-open', '你是一位专业的教案设计师。

服务：
- 根据知识点生成教学目标（Bloom 分类法）
- 设计教学过程：导入→讲解→练习→总结
- 生成分层习题（基础/提高/拓展）
- 设计项目式学习（PBL）方案
- 学情分析与个性化教学建议

支持学段：小学、初中、高中、大学', 6, 78, 2, 0, datetime('now'), datetime('now')),

(11, '学术写作导师', '论文写作、研究方法、学术规范指导', 'book-marked', '你是一位学术写作导师。

擅长：
- 论文结构设计（摘要、引言、方法、结果、讨论）
- 文献综述撰写指导
- 研究方法选择建议
- APA/MLA/GB/T 引用格式
- 学术英语润色
- 查重与降重策略', 6, 65, 3, 0, datetime('now'), datetime('now')),

(12, '面试辅导', '技术面试模拟和职业发展规划', 'briefcase', '你是一位职业发展顾问。

服务：
- 模拟技术面试（算法、系统设计、行为面试）
- 简历审核和优化建议
- 职业规划与技能树梳理
- 薪资谈判策略

风格：客观诚实，提供具体改进建议', 6, 90, 4, 0, datetime('now'), datetime('now'));

-- ============================================================
-- user_persona 表 (用户自定义角色，全部归属 @user1)
-- 7个角色覆盖不同场景
-- ============================================================

INSERT OR IGNORE INTO user_persona (personaId, userId, name, icon, roleInfo, categoryId, aiModelId, deleted, created, updated) VALUES

-- 通用助手
(1, '90a431bee756432492c134f510bad949', '我的通用助手', 'bot', '你是我专属的通用助手，了解我的工作习惯和偏好。我喜欢简洁直接的回答，不要啰嗦。工作环境是 macOS + VSCode。', 1, 1, 0, datetime('now'), datetime('now')),

-- 编程开发
(2, '90a431bee756432492c134f510bad949', '代码审查员', 'shield-check', '你是我的代码审查员。在审查代码时，按以下优先级：安全性 > 正确性 > 性能 > 可读性。使用中文给出审查意见，代码块保留英文。', 2, 5, 0, datetime('now'), datetime('now')),

-- 财务
(3, '90a431bee756432492c134f510bad949', '财务分析助手', 'calculator', '你是我的财务分析助手。帮我分析财务报表、计算关键指标、编制预算。使用 DCF 模型和杜邦分析框架，所有数据标注来源。', 5, 6, 0, datetime('now', '-14 days'), datetime('now')),

-- 写作
(4, '90a431bee756432492c134f510bad949', '写作创作助手', 'pen-line', '你是我的写作助手。我经常需要写技术文档、商务邮件和公众号文章。请根据文体自动调整语气，优先准确性和简洁性。', 4, 3, 0, datetime('now', '-7 days'), datetime('now')),

-- 教育
(5, '90a431bee756432492c134f510bad949', '教案设计助手', 'book-open', '你是我的教学助手。帮我设计教案、出题、评估学生。教学对象是大学本科生，课程为计算机基础。', 6, 7, 0, datetime('now', '-10 days'), datetime('now')),

-- 数据分析
(6, '90a431bee756432492c134f510bad949', '数据分析师', 'bar-chart-3', '你是我的数据分析助手。使用 Python + Pandas 处理数据，需要输出图表时生成代码。数据以 CSV/Excel 格式提供。', 5, 8, 0, datetime('now', '-3 days'), datetime('now')),

-- DevOps
(7, '90a431bee756432492c134f510bad949', 'DevOps 工程师', 'terminal', '你是我的 DevOps 工程师。负责 Docker/K8s 部署、CI/CD 流水线设计、监控告警配置。请给出完整的 YAML 配置示例。', 2, 1, 0, datetime('now', '-21 days'), datetime('now'));

-- ============================================================
-- persona_tool 表 (角色关联的工具：MCP 和 Skill)
-- 全部归属 @user1
-- ============================================================

INSERT OR IGNORE INTO persona_tool (personaId, userId, toolType, toolId, created, updated) VALUES

-- 通用助手: 文件系统 + 搜索 + 天气
(1, '90a431bee756432492c134f510bad949', 'mcp', 1, datetime('now'), datetime('now')),
(1, '90a431bee756432492c134f510bad949', 'mcp', 4, datetime('now'), datetime('now')),
(1, '90a431bee756432492c134f510bad949', 'mcp', 11, datetime('now'), datetime('now')),

-- 代码审查员: 文件系统 + GitHub + PostgreSQL
(2, '90a431bee756432492c134f510bad949', 'mcp', 1, datetime('now'), datetime('now')),
(2, '90a431bee756432492c134f510bad949', 'mcp', 3, datetime('now'), datetime('now')),
(2, '90a431bee756432492c134f510bad949', 'mcp', 2, datetime('now'), datetime('now')),
(2, '90a431bee756432492c134f510bad949', 'skill', 11, datetime('now'), datetime('now')),

-- 财务分析助手: 财务数据 + 税务计算 + 文件系统 + PostgreSQL
(3, '90a431bee756432492c134f510bad949', 'mcp', 5, datetime('now'), datetime('now')),
(3, '90a431bee756432492c134f510bad949', 'mcp', 6, datetime('now'), datetime('now')),
(3, '90a431bee756432492c134f510bad949', 'mcp', 1, datetime('now'), datetime('now')),
(3, '90a431bee756432492c134f510bad949', 'mcp', 2, datetime('now'), datetime('now')),
(3, '90a431bee756432492c134f510bad949', 'skill', 12, datetime('now'), datetime('now')),

-- 写作创作助手: 文档编辑器 + 翻译 + 文件系统
(4, '90a431bee756432492c134f510bad949', 'mcp', 7, datetime('now'), datetime('now')),
(4, '90a431bee756432492c134f510bad949', 'mcp', 8, datetime('now'), datetime('now')),
(4, '90a431bee756432492c134f510bad949', 'mcp', 1, datetime('now'), datetime('now')),
(4, '90a431bee756432492c134f510bad949', 'skill', 13, datetime('now'), datetime('now')),

-- 教案设计助手: Notion + 学术论文 + 文件系统
(5, '90a431bee756432492c134f510bad949', 'mcp', 9, datetime('now'), datetime('now')),
(5, '90a431bee756432492c134f510bad949', 'mcp', 10, datetime('now'), datetime('now')),
(5, '90a431bee756432492c134f510bad949', 'mcp', 1, datetime('now'), datetime('now')),
(5, '90a431bee756432492c134f510bad949', 'skill', 14, datetime('now'), datetime('now')),

-- 数据分析师: 文件系统 + PostgreSQL + 数据可视化skill
(6, '90a431bee756432492c134f510bad949', 'mcp', 1, datetime('now'), datetime('now')),
(6, '90a431bee756432492c134f510bad949', 'mcp', 2, datetime('now'), datetime('now')),
(6, '90a431bee756432492c134f510bad949', 'skill', 6, datetime('now'), datetime('now')),
(6, '90a431bee756432492c134f510bad949', 'skill', 15, datetime('now'), datetime('now')),

-- DevOps 工程师: Docker + GitHub + 文件系统
(7, '90a431bee756432492c134f510bad949', 'mcp', 1, datetime('now'), datetime('now')),
(7, '90a431bee756432492c134f510bad949', 'mcp', 3, datetime('now'), datetime('now')),
(7, '90a431bee756432492c134f510bad949', 'skill', 7, datetime('now'), datetime('now'));

-- ============================================================
-- skill_market 表 (技能市场)
-- 15个技能，覆盖：通用、编程、财务、写作、教育、数据
-- categoryId: 1=通用 2=编程开发 3=数据与AI 4=运维部署 5=设计创意 6=商业效率 7=写作内容 8=学习教育
-- ============================================================

INSERT OR IGNORE INTO skill_market (skillId, name, description, icon, source, sourceUrl, categoryId, status, sortOrder, installedCount, remark, created, updated) VALUES

-- === 通用工具 ===
(1, 'file-organizer', '自动整理文件和文件夹，按类型、日期或规则分类移动', 'folder-tree', 'clawhub', 'https://clawhub.ai/skills/file-organizer', 1, 1, 1, 245, '用户好评最多的技能', datetime('now'), datetime('now')),
(2, 'presentation-builder', '演示文稿生成器：根据大纲自动生成精美PPT/Keynote/Google Slides', 'presentation', 'clawhub', 'https://clawhub.ai/skills/presentation-builder', 1, 1, 2, 178, '支持多种模板风格', datetime('now'), datetime('now')),

-- === 编程开发 ===
(3, 'go-project-scaffold', 'Go 项目脚手架生成器，一键创建 Freedom MVC 项目结构', 'layers', 'clawhub', 'https://clawhub.ai/skills/go-project-scaffold', 2, 1, 1, 89, '专为 Freedom 框架优化', datetime('now'), datetime('now')),
(4, 'react-component-gen', 'React 组件生成器：根据描述自动生成 TSX + Tailwind + Storybook', 'component', 'clawhub', 'https://clawhub.ai/skills/react-component-gen', 2, 1, 2, 156, '支持 shadcn/ui 风格', datetime('now'), datetime('now')),

-- === 运维部署 ===
(5, 'docker-compose-gen', 'Docker Compose 编排生成器：根据服务描述生成完整编排文件', 'container', 'clawhub', 'https://clawhub.ai/skills/docker-compose-gen', 4, 1, 1, 198, '含 Nginx/MySQL/Redis 模板', datetime('now'), datetime('now')),
(6, 'k8s-manifest-gen', 'Kubernetes 清单生成器：从应用描述生成 Deploy/Service/Ingress', 'ship', 'clawhub', 'https://clawhub.ai/skills/k8s-manifest-gen', 4, 1, 2, 45, '支持 Kustomize', datetime('now'), datetime('now')),

-- === 数据与AI ===
(7, 'data-viz-assistant', '数据分析可视化助手：自动选择最佳图表类型并生成 ECharts/Recharts 代码', 'pie-chart', 'clawhub', 'https://clawhub.ai/skills/data-viz-assistant', 3, 1, 1, 134, '支持 ECharts/Recharts/Matplotlib', datetime('now'), datetime('now')),

-- === 财务 (商业效率) ===
(8, 'financial-report-analyzer', '财务报表分析：自动生成财务比率分析、杜邦分析、趋势报告', 'landmark', 'clawhub', 'https://clawhub.ai/skills/financial-report-analyzer', 6, 1, 1, 210, '支持上市公司财报PDF解析', datetime('now'), datetime('now')),
(9, 'tax-helper', '税务助手：个人所得税、增值税、企业所得税计算与筹划建议', 'calculator', 'clawhub', 'https://clawhub.ai/skills/tax-helper', 6, 1, 2, 156, '依据2026年最新税法', datetime('now'), datetime('now')),
(10, 'budget-planner', '预算编制助手：部门预算、项目预算、全面预算编制与执行追踪', 'wallet', 'custom_upload', 'budget-planner-v2.0.0.zip', 6, 1, 3, 67, '支持零基预算和增量预算', datetime('now'), datetime('now')),

-- === 写作内容 ===
(11, 'essay-writer', '文章写作助手：论文、报告、评论、散文等各类文体写作辅助', 'pen-line', 'clawhub', 'https://clawhub.ai/skills/essay-writer', 7, 1, 1, 189, '支持中英文，含查重建议', datetime('now'), datetime('now')),
(12, 'creative-copywriter', '创意文案生成器：广告语、品牌故事、社交媒体文案、产品描述', 'sparkles', 'clawhub', 'https://clawhub.ai/skills/creative-copywriter', 7, 1, 2, 145, '支持多平台风格（小红书/公众号/推特）', datetime('now'), datetime('now')),
(13, 'business-email-composer', '商务邮件撰写：根据要点自动生成专业、得体的中英文商务邮件', 'mail', 'clawhub', 'https://clawhub.ai/skills/business-email-composer', 7, 1, 3, 167, '多语气可选（正式/半正式/轻松）', datetime('now'), datetime('now')),

-- === 学习教育 ===
(14, 'lesson-planner', '教案设计工具：根据知识点自动生成教学方案、课件大纲和分层习题', 'book-open', 'clawhub', 'https://clawhub.ai/skills/lesson-planner', 8, 1, 1, 132, '支持 Bloom 分类法目标设计', datetime('now'), datetime('now')),
(15, 'research-paper-helper', '论文助手：文献综述、研究方法建议、格式排版、学术英语润色', 'book-marked', 'clawhub', 'https://clawhub.ai/skills/research-paper-helper', 8, 1, 2, 98, '支持 APA/MLA/GB/T 格式', datetime('now'), datetime('now')),
(16, 'flashcard-gen', '闪卡生成器：从笔记自动提取知识点生成 Anki/Quizlet 记忆卡片', 'sticky-note', 'clawhub', 'https://clawhub.ai/skills/flashcard-gen', 8, 1, 3, 89, '支持间隔重复算法优化', datetime('now'), datetime('now'));

-- ============================================================
-- user_skill 表 (用户安装的技能，全部归属 @user1)
-- ============================================================

INSERT OR IGNORE INTO user_skill (userSkillId, userId, skillName, description, icon, marketSkillId, categoryId, source, installStatus, installError, created, updated) VALUES

-- 已安装
(1, '90a431bee756432492c134f510bad949', 'go-project-scaffold', 'Go 项目脚手架生成器', 'layers', 3, 2, 'market', 2, '', datetime('now'), datetime('now')),
(2, '90a431bee756432492c134f510bad949', 'docker-compose-gen', 'Docker Compose 编排生成器', 'container', 5, 4, 'market', 2, '', datetime('now'), datetime('now')),
(3, '90a431bee756432492c134f510bad949', 'financial-report-analyzer', '财务报表分析', 'landmark', 8, 6, 'market', 2, '', datetime('now'), datetime('now')),
(4, '90a431bee756432492c134f510bad949', 'tax-helper', '税务助手', 'calculator', 9, 6, 'market', 2, '', datetime('now'), datetime('now')),
(5, '90a431bee756432492c134f510bad949', 'essay-writer', '文章写作助手', 'pen-line', 11, 7, 'market', 2, '', datetime('now'), datetime('now')),
(6, '90a431bee756432492c134f510bad949', 'creative-copywriter', '创意文案生成器', 'sparkles', 12, 7, 'market', 2, '', datetime('now'), datetime('now')),
(7, '90a431bee756432492c134f510bad949', 'lesson-planner', '教案设计工具', 'book-open', 14, 8, 'market', 2, '', datetime('now'), datetime('now')),
(8, '90a431bee756432492c134f510bad949', 'research-paper-helper', '论文助手', 'book-marked', 15, 8, 'market', 2, '', datetime('now'), datetime('now')),
(9, '90a431bee756432492c134f510bad949', 'data-viz-assistant', '数据可视化助手', 'pie-chart', 7, 3, 'market', 2, '', datetime('now'), datetime('now')),
(10, '90a431bee756432492c134f510bad949', 'presentation-builder', '演示文稿生成器', 'presentation', 2, 1, 'market', 2, '', datetime('now'), datetime('now')),
(11, '90a431bee756432492c134f510bad949', 'file-organizer', '自动整理文件', 'folder-tree', 1, 1, 'market', 2, '', datetime('now'), datetime('now')),

-- 安装失败（模拟）
(12, '90a431bee756432492c134f510bad949', 'budget-planner', '预算编制助手', 'wallet', 10, 6, 'market', 3, 'pip install failed: dependency conflict with numpy>=2.0', datetime('now', '-2 days'), datetime('now')),

-- 自定义上传技能
(13, '90a431bee756432492c134f510bad949', 'custom-git-hooks', '自定义 Git Hooks 管理', 'git-branch', 0, 2, 'custom', 2, '', datetime('now', '-5 days'), datetime('now'));
