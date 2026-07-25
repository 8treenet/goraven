package seed

const SystemSkillGoRavenUserUI = `---
name: goraven-user-ui
description: GoRaven 用户侧前端界面与操作指南。详细介绍普通用户可访问的各页面布局、功能与使用方式，帮助模型引导用户完成日常操作。
---

# GoRaven 用户侧前端界面

## 整体布局

GoRaven 采用左侧边栏 + 右侧主内容区布局。侧边栏可折叠为仅图标模式。

### 侧边栏（桌面端）

- **顶部**：GoRaven logo + 折叠/展开按钮
- **新建对话按钮**：点击进入空白对话页，开始新的 AI 交互
- **工作区导航**（折叠分组）：
  - 仪表盘：查看个人用量统计
  - 文件管理：管理个人工作空间文件
  - 技能中心：管理已安装的技能，浏览技能市场
  - 角色管理：创建和管理 AI 角色（预配置模型、MCP、技能的系统提示）
- **会话历史**：按时间分组（今天 / 近7天 / 近30天），支持无限滚动加载，每条会话可重命名或归档删除
- **底部用户区**：显示头像和用户名，点击弹出菜单：
  - 个人设置
  - 管理模式（仅管理员可见，切换到后台管理）
  - 退出登录

### 移动端

侧边栏默认隐藏，通过顶部工具栏的菜单按钮打开，以遮罩形式出现。

## 对话页（核心交互界面）

对话页是用户与 AI 交互的主要界面，分两种状态。

用户可同时进行多个对话。侧边栏会话历史中点击即可切换会话，每个会话独立保持各自的消息历史、模型配置和生成状态。在 A 会话生成过程中切换到 B 会话继续对话时，A 会话会在后台继续完成，不会丢失进度。

### 新建对话（/chat，无会话）

**顶部工具栏**：
- 模型/角色选择下拉框，分为「模型」和「角色」两个区域。选择一个角色会自动配置模型、MCP 工具和技能；选择一个模型会清除角色选择。
- 如果系统尚未配置任何模型，显示「无模型」警告按钮，点击后弹窗说明配置步骤。
- 移动端显示菜单按钮。

**初始状态**：中心显示 GoRaven logo 和「What can I help you build?」问候语。

**输入区域**（完整的 NewChatInput）为一个圆角容器，包含：
- **文本输入框**：支持多行，自动增高（上限约4行），Enter 发送（Shift+Enter 换行）。中文输入法下 Enter 不会误发送。
- **配置按钮（+ 扳手图标）**：打开浮动面板，可选择本次对话使用的 MCP 工具和技能。选中的工具/技能会显示高亮。当选择了角色时，此按钮禁用（角色已预配置工具和技能）。
- **文件上传按钮（回形针图标）**：点击选择本地文件，最多 10 个。上传使用分块方式支持大文件。图片文件会显示缩略图预览。上传完毕或失败有状态提示。可点击移除已上传文件。
- **项目目录按钮（文件夹图标）**：选择 projects/ 下的子目录作为本次对话的工作空间。选择后高亮显示。有「清空选择」选项。若选择的项目为团队共享项目（直接选择自己的已共享项目或通过团队项目列表进入），系统会加锁保护，若被其他会话占用则提示「该团队共享项目正在被其他会话使用」。
- **Thinking 开关**：切换模型的深度思考模式。发送前可随意切换，发送后不可改。
- **发送/停止按钮**：向上箭头样式，高亮色。输入为空时灰色不可用。发送后变为停止按钮（圆形停止图标）。发送后有约 3 秒冷却期，期间停止按钮暂时不可用（防止误触），冷却结束后可随时停止生成。停止后已生成的部分内容会保存为一条截断消息。

**停止与后台运行**：
- 用户可随时点击停止按钮中断生成。停止后，已输出的部分内容保留在对话中。
- 若生成超过 3 分钟仍未完成，前端自动断开 SSE 流，页面切换为「后台思考中」状态并开始轮询。用户可切换到其他会话继续工作，当前会话在后台持续运行。
- 后台运行完成后，侧边栏会话状态恢复正常，消息列表自动刷新显示完整结果。
- 后台运行期间，该会话的输入框禁用，顶部显示运行状态指示。

**发送后**：页面自动跳转到 /chat/:newSessionId，进入会话对话模式。顶部工具栏切换为会话模式。

### 会话对话（/chat/:sessionId，已有会话）

**加载**：进入时显示加载动画，加载完成后显示消息列表。

**顶部工具栏**：
- 移动端菜单按钮
- 点击角色名/模型名可打开详情弹窗，查看当前角色或模型的完整配置（名称、图标、描述、模型、MCP 工具列表、技能列表、Token 用量明细）
- 项目路径标识（如果设置了）
- Token 用量显示（如「12K / 128K」），实时更新
- 上下文压缩按钮：手动触发上下文压缩以释放 token 空间
- 分享按钮：生成会话分享链接（可设置有效期和分享类型）

**消息区域**：按时间顺序展示用户/助手/工具/摘要消息。助手消息实时流式显示（SSE 推送），包括思考内容（reasoning）和正式回复（content）。工具调用过程也会以卡片形式呈现。每条助手消息底部有操作按钮：复制（一键复制回复内容）、点赞、点踩，用于反馈回复质量。消息列表自动滚动到最新内容，用户滚动到上方查看历史时不会强制拉回底部。

**底部输入区**（简化的 ChatInput，无 + 配置按钮）：
- 文件上传和 Thinking 开关同上，但配置已在新建时锁定。
- 思考开关状态指示禁用时不可修改
- 当会话在后台运行时（状态为"后台思考中"），输入框显示提示文字且禁用。
- 发送按钮在压缩过程中样式变化。
- 底部有 AI 免责声明文字。

**模型故障重试**：模型调用失败时系统自动重试（带退避策略），前端会收到 retry 事件。

## 仪表盘页（/dashboard）

展示当前用户的个人使用统计。

- **Token 概览**：以卡片形式展示累计 Token、本周 Token、今日 Token、会话数、新增会话数。
- **存储空间**：展示已用/剩余/总空间及进度条，按目录分类（documents、projects 等）。
- **Token 趋势图**：堆叠柱状图（提示消耗 + 补全消耗），可按 7/30/90 天切换。附带日均值和峰值日。
- **模型用量**：横条图展示各模型的 Token 占比。
- **排行榜**（三列并排）：周工具使用排行、周 MCP 使用排行、周技能使用排行。

顶部有刷新按钮。

## 文件管理页（/files）

管理用户的个人工作空间文件。

**工具栏**：
- 返回上级目录
- 当前目录名
- 上传文件
- 新建文件夹
- 删除选中项（default 目录和只读文件不可删）
- 压缩选中文件（创建 zip 包）
- 解压（仅选中单个 .zip 文件时可点击，可选是否解压到子目录）
- 刷新

**文件列表**：表格形式，列包括复选框（全选/单选，支持 Shift 连选）、名称（双击进入目录）、大小、修改时间。

**键盘快捷键**：F2 重命名选中文件、Delete 或 Backspace 删除选中项、Shift+点击连续多选。

**右键菜单**：预览、下载、重命名、删除、共享到团队（仅 projects/ 下的目录）、取消共享（仅已共享的目录）。

**共享限制**：projects/ 下已共享的子目录不可删除和重命名，需先取消共享。

**预览对话框**：支持以下文件类型的在线预览：

| 类型 | 说明 |
|------|------|
| 图片（png/jpg/gif/svg/webp 等） | 直接嵌入展示，支持放大全屏 |
| 视频（mp4/webm/mov 等） | 内嵌播放器，支持控制条 |
| 音频（mp3/wav/ogg 等） | 内嵌播放器 |
| PDF | iframe 内嵌展示 |
| 文本（txt/json/yaml/code 等） | 纯文本滚动查看，大文件自动截断 |
| Markdown（md/markdown） | 渲染为 HTML，支持标题、表格、代码高亮、图表 |
| Excel（xlsx） | 渲染为 HTML 表格，多 Sheet 标签切换，表头高亮 |
| HTML（html/htm） | iframe 沙箱内渲染，JS 可执行但隔离父页面 |

**HTML 预览注意事项**：
- HTML 中引用的 CSS、JS、图片需使用**相对路径**（如 ./assets/style.css），不要使用根绝对路径（如 /assets/style.css）。相对路径会自动解析为带临时访问凭证的 URL，确保资源可加载；根绝对路径会丢失凭证导致 404 和 CORS 错误。
- 如果是 Vite 等构建工具产物，构建时需设置 base 为 ./ （或命令行加 --base=./）使产物使用相对路径。
- 依赖 localStorage / sessionStorage 的 SPA 应用无法在沙箱预览中正常运行（浏览器安全限制）。

**拖拽上传**：支持拖拽文件到页面区域上传。

用户工作空间下有固定目录，不可删除：

| 目录 | 用途 |
|------|------|
| documents/ | 文档类文件，AI 可读写 |
| projects/ | 项目代码，对话中可选子目录作为工作空间 |
| images/ | 图片文件存储 |
| videos/ | 视频文件存储 |
| downloads/ | 下载文件默认存放位置 |
| temp/ | 临时文件（附件解析结果、中间产物），AI 可读写 |
| skills/ | 已安装技能的文件存放位置（文件管理器中不可见——技能通过数据库管理和授权） |

> **文件管理器隐藏规则**：根目录下「skills/」目录不展示（技能通过数据库管理，禁止文件操作）；以「.」开头的文件和目录（如 .DS_Store、.profile）也不展示——但 .profile 可通过工具栏「环境变量」按钮编辑。

其中 documents/、projects/、temp/ 是 AI Agent 在执行任务时最常操作的目录。

### 环境变量（.profile）

工作空间根目录的 .profile 文件管理用户环境变量，**用户和 AI 均可读写**。

**文件格式**：每行 KEY=VALUE，# 注释，支持 export KEY=VALUE（自动去除 export）。

**用户操作**（/files 页面）：
- 工具栏「环境变量」按钮（{} 图标）打开编辑对话框，增删改键值对，自动保存到 .profile

**AI 操作**（通过文件工具）：
- 用 read_file / write_file 直接读写工作空间根目录的 .profile
- 执行 shell 命令时，系统**自动将 .profile 中所有变量注入命令环境**，无需 source/export，命令中直接用 $VAR_NAME 引用
- 遇鉴权失败或缺少环境变量时，AI 先读 .profile 排查，向用户确认后再写入，**禁止猜测填充**

**注入规则**：.profile 变量后注入，同名覆盖系统环境变量。

**典型场景**：
- 用户在文件管理页设置 BRAVE_SEARCH_API_KEY=xxx，之后所有对话中 AI 的命令都能直接用 $BRAVE_SEARCH_API_KEY
- AI 发现需要 GITHUB_TOKEN 时提示用户：「需要 GITHUB_TOKEN，请在文件管理页的环境变量中添加，或直接告诉我值我帮你写入 .profile」

### 团队项目

文件管理页右上角「团队项目」按钮可切换到团队项目列表视图，查看所有团队成员共享的项目。

**团队项目列表**：
- 以卡片网格展示所有共享项目，每张卡片显示项目名、所有者（头像+姓名）、简介、更新时间。
- 单击卡片进入项目文件浏览。
- 项目所有者卡片右上角有「更多」菜单（悬停显示），可编辑简介和取消共享。
- 顶部有「我的文件」按钮（右上角）可返回个人文件管理。

**项目文件浏览**：
- 进入共享项目后，界面与文件管理页一致：返回上级按钮、当前目录名、上传、新建文件夹、删除、压缩、解压、刷新。
- 支持完整的文件操作：上传、新建、删除、重命名、压缩、解压、预览、下载。
- 顶部右上角有「我的文件」按钮可直接返回个人文件管理。
- 返回上级按钮在项目根目录时返回团队项目列表。

## 技能管理页（/skills）

管理用户技能，分三个标签页：

| 标签 | 内容 |
|------|------|
| 已安装（有数量） | 用户的技能列表。每项显示图标、名称、描述、Always-On 开关、安装时间。点击行打开详情侧边栏（可编辑图标和分类、查看 SKILL.md 内容、切换 Always-On、分享自定义技能、删除） |
| 技能市场（有数量） | 技能市场列表，每项显示安装按钮和安装次数。已安装的显示「已安装」标记。**注意**：点击安装仅复制文件并注册，依赖安装（如 pip/npm install）需在对话中完成 —— 弹窗会提供提示语，用户复制到新对话发给 AI 完成，此过程较耗时。 |
| 团队共享 | 其他用户分享的技能，可安装或（自己的分享）取消分享 |

顶部有搜索框和分类筛选下拉框。右上角有同步（刷新）按钮。

**Always-On 说明**：开启 Always-On 的技能会在用户**每次新建对话时自动加载**，无需手动在配置面板中勾选。适合频繁使用的技能（如代码格式化、常用工具）。可在已安装列表或详情侧边栏中随时切换。开启过多 Always-On 技能会占用每次对话的上下文空间，建议仅对高频使用的技能开启。

## 角色管理页（/personas）

创建和管理 AI 角色（预配置的系统提示 + 模型 + MCP + 技能组合）。

**角色列表（/personas）**：顶部有「新建角色」按钮。每个角色卡片显示图标、名称、分类、角色提示摘要、MCP/技能标签、模型名。点击进入详情，右侧有编辑/删除菜单。

**角色详情/编辑**：可编辑名称、图标、分类、角色设定（textarea，最长 500 字符）、模型配置、MCP 工具（可搜索复选框列表）、技能（可搜索复选框列表）。新建时可从模板库选择模板（按分类分组，支持搜索）。图标通过图标选择器从预设图标库中选取。

保存/取消按钮固定在顶部。

## 个人资料页（/profile）

个人设置，分为：
1. **个人信息**：头像（可点击上传）、昵称（可编辑）、用户名（只读）、角色标签、邮箱（可编辑）、注册时间。
2. **外观**：主题模式选择（浅色/深色/跟随系统）。
3. **账号安全**：修改密码（需验证当前密码，新密码至少 8 位且包含字母和数字）。
4. **关于**：项目介绍 + 退出登录。

## 其他页面

### 分享查看页（/share/:shareId）

通过会话分享功能生成的公开链接，接收者访问此页面查看共享的对话内容。

- **公开分享**：无需登录即可查看对话消息。页面显示分享标题、分享者名称、分享类型、浏览次数和过期时间。
- **内部分享**：需要登录后才能查看消息内容。
- **消息展示**：以只读形式呈现对话历史，包含思考内容（reasoning）和正式回复。消息底部有复制、点赞、点踩按钮。
- **异常状态**：
  - 分享不存在（notFound）
  - 分享已过期（expired）：显示过期提示
  - 加载失败（loadFailed）：显示重试提示
- 页面底部有 GoRaven 品牌信息和 GitHub 链接。

### 登录页（/login）
用户名、密码输入，可能触发验证码（算术题）。URL 带 ?expired=1 时显示「会话已过期」提示。

### 安装向导（/install）
系统首次启动时的初始化流程，分 5 步：选择语言 → 配置域名 → 创建管理员账号 → 配置数据库 → 配置缓存。仅系统未初始化时可访问。

`

const SystemSkillGoRavenUserUIEn = `---
name: goraven-user-ui
description: GoRaven user-side frontend UI and operations guide. Details the layout, features, and usage of pages accessible to regular users, helping the model guide users through daily operations.
---

# GoRaven User-Side Frontend UI

## Overall Layout

GoRaven uses a left sidebar + right main content layout. The sidebar is collapsible to icon-only mode.

### Sidebar (Desktop)

- **Top**: GoRaven logo + collapse/expand toggle button
- **New Chat button**: Navigate to a blank chat page to start a new AI interaction
- **Workspace navigation** (collapsible groups):
  - Dashboard: View personal usage statistics
  - File Manager: Manage personal workspace files
  - Skills Center: Manage installed skills, browse the skill marketplace
  - Personas: Create and manage AI personas (pre-configured system prompts with models, MCP tools, and skills)
- **Session history**: Grouped by time (Today / Past 7 days / Past 30 days), infinite scroll loading. Each session can be renamed or archived (deleted).
- **Bottom user area**: Shows avatar and username. Click opens a popover menu with:
  - Profile
  - Admin Mode (admin-only, switches to backend management)
  - Logout

### Mobile

The sidebar is hidden by default, accessible via a hamburger menu button in the top toolbar, appearing as an overlay.

## Chat Page (Core Interaction)

The chat page is the main AI interaction interface, with two states.

Users can run multiple conversations simultaneously. Click any session in the sidebar history to switch — each session maintains its own message history, model configuration, and generation state independently. If you switch from session A (still generating) to session B, session A continues in the background without losing progress.

### New Chat (/chat, no session)

**Top toolbar**:
- Model/Persona selector dropdown with two sections: Models and Personas. Selecting a persona auto-configures the model, MCP tools, and skills. Selecting a model clears the persona selection.
- If no models are configured, a "No model" warning button appears with a dialog explaining setup steps.
- Menu button on mobile.

**Initial state**: Displays the GoRaven logo centered with "What can I help you build?".

**Input area** (full NewChatInput) is a rounded container with:
- **Text input**: Supports multi-line, auto-grows (max ~4 lines). Enter to send (Shift+Enter for newline). IME input (e.g. Chinese) is handled correctly — Enter while composing won't accidentally send.
- **Config button (+ wrench icon)**: Opens a floating panel to select MCP tools and skills for this conversation. Selected items show highlighted. Disabled when a persona is selected (personas pre-configure tools/skills).
- **File upload button (paperclip icon)**: Select local files, up to 10. Uses chunked upload for large files. Image files show preview thumbnails. Status indicators for upload progress/success/failure. Click to remove uploaded files.
- **Project directory button (folder icon)**: Select a subdirectory under projects/ as the workspace for this chat. Highlighted when selected. Has a "Clear selection" option. If the selected project is a team shared project (whether directly selecting your own shared project or entering via the team project list), the system locks the project — if another session is already using it, an error "the shared project is currently in use by another session" is shown.
- **Thinking toggle**: Enables/disables the model's deep thinking mode. Can be toggled freely before sending, locked after sending.
- **Send/Stop button**: Up-arrow icon, highlight color. Greyed out when input is empty. After sending, becomes a stop button (circle stop icon). There is a ~3 second cooldown after sending during which the stop button is temporarily disabled (to prevent accidental cancellation). After the cooldown, you can stop generation at any time. Stopped content is saved as a truncated message.

**Stop & Background Running**:
- Users can click the stop button at any time (after the 3s cooldown) to interrupt generation. Partial output is preserved in the conversation.
- If generation exceeds 3 minutes without completing, the frontend automatically disconnects the SSE stream, switches the page to "background thinking" status, and begins polling. The user can switch to other sessions while the current one continues running in the background.
- Once background generation completes, the sidebar session status returns to normal and the message list auto-refreshes with full results.
- During background running, that session's input box is disabled and a status indicator is shown at the top.

**After sending**: The page auto-navigates to /chat/:newSessionId, switching to session chat mode. The top toolbar changes to session mode.

### Session Chat (/chat/:sessionId, existing session)

**Loading**: Shows a spinner while loading, then displays the message list.

**Top toolbar**:
- Mobile menu button
- Click persona name/model name to open a detail dialog showing the current persona/model's full configuration (name, icon, description, model, MCP tool list, skill list, token usage breakdown)
- Project path indicator (if set)
- Token usage display (e.g. "12K / 128K"), updating in real time
- Context compress button: Manually trigger context compression to free up token space
- Share button: Generate a shareable session link (configurable expiry and share type)

**Message area**: Displays messages in chronological order (user/assistant/tool/summary roles). Assistant messages stream in real time via SSE, including reasoning content and the actual reply. Tool invocations are shown as cards. Each assistant message has action buttons at the bottom: Copy (one-click copy of reply content), Like, and Dislike for providing feedback on response quality. The message list auto-scrolls to the latest content; scrolling up to review history will not force-scroll back to the bottom.

**Bottom input** (simplified ChatInput, no + config button):
- File upload and Thinking toggle same as above, but config is locked once a session is created.
- Input disabled when session is running in background (status "background thinking").
- Send button styles change during compression.
- AI disclaimer text at the bottom.

**Model failure retry**: On model failure, the system auto-retries with backoff. The frontend receives retry events.

## Dashboard (/dashboard)

Displays personal usage statistics for the current user.

- **Token Overview**: Cards showing total tokens, weekly tokens, daily tokens, session count, new session count.
- **Storage**: Used/free/total with progress bar, broken down by directory (documents, projects, etc.).
- **Token Trend Chart**: Stacked bar chart (prompt + completion tokens), toggleable 7/30/90 days. Shows daily average and peak day.
- **Model Usage**: Horizontal bar chart showing token share per model.
- **Rankings** (three columns): Weekly Tool Usage Rank, Weekly MCP Usage Rank, Weekly Skill Usage Rank.

Refresh button at the top.

## Files Page (/files)

Manages the user's personal workspace files.

**Toolbar**:
- Navigate to parent directory
- Current directory name
- Upload files
- New folder
- Delete selected (default directories and read-only files cannot be deleted)
- Compress selected (create zip archive)
- Decompress (only enabled when a single .zip file is selected; option to extract to subdirectory)
- Refresh

**File list**: Table with columns: checkbox (select all/individual, supports Shift range selection), name (double-click to enter directories), size, modified time.

**Keyboard shortcuts**: F2 to rename selected file, Delete or Backspace to delete selected items, Shift+click for range selection.

**Right-click menu**: Preview, Download, Rename, Delete, Share to Team (only for subdirectories under projects/), Unshare (only for shared directories).

**Sharing restrictions**: Shared subdirectories under projects/ cannot be deleted or renamed — unshare first.

**Preview dialog** supports the following file types:

| Type | Notes |
|------|-------|
| Images (png/jpg/gif/svg/webp etc.) | Embedded display, supports fullscreen |
| Video (mp4/webm/mov etc.) | Inline player with controls |
| Audio (mp3/wav/ogg etc.) | Inline player |
| PDF | Iframe embed |
| Text (txt/json/yaml/code etc.) | Scrollable plain text, large files auto-truncated |
| Markdown (md/markdown) | Rendered as HTML with headings, tables, code highlighting, charts |
| Excel (xlsx) | Rendered as HTML table, multi-sheet tabs, highlighted headers |
| HTML (html/htm) | Rendered in sandboxed iframe, JS executes but isolated from parent |

**HTML preview notes**:
- CSS, JS, and images referenced in HTML must use **relative paths** (e.g. ./assets/style.css), not root-absolute paths (e.g. /assets/style.css). Relative paths auto-resolve to URLs with a temporary access token, ensuring resources load; root-absolute paths lose the token and cause 404 and CORS errors.
- For Vite or similar build output, set base to ./ (or pass --base=./ on CLI) so build artifacts use relative paths.
- SPAs depending on localStorage / sessionStorage cannot run in sandbox preview (browser security restriction).

**Drag-and-drop**: Files can be dragged onto the page area to upload.

Fixed user workspace directories (cannot be deleted):

| Directory | Purpose |
|-----------|---------|
| documents/ | Document files, readable and writable by AI |
| projects/ | Project code; subdirectories can be selected as chat workspace |
| images/ | Image file storage |
| videos/ | Video file storage |
| downloads/ | Default download location |
| temp/ | Temporary files (attachment parsing results, intermediate artifacts), AI can read/write |
| skills/ | Installed skill files (hidden in the file manager — skills are managed and authorized through the database) |

> **File manager hidden rules**: The "skills/" directory is not shown at root level (skills are managed via database, file operations are disallowed); files and directories starting with "." (e.g. .DS_Store, .profile) are also not shown — though .profile can still be edited via the "Environment Variables" button in the toolbar.

Of these, documents/, projects/, and temp/ are the directories most commonly operated on by the AI Agent during task execution.

### Environment Variables (.profile)

The .profile file at the workspace root manages user environment variables. **Both user and AI can read and write it.**

**File format**: One KEY=VALUE per line, # for comments. Also supports export KEY=VALUE (export prefix auto-stripped).

**User operations** (/files page):
- Toolbar "Environment Variables" button ({} icon) opens an editor dialog for add/modify/delete key-value pairs, auto-saved to .profile

**AI operations** (via file tools):
- Read/write .profile directly from workspace root using read_file / write_file tools
- When executing shell commands, the system **auto-injects all .profile variables into the command environment** — no source/export needed, reference with $VAR_NAME directly in commands
- On auth failure or missing env var, AI inspects .profile first, confirms with user before writing — **never guess values**

**Injection rule**: .profile vars are injected last, overriding same-name system environment variables.

**Typical scenarios**:
- User sets BRAVE_SEARCH_API_KEY=xxx in the Files page env vars dialog; AI can then use $BRAVE_SEARCH_API_KEY in all subsequent conversations
- AI discovers it needs GITHUB_TOKEN and prompts: "I need GITHUB_TOKEN. Please add it in the Files page Environment Variables, or tell me the value and I'll write it to .profile."

### Team Projects

The "Team Projects" button in the top-right of the Files page switches to the team project list view, showing all projects shared by team members.

**Team project list**:
- Displays all shared projects as a card grid. Each card shows the project name, owner (avatar + name), description, and update time.
- Click a card to enter the project file browser.
- Project owners see a "More" menu on their cards (on hover) for editing the description and unsharing.
- A "My Files" button in the top-right returns to personal file management.

**Project file browser**:
- Inside a shared project, the interface is consistent with the Files page: back button, current directory name, upload, new folder, delete, compress, decompress, refresh.
- Supports full file operations: upload, create, delete, rename, compress, decompress, preview, and download.
- A "My Files" button in the top-right returns directly to personal file management.
- The back button at the project root returns to the team project list.

## Skills Page (/skills)

Manage user skills across three tabs:

| Tab | Content |
|-----|---------|
| Installed (with count) | User's skill list. Each item shows icon, name, description, Always-On toggle, install date. Click a row to open detail drawer (edit icon/category, view SKILL.md content, toggle Always-On, share custom skills, delete) |
| Marketplace (with count) | Available skills from the marketplace, each with install button and install count. Already installed skills show "Installed" badge. **Note**: clicking Install only copies files and registers the skill. Dependency installation (pip/npm install, etc.) must be done in a conversation — a dialog provides a prompt to copy into a new chat for the AI to complete. This can be time-consuming. |
| Team Shares | Skills shared by other users. Install or cancel share (for own shares) |

Top bar has a search box and category filter dropdown. Sync (refresh) button in the top-right corner.

**Always-On**: Skills with Always-On enabled are **automatically loaded in every new conversation** without needing to manually check them in the configuration panel. Ideal for frequently used skills (e.g., code formatters, common utilities). Can be toggled at any time from the installed list or detail drawer. Enabling too many Always-On skills consumes context space in every conversation — reserve for high-frequency use only.

## Personas Page (/personas)

Create and manage AI personas (pre-configured system prompts + model + MCP + skills).

**Persona list (/personas)**: "New Persona" button at top. Each card shows icon, name, category, role info snippet, MCP/skill tags, model name. Click to enter detail view. Right-side menu for edit/delete.

**Persona detail/edit**: Editable fields include name, icon, category, persona settings (textarea, max 500 chars), model config, MCP tools (searchable checkbox list), skills (searchable checkbox list). When creating, can choose from template library (grouped by category, searchable). Icons are selected from a preset icon picker.

Save/Cancel buttons fixed at top.

## Profile Page (/profile)

Personal settings:
1. **Personal Info**: Avatar (uploadable), nickname (editable), username (read-only), role badge, email (editable), registration date.
2. **Appearance**: Theme mode selector (Light / Dark / System).
3. **Account Security**: Change password (requires current password, new password must be at least 8 chars with letters and numbers).
4. **About**: Project description + Logout.

## Other Pages

### Shared Chat View (/share/:shareId)

Public links generated by the session sharing feature. Recipients visit this page to view the shared conversation.

- **Public shares**: Viewable without login. Page displays share title, sharer name, share type, view count, and expiration time.
- **Internal shares**: Require login to view message content.
- **Message display**: Read-only conversation history including reasoning content and reply text. Copy, Like, and Dislike buttons at the bottom of each message.
- **Error states**:
  - Not found (notFound)
  - Expired (expired): Shows expiry notice
  - Load failed (loadFailed): Shows retry prompt
- Footer with GoRaven branding and GitHub link.

### Login (/login)
Username and password input. May trigger a captcha (simple arithmetic). URL with ?expired=1 shows "Session expired" message.

### Install Wizard (/install)
Initial system setup wizard (only accessible when system is not initialized). 5 steps: language → domain → admin account → database → cache.

`
