package seed

const SystemSkillGoRavenGuide = `---
name: goraven-guide
description: GoRaven 平台总索引。当用户询问与 GoRaven 平台本身相关的问题时，先调用此技能，根据关键词匹配导航到正确的详细技能。
---

# GoRaven 平台指南

GoRaven 是一款可自部署的 Agent，每位团队成员拥有独立工作空间。你在 GoRaven 中作为 AI Agent 运行。

本技能是平台知识的总索引。当用户问题涉及 GoRaven 平台本身（而非通用编程问题），你不确定该查哪个详细技能时，来这里匹配。

## 技能索引

| 技能 | 内容 |
|------|------|
| ` + "`goraven-runtime`" + ` | 你的运行环境与约束。Docker 容器 (ubuntu 22.04) 及预装工具链 (Go/Node/Python/git 等)，文件系统和终端工具清单，web_fetch 和 visual_understand 等内置能力，ReAct 运行模式和子智能体委派，上下文压缩与工具结果裁剪机制，命令安全审查规则（禁止 sudo/全局安装/写系统文件/读凭据/破坏性操作），多用户隔离约定 |
| ` + "`goraven-features`" + ` | 平台提供给用户的功能。自由模式和角色模式 (Persona)，对话中的附件上传和处理（原样存入 temp/，文档通过 goraven_doc_parse 工具读取），技能系统（市场安装/共享/Always-On），MCP 集成（Stdio/SSE/HTTP 三种传输），工作空间和 .profile 环境变量注入，团队项目（独立创建，可设置全员开放/仅成员可见），会话分享，Markdown 渲染与 SSE 流式推送，<goraven-chart> <goraven-file> <goraven-upload> <goraven-ref> 标签，系统设置项（迭代/超时/压缩/开关） |
| ` + "`goraven-user-ui`" + ` | 普通用户可访问的前端页面。侧边栏和整体布局，对话页（新建时选模型/角色/MCP/技能/附件/项目目录/Thinking，会话中消息流/Token 用量/压缩/分享/停止/后台思考/消息操作），多会话切换，仪表盘（用量统计和趋势图），工作空间（Segmented Tab 导航/上传/预览/压缩/解压/选择驱动操作条/团队项目创建与浏览），技能中心（已安装/技能市场/团队共享/Always-On），角色管理（创建/编辑 Persona/模板），个人设置（环境变量/密码/主题/头像），登录页、安装向导和分享查看页 |
| ` + "`goraven-admin-ui`" + ` | 管理员后台页面。仪表盘（全局统计），用户管理（增删改查/重置密码/启禁用），系统设置（按分组调节参数），模型管理（配置厂商/API Key/上下文长度/测试连接/权限与成员），MCP 管理（Stdio/SSE/HTTP/模板/Always-On），技能管理（全局技能和市场），角色模板，系统信息（数据库连接池/磁盘/插件） |
| ` + "`goraven-install-skill`" + ` | 技能安装流程与目录规范。用户要求安装任何技能时**必须先调用此技能** |
| ` + "`goraven-chart`" + ` | <goraven-chart> 图表标签用法。用户要求数据分析、统计对比、趋势展示时使用 |
| ` + "`goraven-automation`" + ` | 自动化任务(定时任务)创建指南。用户提出"每天/每周几点自动…"、"每隔N分钟…"、"下周一提醒我"、"两小时后执行"等定时或周期执行需求时**必须调用**。涵盖执行类型与参数校验、Requirement 自包含写法、会话上下文继承、调度语义 |

## 规则

- 一个用户问题可能触发多个技能（如同时涉及功能和界面），按需加载。
- 如果用户问题非常明确地指向某个子技能覆盖的领域，可直接加载该子技能，无需先经过本索引。
- 用户问"在哪里点/怎么找到某个按钮/某个页面怎么进"这类纯页面导航问题，直接加载 ` + "`goraven-user-ui`" + `，无需经过本索引。
`

const SystemSkillGoRavenGuideEn = `---
name: goraven-guide
description: GoRaven platform master index. When users ask GoRaven platform questions, invoke this skill first to match keywords to the right detailed skill.
---

# GoRaven Platform Guide

GoRaven is a self-hostable Agent. Each team member has their own isolated workspace. You run as an AI Agent within GoRaven.

This skill is the master index for platform knowledge. When a user question involves the GoRaven platform itself (not general programming), and you're unsure which detailed skill to consult, match here.

## Skill Index

| Skill | Content |
|------|---------|
| ` + "`goraven-runtime`" + ` | Your runtime environment and constraints. Docker container (ubuntu 22.04) with pre-installed toolchain (Go/Node/Python/git etc.), filesystem and terminal tool inventory, built-in capabilities (web_fetch, visual_understand, etc.), ReAct mode and sub-agent delegation, context compression and tool-result pruning, command safety review (blocked: sudo/global installs/system file writes/credential reads/destructive ops), multi-user isolation |
| ` + "`goraven-features`" + ` | Platform features for users. Free mode and Persona mode, attachment uploads (stored as-is; documents read via the goraven_doc_parse tool), skill system (marketplace/sharing/Always-On), MCP integration (Stdio/SSE/HTTP transports), workspace and .profile environment variable injection, team projects (independently created, settable to all users or members only), session sharing, Markdown rendering and SSE streaming, <goraven-chart> <goraven-file> <goraven-upload> <goraven-ref> tags, system settings (iterations/timeouts/compression/toggles) |
| ` + "`goraven-user-ui`" + ` | Frontend pages for regular users. Sidebar and overall layout, chat page (new: select model/persona/MCP/skills/attachments/project dir/Thinking; session: message stream/token usage/compress/share/stop/background thinking/message actions), multi-session switching, Dashboard (usage stats and trend charts), Workspace (Segmented Tab navigation/upload/preview/compress/decompress/selection-driven action bar/team project creation and browsing), Skills Center (Installed/Marketplace/Team Shares/Always On), Personas (create/edit/templates), Profile (env vars/password/theme/avatar), login, install wizard, and shared conversation view |
| ` + "`goraven-admin-ui`" + ` | Admin backend pages. Dashboard (global stats), Users (CRUD/reset password/enable-disable), Settings (grouped params), Models (provider config/API Key/context length/test connection/permissions & members), MCP (Stdio/SSE/HTTP/templates/Always-On), Skills (global and marketplace), Persona Templates, System Info (DB pool/disk/plugins) |
| ` + "`goraven-install-skill`" + ` | Skill installation flow and directory conventions. **Must invoke this skill first** when user asks to install any skill |
| ` + "`goraven-chart`" + ` | <goraven-chart> tag usage. Use when user asks for data analysis, statistical comparison, or trend visualization |
| ` + "`goraven-automation`" + ` | Automation task (scheduled task) creation guide. **Must invoke** when users request recurring or delayed execution such as "every day at 9am automatically…", "check every N minutes", "remind me next Monday", "run in two hours". Covers execution types and parameter validation, self-contained requirement writing, session context inheritance, scheduling semantics |

## Rules

- One user question may trigger multiple skills. Load as needed.
- If the user question clearly maps to a single sub-skill's domain, load that sub-skill directly — no need to go through this index.
- For pure page navigation questions ("where is X button", "how to get to Y page"), load ` + "`goraven-user-ui`" + ` directly — skip this index.
`
