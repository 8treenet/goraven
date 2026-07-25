-- ============================================================
-- Mock Data: AI Models, MCP Endpoints, System Skills
-- 表: ai_model, mcp_endpoint, system_skill
-- 数据库: SQLite
-- 依赖: 无
-- ============================================================

-- ============================================================
-- ai_model 表
-- 14个模型，涵盖主流 Provider
-- ============================================================

INSERT OR IGNORE INTO ai_model (aiModelId, providerDisplayName, providerId, displayName, modelName, icon, apiKey, baseUrl, extraFields, proxyUrl, contextLen, isDefault, isFlash, isVisual, status, remark, created, updated) VALUES

-- 1. DeepSeek V3 (默认模型)
(1, 'DeepSeek', 'deepseek', 'deepseek-v4-pro','deepseek-v4-pro', '/logos/deepseek.svg', 'sk-deepseek-mock-key-0000000000000000000000000000001', 'https://api.deepseek.com/v1', '', '', 195, 1, 0, 0, 1, 'DeepSeek V3 主力模型，性价比高', datetime('now'), datetime('now')),

-- 2. DeepSeek R1 (Flash 模型)
(2, 'DeepSeek', 'deepseek', 'deepseek-v4-flash','deepseek-v4-flash', '/logos/deepseek.svg', 'sk-deepseek-mock-key-0000000000000000000000000000001', 'https://api.deepseek.com/v1', '{"thinking":{"type":"enabled"}}', '', 195, 0, 1, 0, 1, 'DeepSeek R1 推理模型，用于对话压缩和子 agent', datetime('now'), datetime('now')),

-- 3. OpenAI GPT 5.5
(3, 'OpenAI', 'openai', 'GPT 5.5', 'GPT 5.5', '/logos/openai.svg', 'sk-openai-mock-key-000000000000000000000000000000002', 'https://api.openai.com/v1', '', '', 195, 0, 0, 0, 1, 'OpenAI 多模态旗舰', datetime('now'), datetime('now')),

-- 4. OpenAI GPT 5.1 (视觉模型)
(4, 'OpenAI', 'openai', 'GPT 5.1', 'GPT 5.1', '/logos/openai.svg', 'sk-openai-mock-key-000000000000000000000000000000002', 'https://api.openai.com/v1', '', '', 195, 0, 0, 1, 1, '轻量多模态模型，用于图片理解', datetime('now'), datetime('now')),

-- 5. Claude Sonnet 4
(5, 'Anthropic','claude', 'claude-sonnet-4.7', 'claude-sonnet-4.7', '/logos/claude.svg', 'sk-ant-mock-key-000000000000000000000000000000000003', 'https://api.anthropic.com/v1', '', '', 200, 0, 0, 0, 1, 'Claude Sonnet 4，代码能力强', datetime('now'), datetime('now')),

-- 6. Claude Opus 4
(6, 'Anthropic','claude', 'claude-opus-4.7', 'claude-opus-4.7', '/logos/claude.svg', 'sk-ant-mock-key-000000000000000000000000000000000003', 'https://api.anthropic.com/v1', '{"thinking":{"type":"enabled"}}', '', 200, 0, 0, 0, 1, 'Claude Opus 4，最强推理', datetime('now'), datetime('now')),

-- 7. 智谱 GLM
(7, 'Zhipu','glm', 'glm-5.1', 'glm-5.1', '/logos/zhipu.svg', 'zhipu-mock-key-00000000000000000000000000000000000004', 'https://open.bigmodel.cn/api/paas/v4', '', '', 128, 0, 0, 0, 1, 'GLM 国产旗舰', datetime('now'), datetime('now')),

-- 8. 通义千问 Qwen-Max
(8, 'Alibaba', 'bailian', 'qwen-3.7max', 'qwen-3.7max', '/logos/bailian.svg', 'bailian-mock-key-00000000000000000000000000000000005', 'https://dashscope.aliyuncs.com/compatible-mode/v1', '', '', 128, 0, 0, 0, 1, '通义千问 Max', datetime('now'), datetime('now')),

-- 9. MiniMax
(9, 'MiniMax', 'minimax', 'MiniMax-M2.7', 'MiniMax-M2.7', '/logos/minimax.svg', 'minimax-mock-key-00000000000000000000000000000000006', 'https://api.minimax.chat/v1', '', '', 128, 0, 0, 0, 1, 'MiniMax abab6.5s', datetime('now'), datetime('now')),

-- 10. Ollama 本地 (Qwen2.5)
(10, 'Ollama Local', 'ollama','qwen2.5:14b', 'qwen2.5:14b', '/logos/ollama.svg', '', 'http://localhost:11434/v1', '', '', 32, 0, 0, 0, 1, '本地 Ollama Qwen2.5 14B，无需 API Key', datetime('now'), datetime('now')),

-- 11. Ollama 本地 (CodeGemma)
(11, 'Ollama Local', 'ollama','codegemma:7b', 'codegemma:7b', '/logos/ollama.svg', '', 'http://localhost:11434/v1', '', '', 32, 0, 0, 0, 1, '本地代码补全模型', datetime('now'), datetime('now')),

-- 12. OpenRouter (聚合)
(12, 'OpenRouter', 'open_router', 'anthropic/claude-opus-4.8', 'anthropic/claude-opus-4.8', '/logos/claude.svg', 'or-mock-key-0000000000000000000000000000000000000007', 'https://openrouter.ai/api/v1', '', 'http://127.0.0.1:7890', 195, 0, 0, 0, 1, '通过 OpenRouter 访问 GPT-4o，走代理', datetime('now'), datetime('now')),

-- 13. Gemini 2.5 Pro
(13, 'Google', 'gemini', 'Gemini 3.5 Flash', 'Gemini 3.5 Flash', '/logos/gemini.svg', 'gemini-mock-key-0000000000000000000000000000000000008', 'https://generativelanguage.googleapis.com/v1beta', '', '', 1000, 0, 0, 0, 1, 'Gemini 超长上下文(1M)', datetime('now'), datetime('now'));

-- ============================================================
-- mcp_endpoint 表
-- 12个 MCP 服务端点，覆盖：开发工具、财务数据、写作编辑、教育研究、办公协作
-- Transport: Stdio / SSE / StreamableHttp
-- ============================================================

INSERT OR IGNORE INTO mcp_endpoint (mcpId, name, displayName, icon, description, transport, httpUrl, httpHeader, proxyUrl, stdioType, stdioEnv, stdioArgs, status, healthLatency, healthCheckedAt, remark, created, updated) VALUES

-- === 开发工具 ===

-- 1. Filesystem (Stdio via npx)
(1, 'filesystem', '文件系统', 'folder', '读写服务器文件系统，支持文件的创建、读取、编辑、删除、搜索等操作', 'Stdio', '', '', '', 'npx', '["NODE_ENV=production"]', '["@anthropic/mcp-filesystem","/raven/user_space"]', 1, 45, datetime('now'), '文件系统操作，限制在 user_space 目录', datetime('now'), datetime('now')),

-- 2. PostgreSQL (Stdio via npx)
(2, 'postgres', 'PostgreSQL', 'database', '连接 PostgreSQL 数据库，支持 SQL 查询、Schema 浏览、表结构分析', 'Stdio', '', '', '', 'npx', '["PGHOST=localhost","PGPORT=5432","PGDATABASE=raven"]', '["@anthropic/mcp-postgres"]', 1, 12, datetime('now'), '数据库查询工具', datetime('now'), datetime('now')),

-- 3. GitHub (Stdio via npx)
(3, 'github', 'GitHub', 'git-branch', 'GitHub API 集成：仓库管理、Issue、PR、文件操作、搜索代码', 'Stdio', '', '', '', 'npx', '["GITHUB_TOKEN=ghp_mock00000000000000000000000000000009"]', '["@anthropic/mcp-github"]', 1, 78, datetime('now'), '需要配置 GITHUB_TOKEN 环境变量', datetime('now'), datetime('now')),

-- 4. Brave Search (Stdio via npx)
(4, 'brave-search', 'Brave 搜索', 'search', 'Brave Search API：网页搜索、新闻搜索，返回结构化结果', 'Stdio', '', '', '', 'npx', '["BRAVE_API_KEY=bsa-mock0000000000000000000000000000000010"]', '["@anthropic/mcp-brave-search"]', 1, 230, datetime('now'), '需要 Brave Search API Key', datetime('now'), datetime('now')),

-- === 财务数据 ===

-- 5. 财务数据 (SSE 远程服务)
(5, 'financial-data', '财务数据服务', 'landmark', '全球金融市场数据：股票行情、外汇汇率、企业财报、宏观经济指标', 'SSE', 'https://finance-mcp.example.com/sse', '{"Authorization":"Bearer fin-mock0000000000000000000000000000000011"}', '', '', '', '', 1, 156, datetime('now'), '覆盖美股、A股、港股、外汇、加密货币行情', datetime('now'), datetime('now')),

-- 6. 税务计算 (SSE 远程服务)
(6, 'tax-calculator', '税务计算器', 'calculator', '中国个人所得税、增值税、企业所得税计算，支持最新税法规则', 'SSE', 'https://tax-mcp.example.com/sse', '{"Authorization":"Bearer tax-mock0000000000000000000000000000000012"}', '', '', '', '', 1, 89, datetime('now'), '依据2026年最新税法', datetime('now'), datetime('now')),

-- === 写作编辑 ===

-- 7. 文档编辑 (StreamableHttp)
(7, 'document-editor', '文档编辑器', 'pen-line', '在线文档创建与编辑：支持 Markdown、富文本、LaTeX 数学公式、表格', 'StreamableHttp', 'https://docs-mcp.example.com/mcp', '{"Authorization":"Bearer doc-mock0000000000000000000000000000000013"}', '', '', '', '', 1, 67, datetime('now'), '协作编辑、版本对比、导出 PDF/DOCX', datetime('now'), datetime('now')),

-- 8. 翻译服务 (SSE)
(8, 'translation', '多语言翻译', 'languages', '支持中英日韩法德等 30+ 语言的翻译，保持原文格式和 Markdown', 'SSE', 'https://translate-mcp.example.com/sse', '{"Authorization":"Bearer tr-mock00000000000000000000000000000000014"}', '', '', '', '', 1, 112, datetime('now'), '专业翻译引擎，术语库可定制', datetime('now'), datetime('now')),

-- === 教育研究 ===

-- 9. 学术论文检索 (Stdio via npx)
(9, 'arxiv-search', '学术论文检索', 'book-marked', '检索 arXiv、知网、PubMed 等学术数据库，获取论文摘要、引用、全文链接', 'Stdio', '', '', '', 'npx', '[]', '["@anthropic/mcp-arxiv"]', 1, 198, datetime('now'), '支持中英文论文检索，按引用量、日期排序', datetime('now'), datetime('now')),

-- 10. 知识库管理 (Stdio via npx)
(10, 'notion', 'Notion 知识库', 'library', 'Notion 工作区集成：页面创建、数据库查询、知识库管理、模板应用', 'Stdio', '', '', '', 'npx', '["NOTION_TOKEN=ntn_mock00000000000000000000000000000000015"]', '["@anthropic/mcp-notion"]', 1, 145, datetime('now'), '需要 Notion Integration Token', datetime('now'), datetime('now')),

-- === 办公协作 ===

-- 11. 天气服务 (SSE)
(11, 'weather', '天气服务', 'cloud-sun', '查询全球城市实时天气和未来15天预报，包含空气质量、紫外线指数', 'SSE', 'https://weather-mcp.example.com/sse', '{"Authorization":"Bearer wtr-mock0000000000000000000000000000000011"}', '', '', '', '', 1, 120, datetime('now'), '第三方天气 MCP 服务', datetime('now'), datetime('now')),

-- 12. 邮箱集成 (StreamableHttp)
(12, 'email', '邮件助手', 'mail', '邮件发送与读取：支持 SMTP/IMAP，模板邮件、批量发送、邮件追踪', 'StreamableHttp', 'https://email-mcp.example.com/mcp', '{"Authorization":"Bearer eml-mock0000000000000000000000000000000016"}', '', '', '', '', 1, 95, datetime('now'), '支持 Gmail、Outlook、企业邮箱', datetime('now'), datetime('now')),

-- === 可用的 Stdio MCP 服务（通过 npx 启动，已验证通过健康检测） ===

-- 13. 知识图谱记忆 (Stdio via npx)
(13, 'memory', '知识图谱记忆', 'brain-circuit', '基于知识图谱的持久化记忆系统，支持实体、关系、观测的创建与检索', 'Stdio', '', '', '', 'npx', '[]', '["@modelcontextprotocol/server-memory"]', 1, 45, datetime('now'), '无需配置，开箱即用', datetime('now'), datetime('now')),

-- 14. 深度思考 (Stdio via npx)
(14, 'sequential-thinking', '深度思考', 'brain', '结构化深度推理工具，支持逐步思考、假设验证、修正与分支推理', 'Stdio', '', '', '', 'npx', '[]', '["@modelcontextprotocol/server-sequential-thinking"]', 1, 60, datetime('now'), '适合需要复杂推理的任务', datetime('now'), datetime('now')),

-- 15. 时间工具 (Stdio via npx)
(15, 'time-tool', '时间工具', 'clock', '时间处理工具集：获取当前时间、时区转换、时间计算、时间对比', 'Stdio', '', '', '', 'npx', '[]', '["@theo.foobar/mcp-time"]', 1, 38, datetime('now'), '零配置时间工具', datetime('now'), datetime('now')),

-- 16. MCP 调试工具 (Stdio via npx)
(16, 'mcp-everything', 'MCP 调试工具', 'wrench', 'MCP 协议参考服务器，提供 echo、文件压缩、资源链接等测试和调试功能', 'Stdio', '', '', '', 'npx', '[]', '["@modelcontextprotocol/server-everything"]', 1, 52, datetime('now'), 'MCP 开发调试用', datetime('now'), datetime('now'));


-- ============================================================
-- system_skill 表
-- 5个系统内置技能，覆盖开发、财务、写作、教育
-- ============================================================

INSERT OR IGNORE INTO system_skill (skillId, name, description, content, status, created, updated) VALUES

-- 编程开发类
(11, 'raven-code-review', '代码审查技能：分析代码质量、安全性、性能问题', '---
name: raven-code-review
description: 对代码变更进行多维度审查，包括正确性、安全性、性能和可维护性
---

# 代码审查

对用户提供的代码进行多维度审查：
1. 正确性：逻辑错误、边界条件
2. 安全性：SQL注入、XSS、敏感信息泄露
3. 性能：N+1查询、不必要的循环
4. 可维护性：命名、注释、模块化
', 0, datetime('now'), datetime('now')),

-- 财务类
(12, 'raven-finance-report', '财务报表分析技能：自动生成财务报告和比率分析', '---
name: raven-finance-report
description: 分析财务报表，计算关键财务指标，生成可视化分析报告
---

# 财务报表分析

- 解析利润表、资产负债表、现金流量表
- 自动计算 ROE、ROA、毛利率、净利率、流动比率、速动比率
- 杜邦分析框架
- 行业对标比较
- 生成 Markdown + 图表报告

注意：区分事实和观点，不提供投资建议
', 0, datetime('now'), datetime('now')),

-- 写作类
(13, 'raven-writing-assistant', '写作助手：支持多种文体，优化表达和结构', '---
name: raven-writing-assistant
description: 多文体写作辅助，包含语法纠正、风格优化、结构建议
---

# 写作助手

支持文体：
- 商务信函与邮件
- 技术文档与白皮书
- 学术论文（APA/MLA/GB格式）
- 创意写作（小说、散文、诗歌）
- 自媒体文案（公众号、小红书、推特）

功能：
- 语法与拼写检查
- 文章结构优化建议
- 语气与风格调整
- 阅读难度评估（Flesch-Kincaid）
', 0, datetime('now'), datetime('now')),

-- 教育类
(14, 'raven-lesson-planner', '教案设计助手：根据知识点自动生成教案和习题', '---
name: raven-lesson-planner
description: 帮助教师快速设计教案、生成习题和测验
---

# 教案设计助手

功能：
- 根据知识点自动生成教学目标（Bloom分类法）
- 设计教学过程：导入 → 讲解 → 练习 → 总结
- 生成分层习题：基础题、提高题、拓展题
- 制作 PPT 大纲和讲义
- 设计项目式学习（PBL）任务

支持学段：小学、初中、高中、大学
', 0, datetime('now'), datetime('now')),

-- 通用工具类
(15, 'raven-data-visualizer', '数据可视化：自动选择图表类型并生成代码', '---
name: raven-data-visualizer
description: 根据数据特征自动推荐最佳图表类型，生成 ECharts/Recharts/Matplotlib 代码
---

# 数据可视化

- 自动分析数据结构和分布
- 推荐最佳图表类型（柱状图、折线图、散点图、热力图等）
- 生成可运行的图表代码
- 支持 ECharts、Recharts、Matplotlib、Plotly

图表选择原则：
- 趋势 → 折线图
- 比较 → 柱状图
- 占比 → 饼图/环形图
- 分布 → 直方图/箱线图
- 关系 → 散点图/气泡图
', 0, datetime('now'), datetime('now'));
