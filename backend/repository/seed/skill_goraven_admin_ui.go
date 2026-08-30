package seed

const SystemSkillGoRavenAdminUI = `---
name: goraven-admin-ui
description: GoRaven 管理端前端界面指南。详细介绍管理员后台各页面的布局与功能，帮助模型引导管理员完成系统配置与维护操作。
---

# GoRaven 管理端前端界面

## 进入管理模式

管理员通过侧边栏底部用户菜单中的「管理模式」选项切换到后台管理。侧边栏导航变为管理员专用菜单。

## 仪表盘（/admin）

系统全局统计概览，顶部有刷新按钮。页面分五排面板：

**第一排 — 系统脉搏**：活跃用户数（含环比变化箭头）、总会话数、周新增会话、本周/今日 Token 消耗、已启用模型数。下方附带近期 Token 消耗的迷你面积图（Sparkline）。

**第二排 — Token 趋势**：堆叠柱状图（提示消耗 + 补全消耗），可通过分段控件切换 7/30/90 天。图表上方显示日均 Token 量和峰值日期及峰值量。图表中虚线标注日均线。

**第三排 — 模型用量 + 用户排名**：左侧水平条形图展示各模型的 Token 消耗占比；右侧用户 Token 消耗排名，含进度条和百分比。

**第四排 — 活跃趋势**：折线图展示每日活跃用户数，含 7 日移动平均辅助线。也支持 7/30/90 天切换。

**第五排 — 排行**：三列并排展示周技能使用排行、周 MCP 使用排行、周工具使用排行。

## 用户管理（/admin/users）

表格展示用户列表（头像、用户名、昵称、邮箱、角色标签、状态开关、会话数、最后活跃时间、操作按钮）。支持搜索和角色筛选，分页浏览。

**添加用户**：侧边滑出面板。用户名字段实时校验——8 到 16 位、只能包含字母数字和下划线连字符、首尾必须是字母或数字。密码字段校验——8 位以上、必须同时包含字母和数字。可选择角色（普通用户 / 管理员）。表单内对已有用户名做重名提示。

**编辑用户**：侧边滑出面板。用户名只读。可修改昵称、邮箱、角色、启用/禁用状态。状态关闭时有提示文字。

**行内操作**：每行可直接切换启用/禁用开关。操作按钮包括编辑、重置密码、删除。当前登录的管理员无法操作自己（禁用和删除按钮不显示）。

**重置密码**：弹窗输入新密码和确认密码，密码需满足强度要求（8 位以上、含字母和数字）。

**删除用户**：弹窗确认，软删除。

## 系统设置（/admin/settings）

系统级配置，按分组以表单行展示。每行显示设置名称、描述和输入控件，修改过的行有视觉标记。顶部保存按钮显示已修改项数量，点击一次性提交所有变更。

七个分组：Agent 设置、上下文管理、ClawHub 设置、通用设置、分享设置、知识库设置、工具设置。

输入控件根据值类型自动选择：文本输入框、数字输入框（含 min/max 范围限制）、滑块（百分比）、开关按钮、密码输入框、多行文本框、JSON 编辑器（等宽字体）、日期选择器、日期时间选择器、下拉选择框。

## 模型管理（/admin/models）

表格展示已配置模型的列表（提供商名称、显示名、模型名、标签、上下文长度、更新时间、行操作菜单）。支持搜索和分页。

**添加模型**：侧边滑出面板（右侧 400px）。

- 选择提供商后自动填充默认图标和 BaseURL，清空已选模型名
- 提供商名称可自定义（仅用于前端展示区分同厂商多模型）
- 显示名称可选（前端列表展示用，不填则用模型名）
- 图标通过九宫格预设图标库选择
- API Key 字段：显示/隐藏切换按钮、一键复制按钮
- 模型名：输入框带搜索下拉，点击「获取推荐」按钮从提供商 API 拉取可用模型列表，可输入关键词过滤
- 网络代理 URL：可选
- 上下文窗口长度（KB 单位）
- 三个开关：设为默认模型、压缩专用模型、多模态模型
- 压缩模型说明：用于上下文压缩和会话标题生成
- 多模态模型说明：用于 goraven_visual_understand 工具
- openai_compatible 和 claude_compatible 提供商：额外显示 extra_body 文本框（JSON 格式，如 thinking 配置）
- 「保存并测试」按钮：保存后自动测试连接
- 必填：提供商名称、模型名；按提供商要求可能需要 API Key 和 Base URL

**编辑模型**：侧边滑出面板。提供商 ID 只读。API Key、BaseURL、代理 URL 被修改时标记「连接信息已变更」，保存时自动触发连接测试。API Key 和 extra_body 等敏感字段需异步加载详情后显示。

**行操作菜单**：每行右侧三点按钮弹出下拉菜单，含编辑、权限和成员、复制、设为默认、设为压缩、设为多模态、删除。设默认/压缩/多模态仅在该标记未设置时显示对应选项。

**权限和成员**：弹窗设置模型访问权限（全员开放/仅成员可见）并管理可见成员列表（选择器分页搜索用户，可添加/移除）。仅成员可见时，只有指定成员在对话中可选该模型。

**复制模型**：基于现有模型创建，显示名自动追加 "-duplicate" 后缀，API Key 和 extra_body 从详情接口获取。

**删除模型**：弹窗确认。被标记为默认/压缩/多模态的模型不可删除。

## MCP 管理（/admin/mcp）

MCP 端点表格（名称/标识名、传输类型徽标、健康延迟、最后检查时间、启用开关、Always-On 开关、更新时间、编辑/删除按钮）。支持搜索和传输类型筛选，分页浏览。

三种传输类型：**Stdio**（通过 npx/uvx 启动）、**SSE**、**StreamableHTTP**（注意不是普通 HTTP）。

**添加/编辑表单**：宽抽屉（680px），左侧表单 + 右侧实时 JSON 预览。

- 标识名：字母数字和连字符，2-64 位，首字符必须是字母。编辑时只读。
- 显示名和图标（图标选择器）
- 描述文本
- 传输类型下拉：编辑时不可修改
- Stdio 配置：选择运行器（npx/uvx），动态添加/删除启动参数和键值对环境变量
- SSE/StreamableHTTP 配置：服务地址、动态添加/删除请求头（键值对）、代理 URL
- 备注（仅管理员可见）
- 右侧 JSON 预览实时同步表单内容，支持一键复制

**表格中的健康状态**：延迟 <50ms 显示绿色，<200ms 正常色，≥200ms 暗色。延迟为 0 表示离线（灰色圆点）。

**两个独立行内开关**：启用/禁用（控制是否可用）和 Always-On（控制是否对所有用户自动生效）。各自独立切换，有独立的加载状态。

**推荐模板库**：弹窗网格展示内置 MCP 模板，每张卡片含图标、名称、描述、传输类型徽标、安装状态。已安装的显示状态标签。点击安装后刷新列表。

## 技能管理（/admin/skills）

两个标签页：全局技能（goraven-* 系统技能）和市场技能（面向用户的技能商店）。

**全局技能标签页**：
- 表格显示技能描述、goraven-* 名称、启用/禁用开关、更新时间、编辑/删除按钮
- 编辑/新建打开大抽屉（720px）：左侧栏实时解析 YAML frontmatter 中的 name/description 并显示校验结果（名称缺失、描述缺失、goraven- 前缀缺失、名称格式不合法）；右侧为 Markdown 全文编辑器
- 名称必须以 "goraven-" 开头，小写字母数字和连字符组成
- 保存时提示「仅对新会话生效」
- 支持搜索和状态筛选

**市场技能标签页**：
- 表格显示技能图标+名称+描述、来源（ClawHub/自定义上传）、分类、安装次数（点击查看用户列表）、发布状态开关、排序、编辑/查看用户/删除按钮
- 编辑弹窗：可修改图标、分类、排序值、备注；查看 SKILL.md 原文；不可修改技能名称
- 删除时可选级联删除（同时清理已安装此技能的用户记录）
- 支持按来源（ClawHub/自定义上传）和发布状态筛选

**从 ClawHub 导入**：
- 两种模式：探索（按热门/最新/更新/下载/收藏/安装量排序浏览）和搜索（关键词检索，需配置 ClawHub Token）
- 选择技能后显示详情（名称、版本、描述、SKILL.md 内容）
- 导入时需选择目标分类和图标
- ClawHub Token 未配置时显示提示

**发布自定义技能**：上传 zip 文件（使用分块上传），选择目标分类和图标，点击发布。

**分类管理**：弹窗表格显示分类名称、图标、技能数。可增删改分类。名称有宽度限制。

## 系统信息（/admin/systemInfo）

只读运维快照页面。顶部刷新按钮。

**生态概览**（首行全宽）：用户（总数/活跃数/管理员数）、模型（总数/已启用）、MCP（总数/已启用）、技能（全局/市场）、会话和消息数、文件（总数/活跃）、分享（总数/活跃/浏览次数）。

**系统概览**：版本号、系统语言、缓存类型及内存占用量、时区、上传文件总大小、临时文件大小。

**数据库**：类型、版本、库名、数据大小。连接池可视化进度条（使用中 vs 空闲的占比）、开放连接数细分。等待连接数、等待时长、最大空闲关闭数、最大存活关闭数。

**MCP 健康**：分页表格（每页 7 条）。状态指示（绿点=正常延迟<3000ms、黄点=降级延迟≥3000ms、灰点=离线延迟=0）、显示名、延迟毫秒数、最后检查时间。

**磁盘**：设备名、挂载点、文件系统类型、已用/总容量、用量百分比进度条（≥90% 红色、≥80% 黄色、<80% 蓝色）。

**插件列表**：已启用插件的名称和版本号（页脚显示）。
`

const SystemSkillGoRavenAdminUIEn = `---
name: goraven-admin-ui
description: GoRaven admin-side frontend UI guide. Details the layout and functions of each admin backend page, helping the model guide administrators through system configuration and maintenance.
---

# GoRaven Admin-Side Frontend UI

## Entering Admin Mode

Administrators switch to backend management via the "Admin Mode" option in the sidebar user menu at the bottom. The sidebar navigation changes to admin-specific menus.

## Dashboard (/admin)

System-wide statistics overview with a refresh button at the top. Five rows of panels:

**Row 1 — System Pulse**: Active users (with trend arrow showing change), total sessions, new sessions this week, weekly/daily token consumption, enabled model count. Below is a mini sparkline area chart of recent token activity.

**Row 2 — Token Trend**: Stacked bar chart (prompt + completion tokens), toggleable 7/30/90 days via segmented control. Above the chart: daily average tokens and peak date + peak total. A dashed reference line shows the daily average.

**Row 3 — Model Usage + User Ranking**: Left side horizontal bar chart of token consumption share per model; right side user token ranking with progress bars and percentages.

**Row 4 — Active Trend**: Line chart of daily active users with a 7-day moving average auxiliary line. Also switchable 7/30/90 days.

**Row 5 — Rankings**: Three columns side by side: weekly skill usage ranking, weekly MCP usage ranking, weekly tool usage ranking.

## Users (/admin/users)

Table display (avatar, username, nickname, email, role badge, status toggle, session count, last active time, action buttons). Search and role filter supported, with pagination.

**Add user**: Slide-in drawer. Username field with live validation — 8 to 16 characters, only alphanumeric + underscore/hyphen, must start and end with letter or digit. Password field validation — 8+ characters, must contain both letters and numbers. Role selection (regular / admin). Duplicate username detection against existing list.

**Edit user**: Slide-in drawer. Username is read-only. Editable: nickname, email, role, enable/disable status. Hint text shown when status is disabled.

**In-row actions**: Enable/disable toggle directly in the row. Action buttons: edit, reset password, delete. The currently logged-in admin cannot operate on themselves (disable and delete buttons hidden).

**Reset password**: Dialog with new password and confirmation input; must satisfy strength requirements (8+ chars, letters + numbers).

**Delete user**: Confirmation dialog, soft delete.

## Settings (/admin/settings)

System config key-value store displayed as form rows grouped by domain. Modified rows are visually marked. The top save button shows the count of changed settings; click to submit all changes at once.

Seven groups: Agent, Context, ClawHub, General, Sharing, Knowledge, Tools.

Input controls vary by value type: text input, number input (with min/max range), slider (percentage), toggle switch, password input, textarea, JSON editor (monospace font), date picker, datetime picker, dropdown select.

## Models (/admin/models)

Table listing configured models (provider display name, model display name, model name, badges, context length, update time, row action menu). Search and pagination supported.

**Add model**: Slide-in drawer (right side, 400px).

- Selecting a provider auto-fills default icon and BaseURL, clears model name
- Provider display name is customizable (for distinguishing multiple models from the same provider in the UI)
- Model display name is optional (shown in list; defaults to model name if empty)
- Icon selected from a 3x5 grid of preset provider logos
- API Key field: show/hide toggle button, one-click copy button
- Model name: input with searchable dropdown; "Fetch Recommended" button pulls available model list from the provider API; type to filter
- Network proxy URL: optional
- Context window length (in KB)
- Three toggles: set as default model, flash model, multimodal model
- Compression model note: used for context compression and session title generation
- Multimodal model note: used by goraven_visual_understand tool
- For openai_compatible and claude_compatible providers: extra_body textarea appears (JSON format, e.g. thinking configuration)
- "Save & Test" button: saves then auto-tests connectivity
- Required fields: provider display name, model name; API Key and Base URL may be required depending on provider

**Edit model**: Slide-in drawer. Provider ID is read-only. When API Key, BaseURL, or proxy URL are modified, "connection info changed" state is set, triggering an automatic connectivity test on save. Sensitive fields (apiKey, extraFields) are loaded asynchronously from the detail API.

**Row action menu**: Three-dot button per row opens a dropdown with: edit, permissions & members, duplicate, set as default, set as compression, set as multimodal, delete. Set default/compression/multimodal options only appear when that badge is not currently active.

**Permissions & Members**: Dialog to set the model's access scope (All users / Members only) and manage the visible member list (paged user selector with search; add/remove members). When set to "Members only", only the listed members can select the model in chat.

**Duplicate model**: Creates a copy from existing model — display name gets "-duplicate" suffix, API Key and extraFields fetched from detail API.

**Delete model**: Confirmation dialog. Models marked as default/compression/multimodal cannot be deleted.

## MCP (/admin/mcp)

MCP endpoint table (name/identifier, transport type badge, health latency, last check time, enable toggle, Always-On toggle, update time, edit/delete buttons). Search and transport type filter supported, with pagination.

Three transport types: **Stdio** (launches via npx/uvx), **SSE**, and **StreamableHTTP** (note: not plain HTTP).

**Add/Edit form**: Wide drawer (680px), left form + right real-time JSON preview.

- Identifier name: alphanumeric + hyphen, 2-64 chars, must start with a letter. Read-only when editing.
- Display name and icon (icon picker)
- Description textarea
- Transport type dropdown: locked when editing
- Stdio config: select runner (npx/uvx), dynamically add/remove startup arguments and key-value environment variables
- SSE/StreamableHTTP config: service URL, dynamically add/remove request headers (key-value pairs), proxy URL
- Remark (admin-only note)
- Right-side JSON preview syncs live with the form; one-click copy button

**Health status in table**: latency <50ms shown in green, <200ms normal color, >=200ms muted. Latency of 0 means offline (grey dot).

**Two independent in-row toggles**: Enable/Disable (controls availability) and Always-On (auto-applies to all users). Each has its own loading state.

**Recommended template library**: Dialog with a grid of built-in MCP templates. Each card shows icon, name, description, transport type badge, install status. Already installed items show a status label. Click install to add and refresh the list.

## Skills (/admin/skills)

Two tabs: Global skills (goraven-* system skills) and Market skills (user-facing skill store).

**Global skills tab**:
- Table: skill description, goraven-* name, enable/disable toggle, update time, edit/delete buttons
- Create/Edit opens a wide drawer (720px): left sidebar live-parses the YAML frontmatter (name, description) and shows validation results (missing name, missing description, missing "goraven-" prefix, invalid name format); right side is a full Markdown editor
- Name must start with "goraven-" prefix; lowercase alphanumeric + hyphens only
- Save triggers a "takes effect in new sessions" reminder
- Search and status filtering supported

**Marketplace skills tab**:
- Table: skill icon + name + description, source (ClawHub/custom upload), category, install count (clickable to view user list), publish status toggle, sort order, edit/view users/delete buttons
- Edit drawer: modify icon, category, sort order, remark; view SKILL.md content; name is not editable
- Delete with optional cascade (also remove from users' installed records)
- Filter by source (ClawHub/custom upload) and publish status

**ClawHub import**:
- Two modes: Explore (browse sorted by trending/newest/updated/downloads/stars/installs) and Search (keyword query, requires ClawHub token)
- Select a skill to view details (name, version, description, SKILL.md content)
- Import requires selecting a target category and icon
- ClawHub token not configured: shows a prompt message

**Publish custom skill**: Upload a zip file (chunked upload), select target category and icon, then publish.

**Category management**: Dialog table showing category name, icon, skill count. Add/edit/delete categories. Name has a display-width limit.

## System Info (/admin/systemInfo)

Read-only ops snapshot page. Refresh button at the top.

**Ecosystem overview** (full-width top row): Users (total/active/admin), Models (total/enabled), MCPs (total/enabled), Skills (system/market), Sessions & Messages, Files (total/active), Shares (total/active/views).

**System overview**: Version, language, cache type and memory usage, timezone, uploaded files total size, temp files size.

**Database**: Type, version, database name, data size. Connection pool visual bar (in-use vs idle proportion), open connection breakdown. Wait count, wait duration, max idle closed count, max lifetime closed count.

**MCP Health**: Paginated table (7 per page). Status indicator (green dot = normal latency <3000ms, amber dot = degraded latency >=3000ms, grey dot = offline latency = 0), display name, latency in ms, last check time.

**Disks**: Device, mount point, filesystem type, used/total capacity, usage percentage progress bar (>=90% red, >=80% amber, <80% blue).

**Plugin list**: Enabled plugins with name and version (footer display).
`
