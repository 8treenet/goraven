package seed

const SystemSkillGoRavenFeatures = `---
name: goraven-features
description: GoRaven 平台功能概述。涵盖对话模式、附件处理、技能系统、MCP 集成、文件管理、会话分享、前端渲染与 goraven 标签、系统设置。当需了解业务功能如何工作或用户询问具体功能时调用。
---

# 平台功能

## 用户与权限

系统有三种角色：

| 角色 | 能力 |
|------|------|
| 超级管理员 | 安装时创建，唯一。管理所有用户、模型、MCP、系统设置 |
| 管理员 | 由超级管理员创建。管理用户、模型、系统设置 |
| 普通用户 | 使用对话、文件、技能、角色等功能 |

### 管理员用户管理
- 创建/编辑/删除用户（软删除，清理其工作空间）
- 重置用户密码（使其所有登录 token 失效）
- 禁用用户（使其所有 token 失效，阻止登录）
- 用户列表支持搜索和角色筛选

### 登录安全
- 密码认证。连续失败达到阈值时触发算术验证码（图形展示两个数字，需回答和）
- 修改密码时验证当前密码，新密码不可与旧密码相同

## 对话模式

用户以**会话**为单位发起对话，每条会话绑定模型和工具配置。有两种使用方式：

### 自由模式
用户每次手动选择模型、MCP、技能。适合临时任务、尝试不同模型或工具组合。

对话前可在输入区通过配置按钮（+ 扳手图标）选择本次使用的 MCP 和技能。

### 角色模式（Persona）
用户预先创建角色，为角色设定：
- 系统提示词（最大 500 字符）
- 固定模型
- MCP 工具组合
- 技能组合

选择角色即可快速开始同类任务，无需每次重新配置。角色在 /personas 页面管理，对话时通过顶部下拉框选择。选择角色后配置按钮禁用（角色已预配置工具和技能）。

## 对话中的附件

用户在聊天中上传的图片、文档等附件，系统在对话开始前自动处理并注入到消息中：

- **图片**（jpg/png/gif/webp 等）：原样存入 <根路径>/temp/ 目录，你可用文件读取工具查看。
- **文档**（PDF、Word、PPT、Excel、HTML 等）：系统自动将其内容解析为 Markdown 文本，生成 .md 文件存入 <根路径>/temp/ 目录，你可用文件读取工具获取文档的文字内容。
- **纯文本文件**（txt、log、json、xml、yaml 等）：原样复制到 <根路径>/temp/ 目录。

附件文件的路径通过 <goraven-upload> 标签出现在消息内容中（该标签由系统生成，你不需要输出）。消息中的附件是用户任务的重要上下文，理解附件内容后再执行任务。

### 大文件上传

前端支持分块上传协议：大文件自动拆分为分片上传，每片通过 MD5 校验完整性，全部上传后合并。上传完成后文件移入用户空间。

### 停止对话

运行中的对话可通过前端停止按钮终止。后端向 Agent 发送取消信号，当前轮次的 shell 进程及其子进程全部终止。

### 手动压缩

用户可在对话中手动触发上下文压缩（前端操作），系统异步执行后将早期轮次归纳为摘要。前端可轮询压缩进度。

## 技能系统

技能是用户安装的专项能力扩展，遵循 SKILL.md 规范。

- **用户技能**：用户从技能市场安装，或自行创建自定义技能。可按用户分享给团队内其他成员。
- **全局技能**：由管理员管理（如 goraven-guide、goraven-runtime、goraven-chart 等），对所有用户可见。管理员可在 /admin/skills 页面启用/禁用。
- **技能市场**：用户可在 /skills 页面浏览和安装市场中的技能。点击安装仅复制文件并注册——依赖安装（如 pip/npm install）需在对话中完成。
- **Always-On 开关**：用户可将常用技能设为 Always-On，每次对话自动加载。

**重要**：当用户要求安装任何技能时，必须先调用 ` + "`goraven-install-skill`" + ` 技能获取目录规范和配置流程。

### 技能共享
用户可将自己创建的自定义技能共享给团队：

- 共享时文件从用户沙盒 skills/ 复制到团队共享目录
- 团队其他成员可在技能页面看到共享列表（标注所有者姓名）
- 成员点击安装后，文件复制到自己的沙盒 skills/ 目录
- 共享者可以最新文件覆盖更新共享
- 共享者或管理员可删除共享

### 技能自动同步
每次对话完成后，系统自动扫描用户沙盒 skills/ 目录，与新目录名同步数据库记录（自动注册未录入的自定义技能、清理已删除目录的记录）。

## MCP 集成

MCP（Model Context Protocol）允许连接外部服务扩展你的工具能力。

- **管理员配置**：管理员在 /admin/mcp 页面添加 MCP 端点，支持三种传输类型：
  - Stdio：通过 npx/uvx 启动本地程序，可配置参数和环境变量
  - SSE/HTTP：连接远程服务 URL，可配置请求头
- **自动注册**：启用的 MCP 端点自动注册为你的工具，在对话中可直接调用。
- **Always-On**：管理员可将某些 MCP 设为「始终启用」，对所有用户生效，无需用户在对话中手动选择。
- 有推荐模板库可快速安装常用 MCP。

### MCP 健康检查
系统每 3 小时自动检测所有已启用 MCP 端点的连通性：调用工具列表接口，验证失败则自动禁用该端点并清除角色中的工具关联。管理员也可手动触发全量健康检查。

## 仪表盘

系统提供管理员仪表盘（全局 Token 趋势、模型分布、用户排名、工具使用排行，当日异常消耗自动告警）和用户个人仪表盘（个人用量趋势、模型分布、文件空间占用）。仪表盘数据 10 分钟缓存。界面操作见 ` + "`goraven-admin-ui`" + ` 和 ` + "`goraven-user-ui`" + `。

## AI 模型管理

管理员可配置多个 LLM 模型，每个模型指定提供商、API 密钥、BaseURL、模型名、上下文长度。模型分为三种标记：

- **默认模型** —— 用户对话的默认选择
- **压缩模型** —— 用于上下文压缩和会话标题生成
- **多模态模型** —— 用于 goraven_visual_understand 工具

这三种模型不可删除，防止关键功能静默丢失。添加/编辑模型时支持「保存并测试」验证连通性。界面操作见 ` + "`goraven-admin-ui`" + `。

## 角色（Persona）详情

角色是用户预设的对话配置模板，绑定系统提示词（最大 500 字符）、固定模型、MCP 工具组合和技能组合。

创建或更新角色时系统自动检测 MCP 工具名称冲突和技能名称冲突，冲突存在时阻止保存。角色可基于管理员预设模板创建，也可完全自定义。查看角色详情时自动清理已过期的 MCP/技能关联。

界面操作见 ` + "`goraven-user-ui`" + `，模板管理见 ` + "`goraven-admin-ui`" + `。

## 文件管理

用户拥有独立的文件空间，固定目录结构：

| 目录 | 用途 |
|------|------|
| documents/ | 文档存储 |
| projects/ | 项目代码，可设为对话工作空间 |
| images/ | 图片存储 |
| videos/ | 视频存储 |
| downloads/ | 下载文件 |
| temp/ | 临时文件（附件自动存放于此） |
| skills/ | 已安装技能 |

支持上传、新建文件夹、删除、压缩/解压、重命名、预览（文本/图片/视频/音频/PDF）。界面操作见 ` + "`goraven-user-ui`" + `。

### 团队项目共享

用户可将 projects/ 下的子目录共享给团队成员，实现跨用户协作。

- **共享**：在文件管理页右键 projects/ 下的子目录，选择「共享到团队」，填写简介后共享。已共享的目录显示 Users 图标徽章。
- **共享限制**：已共享的项目子目录不可删除和重命名，需先取消共享。
- **团队项目列表**：在文件管理页右上角点击「团队项目」按钮，查看所有团队成员共享的项目卡片（显示项目名、所有者、简介、更新时间）。单击卡片进入项目文件浏览。
- **项目文件浏览**：进入共享项目后，可像普通文件管理一样上传、新建、删除、重命名、压缩、解压、预览和下载文件。顶部导航与文件管理一致（返回上级 + 当前目录名）。
- **管理权限**：仅项目所有者可编辑简介和取消共享（卡片右上角菜单）。
- **取消共享**：取消共享后团队成员将无法访问该项目。
- **并发控制**：开始对话时系统对共享项目加锁（Redis SetNX，30 分钟 TTL），同一时刻只有一个 Agent 可操作该项目。对话完成后自动释放锁。若项目正被其他会话占用则返回错误提示。该锁对两种访问方式均生效：通过团队项目列表访问（会话记录中保存 shared_project_id），以及项目所有者在对话中直接选择自己的已共享项目。

### 环境变量（.profile）
工作空间根目录的 .profile 文件管理用户环境变量，**用户和 AI 均可读写**。

- 格式：每行 KEY=VALUE，# 注释，支持 export KEY=VALUE
- 用户可在 /files 页面通过「环境变量」按钮（{} 图标）编辑
- 你执行 shell 命令时，系统**自动将 .profile 中所有变量注入命令环境**，无需 source/export，直接用 $VAR_NAME 引用
- .profile 变量后注入，同名覆盖系统环境变量
- 遇鉴权失败或缺少环境变量时，先读 .profile 排查，向用户确认后再写入，**禁止猜测填充**

### 文件外链
用户可为工作空间内的文件生成外部访问链接（可设置过期时间，由系统设置控制），支持公开访问。适用于向团队外部分享文件。

## 系统信息

管理员可在 /admin/systemInfo 查看运维快照（版本、数据库状态、磁盘使用、MCP 健康、生态计数等）。界面详见 ` + "`goraven-admin-ui`" + `。

## 会话分享

用户可将会话生成分享链接：

- 在对话页顶部工具栏点击分享按钮
- 可设置有效期和分享类型（公开/内部）
- 其他人通过链接查看对话内容（只读）
- 技能也可分享给团队内多用户

## 前端渲染与标签

### Markdown 渲染
你输出的 Markdown 被前端渲染为富文本：标题、列表、表格、代码高亮（围栏代码块）、LaTeX 数学公式。

### SSE 流式推送
回复通过 SSE（Server-Sent Events）流式推送到前端，用户实时看到输出。事件类型：

| 事件 | 含义 |
|------|------|
| reasoning | 思考过程（显示在可折叠区域） |
| content | 正式回复内容（逐字流式显示） |
| tool | 工具调用过程（以卡片形式呈现） |
| retry | 模型请求失败，系统按退避策略自动重试 |
| end | 本轮回复结束 |
| context | 携带当前 token 用量与模型上下文长度，前端据此渲染用量条 |
| heartbeat | 连接保活 |

### GoRaven 标签
以下标签有特殊渲染效果，规则详见系统提示词：

- <goraven-chart>：嵌入数据可视化图表（柱状图、折线图、饼图、面积图、散点图）。用法详见 ` + "`goraven-chart`" + ` 技能。
- <goraven-file>：回复文件引用标签（由你输出），前端渲染为可点击的链接或预览卡片。
- <goraven-upload>：附件标签，由系统生成，你不需要输出。标记用户上传的附件路径。
- <goraven-ref>：用户文件引用标签，由系统生成，你不需要输出。标记用户在输入框中 @ 选择的文件或目录的路径。type="file" 表示文件，type="dir" 表示目录。

## 系统设置要点

管理员可在 /admin/settings 调节以下关键配置，这些会影响你的运行行为：

- **迭代与超时**：最大迭代步数、单次查询超时。
- **上下文管理**：压缩触发阈值与保留轮数、裁剪 token 阈值与工具结果截断长度。
- **请求策略**：LLM 请求间隔、失败重试次数、限流等待与退避基数。
- **能力开关**：Web 抓取、视觉理解、OCR 是否启用；Shell 单命令超时；文件分享链接有效期。

某项能力不可用（如 web_fetch 被禁用）通常是管理员配置所致，可提示用户联系管理员。界面操作见 ` + "`goraven-admin-ui`" + `。

## 前端界面

用户通过左侧边栏操作平台。普通用户可访问仪表盘、文件管理、技能中心、角色管理及历史会话；管理员额外拥有用户管理、系统设置、模型配置等页面。详细界面与操作说明见 ` + "`goraven-user-ui`" + `（用户侧）和 ` + "`goraven-admin-ui`" + `（管理端）技能。
`

const SystemSkillGoRavenFeaturesEn = `---
name: goraven-features
description: GoRaven platform features overview. Covers chat modes, attachments, skill system, MCP integration, file management, session sharing, frontend rendering and goraven tags, and system settings. Invoke when you need to understand how business features work or the user asks about specific functionality.
---

# Platform Features

## Users & Permissions

Three role tiers:

| Role | Capabilities |
|------|-------------|
| Super Admin | Created at install, unique. Manages all users, models, MCPs, system settings |
| Admin | Created by super admin. Manages users, models, system settings |
| User | Uses chat, files, skills, personas |

### Admin User Management
- Create/edit/delete users (soft delete, cleans up their workspace)
- Reset user passwords (invalidates all their login tokens)
- Disable users (invalidates all tokens, blocks login)
- User list supports search and role filtering

### Login Security
- Password authentication. Consecutive failures beyond threshold trigger an arithmetic captcha (two numbers displayed as an image; user must answer the sum)
- Changing password verifies current password; new password must differ from old

## Chat Modes

Users converse in **sessions**, each bound to a model and tool configuration. Two modes are available:

### Free Mode
Users manually select the model, MCP tools, and skills for each session. Best for ad-hoc tasks or trying different model/tool combinations.

Before starting a chat, users can select MCP tools and skills via the config button (+ wrench icon) in the input area.

### Persona Mode
Users pre-create personas with:
- A system prompt (max 500 characters)
- A fixed model
- An MCP tool set
- A skill set

Selecting a persona instantly sets up the session for recurring task types — no reconfiguration needed. Personas are managed at /personas and selected from the top dropdown in chat. When a persona is selected, the config button is disabled (the persona has pre-configured tools and skills).

## Attachments in Chat

When users upload images, documents, or other files in a conversation, the system processes them before the agent runs and injects them into the message:

- **Images** (jpg/png/gif/webp, etc.): stored as-is under <root>/temp/. Read with file tools.
- **Documents** (PDF, Word, PPT, Excel, HTML, etc.): the system automatically extracts their text content into a Markdown .md file under <root>/temp/. Read with file tools to get the document's text content.
- **Plain text files** (txt, log, json, xml, yaml, etc.): copied as-is to <root>/temp/.

Attachment paths appear in the message via <goraven-upload> tags (system-generated; you do not output these). Attachments provide important context — understand them before executing the task.

### Large File Upload
The frontend supports chunked upload: large files are split into chunks, each verified by MD5 checksum, then merged. Completed uploads are moved into the user space.

### Stop Chat
A running chat can be stopped via the frontend stop button. The backend sends a cancel signal to the agent; the current round's shell process and all child processes are terminated.

### Manual Compression
Users can manually trigger context compression during a chat (frontend action). The system compresses earlier rounds into a summary asynchronously; the frontend can poll for progress.

## Skill System

Skills are user-installed capability extensions following SKILL.md conventions.

- **User skills**: Users install skills from the marketplace or create their own custom skills. Skills can be shared with team members.
- **System (global) skills**: Managed by admins (e.g., goraven-guide, goraven-runtime, goraven-chart), visible to all users. Admins enable/disable them at /admin/skills.
- **Skill marketplace**: Users browse and install marketplace skills at /skills. Clicking Install only copies files and registers the skill — dependency installation (pip/npm install, etc.) must be completed in a conversation.
- **Always-On toggle**: Users can set frequently-used skills as Always-On so they load automatically in every conversation.

**Important**: When a user asks to install any skill, invoke ` + "`goraven-install-skill`" + ` first to get the directory conventions and configuration flow.

### Skill Sharing
Users can share their custom skills with the team:

- Sharing copies files from the user's sandbox skills/ to the team shared directory
- Team members see the shared list on the skills page (with owner name)
- Members install with one click — files copy to their own sandbox skills/ directory
- The sharer can update the shared skill with the latest files
- The sharer or admin can delete the share

### Skill Auto-sync
After each chat completes, the system scans the user's sandbox skills/ directory and syncs database records (auto-registering unrecorded custom skills, removing records for deleted directories).

## MCP Integration

MCP (Model Context Protocol) allows connecting external services to extend your tool capabilities.

- **Admin configuration**: Admins add MCP endpoints at /admin/mcp. Three transport types are supported:
  - Stdio: launch a local program via npx/uvx, with configurable arguments and environment variables
  - SSE/HTTP: connect to a remote service URL, with configurable request headers
- **Auto-registration**: Enabled MCP endpoints auto-register as your tools and can be called directly in conversations.
- **Always-On**: Admins can set certain MCPs as "always on" for all users — no manual selection needed per session.
- A recommended template library is available for quick MCP installation.

### MCP Health Check
Every 3 hours, the system auto-checks connectivity of all enabled MCP endpoints by calling their tool list. On failure, the endpoint is auto-disabled and its tool associations are cleared from personas. Admins can also manually trigger a full health check.

## Dashboard

The system provides an admin dashboard (global token trends, model distribution, user rankings, tool usage rankings; auto-alerts on abnormal daily consumption) and a user personal dashboard (personal usage trends, model distribution, file space usage). Dashboard data is cached for 10 minutes. UI details in ` + "`goraven-admin-ui`" + ` and ` + "`goraven-user-ui`" + `.

## AI Model Management

Admins can configure multiple LLM models, each specifying a provider, API key, BaseURL, model name, and context length. Models have three special markers:

- **Default model** — the default choice for user chats
- **Compress model** — used for context compression and session title generation
- **Visual model** — used by the goraven_visual_understand tool

These three models cannot be deleted to prevent silent loss of critical functionality. A "Save & Test" button verifies connectivity when adding/editing. UI details in ` + "`goraven-admin-ui`" + `.

## Persona Details

A persona is a user's preset chat configuration binding a system prompt (max 500 chars), a fixed model, an MCP tool set, and a skill set.

When creating or updating a persona, the system auto-detects MCP tool name conflicts and skill name conflicts, blocking save if found. Personas can be based on admin-preset templates or fully custom. Expired MCP/skill associations are auto-cleaned when viewing persona details.

UI details in ` + "`goraven-user-ui`" + `, template management in ` + "`goraven-admin-ui`" + `.

## File Management

Each user has an independent file space with a fixed directory structure:

| Directory | Purpose |
|-----------|---------|
| documents/ | Document storage |
| projects/ | Project code; can be set as the chat workspace |
| images/ | Image storage |
| videos/ | Video storage |
| downloads/ | Downloaded files |
| temp/ | Temporary files (attachments auto-placed here) |
| skills/ | Installed skills |

Supports upload, new folder, delete, compress/decompress, rename, and preview (text/images/video/audio/PDF). UI details in ` + "`goraven-user-ui`" + `.

### Team Project Sharing

Users can share subdirectories under projects/ with team members for cross-user collaboration.

- **Sharing**: In the File Manager, right-click a subdirectory under projects/ and select "Share to Team". Enter a description and share. Shared directories show a Users icon badge.
- **Sharing restrictions**: Shared project subdirectories cannot be deleted or renamed — unshare first.
- **Team project list**: Click the "Team Projects" button in the top-right of the File Manager to see all shared project cards (project name, owner, description, update time). Click a card to enter the project file browser.
- **Project file browser**: Inside a shared project, you can upload, create, delete, rename, compress, decompress, preview, and download files just like normal file management. The top navigation is consistent with the File Manager (back button + current directory name).
- **Management permissions**: Only the project owner can edit the description and unshare (card top-right menu).
- **Unsharing**: Unsharing removes team members' access to the project.
- **Concurrency control**: When starting a chat, the system acquires a lock on the shared project (Redis SetNX, 30min TTL), ensuring only one Agent operates on it at a time. The lock is automatically released when the chat completes. If the project is already in use by another session, an error is returned. This lock applies both when accessing a shared project via the team project list (the session stores a shared_project_id) and when the project owner directly selects their own shared project in a chat.

### Environment Variables (.profile)
The .profile file at the workspace root manages user environment variables. **Both user and AI can read and write it.**

- Format: one KEY=VALUE per line, # for comments. Also supports export KEY=VALUE.
- Users can edit via the "Environment Variables" button ({} icon) on the /files page
- When you execute shell commands, the system **auto-injects all .profile variables into the command environment** — no source/export needed, reference with $VAR_NAME directly
- .profile vars are injected last, overriding same-name system environment variables
- On auth failure or missing env var, inspect .profile first, confirm with user before writing — **never guess values**

### File External Links
Users can generate external access URLs for files in their workspace (configurable expiry, controlled by system settings). Supports public access — useful for sharing files outside the team.

## System Info

Admins can view an ops snapshot at /admin/systemInfo (version, database status, disk usage, MCP health, ecosystem counts, etc.). UI details in ` + "`goraven-admin-ui`" + `.

## Session Sharing

Users can generate share links for sessions:

- Click the share button in the chat page top toolbar
- Configurable expiry and share type (public/internal)
- Others view the shared conversation via the link (read-only)
- Skills can also be shared with team members

## Frontend Rendering & Tags

### Markdown Rendering
Your Markdown output is rendered as rich content: headings, lists, tables, code highlighting (fenced code blocks), and LaTeX math.

### SSE Streaming
Replies are streamed to the frontend via SSE (Server-Sent Events) in real time. Event types:

| Event | Meaning |
|-------|---------|
| reasoning | Thinking process (shown in a collapsible area) |
| content | Actual reply content (streamed character by character) |
| tool | Tool invocation process (shown as cards) |
| retry | Model request failed; system auto-retries with backoff |
| end | Current round finished |
| context | Carries current token usage and model context length; frontend renders a usage bar |
| heartbeat | Keep-alive signal |

### GoRaven Tags
The following tags have special rendering behavior. Detailed rules are in the system prompt:

- <goraven-chart>: Embed data visualization charts (bar, line, pie, area, scatter). Usage details in the ` + "`goraven-chart`" + ` skill.
- <goraven-file>: Reply file reference tag (output by you); frontend renders as a clickable link or preview card.
- <goraven-upload>: Attachment tag, system-generated. You do not output these. Marks paths of user-uploaded attachments.
- <goraven-ref>: User file reference tag, system-generated. You do not output these. Marks paths of files or directories the user @-selected in the input. type="file" for files, type="dir" for directories.

## System Settings

Admins can tune key settings at /admin/settings that affect your runtime behavior: max iteration steps, per-query timeout, compression/pruning thresholds, LLM request retry policy, and capability toggles (web fetch, visual understanding, OCR, shell timeout, file share-link expiry). If a capability is unavailable, it is usually an admin setting — suggest the user contact their admin. UI details in ` + "`goraven-admin-ui`" + `.

## Frontend UI

Users navigate through the left sidebar. Regular users access Dashboard, Files, Skills, Personas, and session history; admins additionally have User Management, System Settings, Model Configuration, and more. For detailed UI and operation instructions, see the ` + "`goraven-user-ui`" + ` (user-side) and ` + "`goraven-admin-ui`" + ` (admin-side) skills.
`
