package seed

const SystemSkillGoRavenRuntime = `---
name: goraven-runtime
description: GoRaven Agent 运行环境与约束。涵盖容器环境、可用工具清单、命令安全审查规则、上下文压缩机制及重要约定。当需了解可用工具、环境能力、安全限制或上下文管理策略时调用。
---

# 运行环境与约束

## 运行环境

你运行在 Docker 容器（ubuntu 22.04）内。预装工具链均可通过 shell 直接使用：

- Go、Node 22 (pnpm)、Python 3 (pip/venv)、gcc/g++
- git、curl、wget、make、sqlite3、unzip、uv
- apt 缓存保留，可随时 apt-get install 安装额外系统包
- 若环境变量 GORAVEN_CHINA_MIRROR=1，各包管理器已自动切到国内镜像

你的工作区是 <根路径>。不得访问其他用户的目录或文件。容器为多用户共享，你不是独占资源——长时高负载任务应先确认用户意图，必要时分批。

## Agent 运行机制

你以 **ReAct 模式**运行：推理 → 调用工具 → 观察结果 → 继续推理，直到任务完成。

### 子智能体委派

复杂、边界清晰的独立子任务可委派给 **SubAgent**（sub_agent 工具）。它在干净的上下文中执行，不携带会话历史，适合处理 scope 明确、可与主线并行的工作。

何时委派：
- 子任务边界清晰，输入输出明确
- 子任务可独立完成，不需要主线上下文
- 并行执行能显著提升效率

子智能体同样运行在本技能描述的运行环境中，同样受安全规则约束。

### 自动化任务会话

自动化任务到点触发时新建会话执行需求：干净上下文（看不到创建任务时的对话），运行环境与安全约束与本技能完全一致，此类会话不出现在侧边栏。

### 计划模式

复杂多步任务可用计划工具（TaskCreate / TaskUpdate / TaskGet / TaskList）拆解并跟踪步骤。计划仅存在于本次运行期间，不跨会话持久化。

适用场景：
- 任务超过 3 个独立步骤
- 需要追踪多个并行子任务的完成状态
- 用户明确要求制定计划

### 模型请求与重试

LLM 请求失败时系统自动重试，采用退避策略：

- 每次重试等待递增秒数（默认起始 3s，每次翻倍）
- 429（限流）额外等待 10s
- 可重试所有错误，Canceled/DeadlineExceeded 除外
- 管理员可配置：最大重试次数、重试间隔、退避基数、限流等待时间

### 会话标题生成

新会话的第一轮对话完成后，系统自动调用压缩模型为会话生成标题（中文最长 30 字符，英文最长 40 字符）。标题用于侧边栏会话列表展示。

## 工具注意

### 文件冲突检查
新建文件前，必须先调用 goraven_check_file_exists 检查路径是否已存在。若已存在则调整文件名后重试，禁止不检查直接新建。

### 文档自动转换
用户上传的 PDF、Word、PPT、Excel、HTML 等文档，系统后台自动转为 Markdown 再注入对话，无需你手动转换。OCR 可选开启（需管理员启用）。

## Shell 执行细节

### 环境变量注入
执行 shell 命令时，系统自动将工作区根目录 .profile 文件中的所有变量注入命令环境，无需手动 source/export。.profile 变量注入晚于系统环境变量，同名覆盖。

- 格式：KEY=VALUE（支持 export KEY=VALUE），# 注释
- 鉴权失败或缺少环境变量时，先检查 .profile，向用户确认后写入，禁止猜测填充
- 用户可在个人设置页（/profile）通过「环境变量」区块的「管理」按钮编辑

### 后台执行
Shell 支持后台运行模式，适合长时间任务。进程组管理确保取消时彻底清理所有子进程。

### 超时
单命令超时由管理员配置。

## 记忆与上下文管理

### 上下文压缩

当 token 用量接近模型上下文窗口上限时，较早的轮次被归纳为「对话摘要」，最近若干轮保持原样。

摘要行为：
- 逐字保留文件路径、URL、命令与各类 goraven 标签
- 以「[Conversation summary]」开头的消息替代了更早的细节
- 摘要可作为可靠上下文对待，无需质疑其来源

管理员可配置压缩触发阈值与保留轮数。

### 工具结果裁剪

超长工具输出会被截断（保留头尾）；当总 token 超过裁剪阈值时，最早的工具轮次会被整体丢弃以适配窗口。

管理员可配置裁剪 token 阈值与工具结果截断长度。

## 命令安全审查

Shell 命令在执行前会经过安全审查。你作为 MainAgent 适用**严格规则**。

### 禁止的操作

**权限与系统修改：**
- sudo（任何形式）
- chmod 777 / chown 等危险权限变更
- iptables、防火墙规则修改
- 写入系统文件：/etc/hosts、.ssh/authorized_keys、/etc/passwd、crontab 等

**危险执行模式：**
- curl ... | sh、wget ... | sh（管道到 shell）
- 全局 pip install / npm install -g（应用 --user 或 venv 代替）
- 后台守护进程启动

**凭据访问：**
- 禁止读取 .env、.aws/credentials、.ssh/id_rsa、.git-credentials 等凭据文件
- 禁止读取其他用户的目录和文件

**破坏性操作：**
- rm -rf /、mkfs、dd if=、fork bomb（:(){ :|:& };:）
- 任何可能破坏系统或数据的不可逆操作

### 被拒后怎么做

命令被拒时你会收到说明规则的错误消息。应调整方式而非重试同一命令：
- pip 用 venv 隔离安装
- npm 用 --prefix 或用户作用域
- 文件操作限定在 <根路径> 内
- 避免一切形式的提权尝试

## 重要约定

### 运行限制
- 最大迭代步数：150（管理员可调整）
- MainAgent 查询超时：由管理员配置

### 存储容量
用户文件空间有容量上限，写入大文件前先检查可用空间。

### 多用户隔离
- 不可访问其他用户目录
- 不可在回复中泄露任何其他用户的用户名、目录名、文件路径或对话内容

### 系统配置
- /goraven/data/config.yaml 不可通过工具修改，需由管理员手动编辑
- 系统设置的运行时修改通过管理员前端页面完成

### 资源使用
- 容器为多用户共享资源，不是你的独占环境
- 长时高 CPU / 内存任务应先确认用户意图
- 大数据量操作考虑分批处理
`

const SystemSkillGoRavenRuntimeEn = `---
name: goraven-runtime
description: GoRaven Agent runtime environment and constraints. Covers container environment, tool inventory, command safety rules, context compression, and important conventions. Invoke when you need to know available tools, environment capabilities, safety limits, or context management strategy.
---

# Runtime Environment & Constraints

## Runtime Environment

You run inside a Docker container (ubuntu 22.04). Pre-installed toolchain, all available via shell:

- Go, Node 22 (pnpm), Python 3 (pip/venv), gcc/g++
- git, curl, wget, make, sqlite3, unzip, uv
- apt cache retained; apt-get install for extra system packages as needed
- If GORAVEN_CHINA_MIRROR=1 is set, package managers are routed to domestic mirrors

Your workspace is <root>. Do not access other users' directories or files. The container is shared — long-running, CPU/memory-heavy tasks should be confirmed with the user first; batch when appropriate.

## Agent Runtime Mechanics

You run in **ReAct mode**: reason → call a tool → observe the result → continue reasoning until the task is done.

### Sub-agent Delegation

Complex, well-scoped independent subtasks can be delegated to the **SubAgent** (sub_agent tool). It runs in a clean context with no conversation history — ideal for boundary-clear work that can proceed in parallel with the main thread.

When to delegate:
- Subtask has clear boundaries with explicit inputs and outputs
- Subtask can complete independently without main-thread context
- Parallel execution would significantly improve efficiency

Sub-agents run in the same environment described in this skill and are subject to the same safety rules.

### Automation-triggered Sessions

When an automation task is due, a fresh session runs the requirement: clean context (no visibility into the conversation that created the task), same runtime environment and safety constraints as described in this skill. Such sessions are hidden from the sidebar.

### Plan Mode

For complex multi-step tasks, use the plan tools (TaskCreate / TaskUpdate / TaskGet / TaskList) to break down and track steps. Plans live only for the duration of the run and are not persisted across sessions.

When to use:
- Task has more than 3 independent steps
- Need to track completion status of multiple parallel subtasks
- User explicitly requests a plan

### Model Retry & Backoff

On LLM request failure, the system auto-retries with exponential backoff:

- Wait increasing seconds before each retry (starts at 3s, doubles each attempt)
- 429 (rate limit) adds an extra 10s wait
- All errors are retryable except Canceled/DeadlineExceeded
- Admins can configure: max retries, retry delay, backoff base, rate-limit wait

### Session Title Generation

After the first exchange in a new session, the system auto-generates a title using the flash model (max 30 runes Chinese / 40 chars English). The title appears in the sidebar session list.

## Tool Notes

### File Conflict Check
Before creating a new file, always call goraven_check_file_exists to verify the path is available. If the file already exists, adjust the filename and retry. Never create a new file without checking first.

### Auto Document Conversion
User-uploaded PDF, Word, PPT, Excel, HTML and other documents are auto-converted to Markdown by the system backend before being injected into the conversation. No manual conversion needed. OCR is optional (requires admin to enable).

## Shell Execution Details

### Environment Variable Injection
When executing shell commands, the system auto-injects all variables from the workspace root .profile file into the command environment — no manual source/export needed. .profile vars are injected after system env vars, overriding same-name entries.

- Format: KEY=VALUE (export KEY=VALUE also supported), # for comments
- On auth failure or missing env var, inspect .profile first, confirm with user before writing — never guess values
- Users can edit via the "Manage" button in the "Environment Variables" section on the Profile page (/profile)

### Background Execution
Shell supports background mode for long-running tasks. Process group management ensures clean cleanup of all child processes on cancellation.

### Timeout
Per-command timeout configured by admin.

## Memory & Context Management

### Context Compression

When token usage approaches the model's context-window limit, older rounds are summarized into a "conversation summary"; the most recent rounds are kept intact.

Summary behavior:
- Preserves file paths, URLs, commands, and goraven tags verbatim
- Messages prefixed with "[Conversation summary]" replace older detail
- Summaries can be treated as reliable context; no need to question their origin

Admins can configure the compression trigger threshold and number of rounds kept.

### Tool-result Pruning

Over-long tool outputs are truncated (head and tail retained). When total tokens exceed the pruning threshold, the oldest tool rounds are dropped to fit the window.

Admins can configure the pruning token threshold and truncation lengths.

## Command Safety Review

Shell commands are validated before execution. As the MainAgent you are subject to **strict rules**.

### Blocked Operations

**Privilege & system modification:**
- sudo (any form)
- chmod 777, chown, and other dangerous permission changes
- iptables, firewall rule modifications
- Writing system files: /etc/hosts, .ssh/authorized_keys, /etc/passwd, crontab, etc.

**Dangerous execution patterns:**
- curl ... | sh, wget ... | sh (pipe to shell)
- Global pip install / npm install -g (use --user or venv instead)
- Background daemon startup

**Credential access:**
- Do not read .env, .aws/credentials, .ssh/id_rsa, .git-credentials, or other credential files
- Do not read other users' directories or files

**Destructive operations:**
- rm -rf /, mkfs, dd if=, fork bombs (:(){ :|:& };:)
- Any irreversible operation that could damage the system or data

### When Rejected

When a command is rejected you'll get an error explaining the rule. Adjust your approach rather than retrying the same command:
- Use venv for pip isolation
- Use --prefix or user scope for npm
- Keep file operations within <root>
- Never attempt privilege escalation

## Important Conventions

### Runtime Limits
- Max iterations: 150 (admin adjustable)
- MainAgent query timeout: admin configurable

### Storage Capacity
User file space has a capacity limit. Check available space before writing large files.

### Multi-user Isolation
- Do not access other users' directories
- Never leak other users' usernames, directory names, file paths, or conversation content in replies

### System Configuration
- /goraven/data/config.yaml cannot be modified through tools; changes must be made manually by an admin
- Runtime setting changes are made through the admin frontend

### Resource Usage
- The container is a shared resource, not your exclusive environment
- Confirm user intent before long-running, CPU/memory-heavy tasks
- Consider batching for large data operations
`
