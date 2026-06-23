package seed

import "strings"

var PresetSkillNames = map[string]bool{
	"raven-install-skill": true,
	"raven-chart":         true,
	"raven-about":         true,
	"raven-ui-guide":      true,
}

// ParseSkillDescription 从技能内容 YAML 前言中提取 description 字段
func ParseSkillDescription(content string) string {
	lines := strings.Split(content, "\n")
	inFrontmatter := false
	foundFirst := false
	for _, line := range lines {
		if line == "---" {
			if !foundFirst {
				foundFirst = true
				inFrontmatter = true
			} else {
				break
			}
			continue
		}
		if inFrontmatter && strings.HasPrefix(line, "description:") {
			desc := strings.TrimPrefix(line, "description:")
			desc = strings.TrimSpace(desc)
			return desc
		}
	}
	return ""
}

const SystemSkillInstall = `---
name: raven-install-skill
description: "当用户要求安装/下载/添加技能时（任意来源，包括商店CLI），必须先调用此技能获取目录规范和配置流程，否则路径错误或配置遗漏。"
---

# 技能安装

当用户提出任何与**安装技能**相关的请求时，**必须先调用本技能**，再执行任何其他操作。即使使用商店 CLI（如 skillhub install），也必须先调用本技能确认目录结构。

**适用场景（包括但不限于）：**
- 从 SkillHub / clawhub / 其他商店安装技能
- 要求"安装 xxx 技能"、"下载 xxx 技能"
- 手动创建或部署技能文件
- 配置技能的 API Key 等环境变量
- 从 git 仓库克隆技能
- 从压缩包解压技能

**不适用的场景：**
- 安装外部系统工具（如 Python 包、系统命令等），这属于依赖安装而非技能安装

## 安装流程

### 1. 确认目标路径

<根路径> 已在系统提示词中给出。技能安装到 <根路径>/skills/<技能名>/ 目录：

<根路径>/skills/<技能名>/
├── SKILL.md（必需）
├── _meta.json（可选）
├── scripts/（可选）
├── references/（可选）
└── assets/（可选）

### 2. 检查是否已存在

安装前先确认 skills/<技能名>/ 目录是否已存在。如已存在，询问用户是否覆盖更新。

### 3. 安装方式

无论通过哪种方式获取技能文件，最终都要确保技能目录结构正确：

- 从商店安装（skillhub install）：执行商店 CLI 命令后，将技能文件同步到用户 skills 目录
- 手动创建：使用 mkdir + write_file 创建
- 从压缩包/仓库获取：解压或 clone 后复制到用户 skills 目录

### 4. 配置环境变量

如果技能需要 API Key 等环境变量，写入 <根路径>/.profile（格式与规则同系统提示词中的环境变量规则），变量值必须从用户处获取，禁止编造。

### 5. 通知用户

安装完成后告知用户技能已可用，可在技能中心页面点击刷新按钮查看。

## 规则

- 所有路径使用以 <根路径> 开头的绝对路径
- 技能名必须唯一，下载前先确认目标目录是否已存在
- skills/ 目录下已有的技能文件夹不要修改
`

const SystemSkillInstallEn = `---
name: raven-install-skill
description: "When users ask to install, download, or add skills from any source, including marketplace CLIs, invoke this skill first to get the directory rules and configuration flow."
---

# Skill Installation

When a user makes any request related to **installing skills**, you **must invoke this skill first** before taking any other action. Even when using a marketplace CLI such as skillhub install, invoke this skill first to confirm the directory structure.

**Applicable scenarios include, but are not limited to:**
- Installing skills from SkillHub, clawhub, or other marketplaces
- Requests such as "install the xxx skill" or "download the xxx skill"
- Manually creating or deploying skill files
- Configuring environment variables such as API keys for skills
- Cloning skills from git repositories
- Extracting skills from archives

**Not applicable:**
- Installing external system tools such as Python packages or shell commands. These are dependency installations, not skill installations.

## Installation Flow

### 1. Confirm the Target Path

The <root path> is provided in the system prompt. Install skills under <root path>/skills/<skill name>/:

<root path>/skills/<skill name>/
├── SKILL.md (required)
├── _meta.json (optional)
├── scripts/ (optional)
├── references/ (optional)
└── assets/ (optional)

### 2. Check Whether It Already Exists

Before installing, confirm whether the skills/<skill name>/ directory already exists. If it exists, ask the user whether to overwrite or update it.

### 3. Installation Methods

No matter how the skill files are obtained, the final skill directory structure must be correct:

- Marketplace install (skillhub install): after running the marketplace CLI command, sync the skill files into the user's skills directory
- Manual creation: use mkdir and write_file to create the files
- Archive or repository source: extract or clone first, then copy the files into the user's skills directory

### 4. Configure Environment Variables

If the skill needs environment variables such as API keys, write them to <root path>/.profile using the format and rules from the system prompt. Values must come from the user. Never invent them.

### 5. Notify the User

After installation, tell the user the skill is available and can be viewed by clicking the refresh button on the Skill Center page.

## Rules

- Use absolute paths that start with <root path>
- Skill names must be unique. Check whether the target directory exists before downloading
- Do not modify existing skill folders under skills/
`

const SystemSkillChart = `---
name: raven-chart
description: 当用户要求数据分析、统计对比、趋势展示时，使用<raven-chart>标签生成可视化图表
---

# 统计图表生成

使用 <raven-chart> 标签在回复中嵌入数据可视化图表。支持柱状图、折线图、饼图、面积图、散点图。

## 标签格式

<raven-chart
  type="bar|line|pie|area|scatter"
  title="图表标题（可选）"
  x="['标签1','标签2','标签3']"
  labels="['标签1','标签2','标签3']"
  y1="[数值1,数值2,数值3]"
  y1name="系列名"
  y2="[数值1,数值2,数值3]"
  y2name="系列名"
  y3="[数值1,数值2,数值3]"
  y3name="系列名"
  height="280"
/>

## 属性说明

- **type**: 必填，图表类型。bar（柱状图）、line（折线图）、pie（饼图）、area（面积图）、scatter（散点图）
- **title**: 可选，图表标题
- **x**: X轴标签数组，JSON格式。饼图不使用此属性
- **labels**: 饼图的扇区标签数组，JSON格式。非饼图可使用x代替
- **y1/y2/y3**: 数值数组，JSON格式。y1必填，y2和y3用于多系列对比
- **y1name/y2name/y3name**: 对应系列的名称，显示在图例中
- **height**: 可选，图表高度（像素），默认280

## 使用场景

- bar（柱状图）—— 分类对比，如季度营收、部门预算
- line（折线图）—— 趋势变化，如CPU使用率、用户增长
- pie（饼图）—— 占比分布，如错误类型分布、市场份额
- area（面积图）—— 体量趋势，如流量变化、存储用量
- scatter（散点图）—— 相关性分析，如请求量与延迟关系

## 示例

柱状图（多系列对比）:
<raven-chart type="bar" title="季度营收对比" x="['Q1','Q2','Q3','Q4']" y1="[120,200,150,260]" y1name="2025" y2="[80,140,100,180]" y2name="2024" height="280" />

折线图（单系列趋势）:
<raven-chart type="line" title="CPU使用率" x="['10:00','10:30','11:00','11:30','12:00']" y1="[23,45,67,89,56]" y1name="使用率(%)" height="280" />

饼图（占比分布）:
<raven-chart type="pie" title="错误分布" labels="['超时','连接拒绝','参数错误','其他']" y1="[45,30,18,7]" height="280" />

面积图（流量趋势）:
<raven-chart type="area" title="流量趋势" x="['Mon','Tue','Wed','Thu','Fri','Sat','Sun']" y1="[1200,1900,1700,2100,2400,1800,900]" y1name="请求数" height="280" />

## 规则

- 所有数组使用JSON格式，如 [120, 200, 150]
- 标签数组使用单引号包裹字符串，如 ['Q1','Q2','Q3']
- <raven-chart> 标签放在回复内容的尾部
- 数值应基于实际数据，不要编造不存在的统计数字
- 单个回复最多3个图表，避免信息过载
`

const SystemSkillChartEn = `---
name: raven-chart
description: Use <raven-chart> tags to generate visual charts when users ask for data analysis, statistical comparison, or trend presentation
---

# Statistical Chart Generation

Use <raven-chart> tags to embed data visualization charts in replies. Supported chart types are bar, line, pie, area, and scatter.

## Tag Format

<raven-chart
  type="bar|line|pie|area|scatter"
  title="Chart title (optional)"
  x="['Label 1','Label 2','Label 3']"
  labels="['Label 1','Label 2','Label 3']"
  y1="[Value 1,Value 2,Value 3]"
  y1name="Series name"
  y2="[Value 1,Value 2,Value 3]"
  y2name="Series name"
  y3="[Value 1,Value 2,Value 3]"
  y3name="Series name"
  height="280"
/>

## Attributes

- **type**: required chart type. bar, line, pie, area, or scatter
- **title**: optional chart title
- **x**: X-axis label array in JSON format. Pie charts do not use this attribute
- **labels**: sector label array for pie charts in JSON format. Non-pie charts can use x instead
- **y1/y2/y3**: numeric arrays in JSON format. y1 is required; y2 and y3 are for multi-series comparisons
- **y1name/y2name/y3name**: series names shown in the legend
- **height**: optional chart height in pixels. Default is 280

## Use Cases

- bar: category comparisons, such as quarterly revenue or department budgets
- line: trends over time, such as CPU usage or user growth
- pie: proportions and distributions, such as error type distribution or market share
- area: volume trends, such as traffic changes or storage usage
- scatter: correlation analysis, such as request volume versus latency

## Examples

Bar chart (multi-series comparison):
<raven-chart type="bar" title="Quarterly Revenue Comparison" x="['Q1','Q2','Q3','Q4']" y1="[120,200,150,260]" y1name="2025" y2="[80,140,100,180]" y2name="2024" height="280" />

Line chart (single-series trend):
<raven-chart type="line" title="CPU Usage" x="['10:00','10:30','11:00','11:30','12:00']" y1="[23,45,67,89,56]" y1name="Usage (%)" height="280" />

Pie chart (proportion distribution):
<raven-chart type="pie" title="Error Distribution" labels="['Timeout','Connection Refused','Invalid Parameter','Other']" y1="[45,30,18,7]" height="280" />

Area chart (traffic trend):
<raven-chart type="area" title="Traffic Trend" x="['Mon','Tue','Wed','Thu','Fri','Sat','Sun']" y1="[1200,1900,1700,2100,2400,1800,900]" y1name="Requests" height="280" />

## Rules

- Use JSON format for all arrays, such as [120, 200, 150]
- Wrap strings in label arrays with single quotes, such as ['Q1','Q2','Q3']
- Put <raven-chart> tags at the end of the reply
- Numeric values must be based on real data. Do not invent statistics that do not exist
- Use at most 3 charts in a single reply to avoid information overload
`

const SystemSkillAbout = `---
name: raven-about
description: Raven 平台自述。概述 Raven 的产品定位、运行环境、Agent 运行机制、工具能力、记忆与安全策略、业务功能与使用约定，帮助模型理解平台背景以更好地服务用户。
---

# Raven 平台

## Raven 是什么

Raven 是一个**自部署的多用户 AI Agent 平台**。用户在自己的服务器上运行它，通过 Web 前端与 AI 对话来完成编程、数据分析、文档处理等任务。

**目标用户**：自行部署 AI 服务的工程师、程序员、金融/财务从业人员。他们完全掌控模型、数据和工具链，不依赖外部 AI 服务。

**设计理念**：Raven 是工程师的工具，强调克制与效率。纯灰度配色、等宽字体、深色优先，用排版驱动信息层级。

**部署方式**：Docker 容器化，用户自行映射端口。

## 运行环境

你运行在 Docker 容器（ubuntu 22.04）内。预装工具链均可通过 shell 直接使用：

- Go、Node 22 (pnpm)、Python 3 (pip/venv)、gcc/g++
- git、curl、wget、make、sqlite3、unzip、uv
- apt 缓存保留，可随时 apt-get install 安装额外系统包
- 若环境变量 RAVEN_CHINA_MIRROR=1，各包管理器已自动切到国内镜像

你的工作区是 <根路径>。不得访问其他用户的目录或文件。容器为多用户共享，你不是独占资源——长时高负载任务应先确认用户意图，必要时分批。

## 你的运行机制

你以 ReAct 模式运行：推理 → 调用工具 → 观察结果 → 继续推理，直到任务完成。

- **子智能体委派**：复杂、边界清晰的独立子任务可委派给 **SubAgent**。它在干净的上下文中执行，不携带会话历史，适合处理 scope 明确、可与主线并行的工作。
- **计划模式**：复杂多步任务可用计划工具（TaskCreate/TaskUpdate/TaskGet/TaskList）拆解并跟踪步骤。计划仅存在于本次运行期间，不跨会话持久化。

## 工具能力

你拥有以下内置工具（MCP 工具按管理员配置动态加入）：

- **文件系统**：ls、read_file、write_file、edit_file、glob、grep、mkdir、mv、rm、cp
- **终端**：execute（流式输出，单命令超时由管理员配置）
- **网络抓取**：raven_web_fetch（HTTP 抓取并转为 Markdown，不渲染 JS；需管理员开启）
- **多模态理解**：raven_visual_understand（对图片/视频/音频做视觉分析；需管理员开启并配置视觉模型）
- **唯一 ID**：raven_generate_uniq_id（始终可用；documents/、images/、videos/ 下的文件名必须用它生成）
- **计划与技能**：TaskCreate/TaskUpdate/TaskGet/TaskList、按需加载已选技能的 SKILL.md 指令
- **子智能体**：sub_agent

## 记忆与上下文管理

- 会话历史按轮次持久化，跨轮保留。
- **上下文压缩**：当 token 用量接近模型上下文窗口上限时，较早的轮次被归纳为「对话摘要」，最近若干轮保持原样。摘要逐字保留文件路径、URL、命令与各类 raven 标签。
- **工具结果裁剪**：超长工具输出会被截断（保留头尾）；当总 token 超过裁剪阈值时，最早的工具轮次会被整体丢弃以适配窗口。
- 你可能看到以「[Conversation summary]」开头的消息——它们替代了更早的细节，可作为可靠上下文对待，无需质疑其来源。

## 命令安全审查

Shell 命令在执行前会经过安全审查。作为 MainAgent，你适用**严格规则**：

- 禁止：sudo、chmod 777、curl|sh / wget|sh、全局 pip/npm 安装、iptables 与防火墙改动、写入系统文件（/etc/hosts、.ssh/authorized_keys 等）、读取凭据（.env、.aws/credentials、.ssh/id_rsa 等）。
- 禁止破坏性操作：rm -rf /、mkfs、dd if=、fork bomb 等。
- 命令被拒时你会收到说明规则的错误——应调整方式（例如 pip 用 venv、npm 用用户作用域、避免 sudo）而非重试同一命令。

## 业务功能一览

### 对话模式

用户以**会话**为单位发起对话，每条会话绑定模型和工具配置。有两种使用方式：

1. **自由模式**：每次手动选择模型、MCP、技能。
2. **角色模式**：用户预先创建角色（Persona），为角色设定系统提示词、固定模型、MCP 和技能组合。选择角色即可快速开始同类任务。

### 对话中的附件

用户在聊天中上传的图片、文档等附件，系统在对话开始前自动处理并注入到消息中：

- **图片**（jpg/png/gif/webp 等）：原样存入 <根路径>/temp/ 目录，你可用文件读取工具查看。
- **文档**（PDF、Word、PPT、Excel、HTML 等）：系统自动将其内容解析为 Markdown 文本，生成 .md 文件存入 <根路径>/temp/ 目录，你可用文件读取工具获取文档的文字内容。
- **纯文本文件**（txt、log、json、xml、yaml 等）：原样复制到 <根路径>/temp/ 目录。

附件文件的路径通过 <raven-upload> 标签出现在消息内容中（该标签由系统生成，你不需要输出）。消息中的附件是用户任务的重要上下文，理解附件内容后再执行任务。附件文件位于 <根路径>/temp/ 目录。

### 技能系统

技能是用户安装的专项能力扩展，遵循 SKILL.md 规范。用户可从技能市场安装、多用户共享，也可自行创建。全局技能（raven-install-skill、raven-chart、raven-about）由管理员管理，对所有用户可见。当用户要求安装技能时，必须先调用 raven-install-skill。

### MCP 集成

管理员可配置 MCP (Model Context Protocol) 端点来连接外部服务。启用的 MCP 端点会自动注册为你的工具。管理员可将某些 MCP 设为「始终启用」，对所有用户生效。

### 文件管理

用户可通过前端文件浏览器管理自己的文件空间（目录：projects/、documents/、images/、videos/、downloads/、temp/、skills/）。上传采用分块 API，支持大文件。

### 会话分享

用户可将会话生成分享链接（公开或内部），其他人通过链接查看对话内容。技能也可分享给多用户。

## 前端界面

用户通过左侧边栏操作平台：普通用户可访问仪表盘、文件管理、技能中心、角色管理及历史会话；管理员额外拥有用户管理、系统设置、模型配置等页面。详细界面与操作说明见 raven-ui-guide 技能。

## 前端渲染与标签

你输出的 Markdown 会被前端渲染为富文本（标题、列表、表格、代码高亮、LaTeX）。回复通过 SSE 流式推送，用户实时看到推理与内容。事件类型包括 reasoning、content、tool、retry、end、context、heartbeat：context 事件携带当前 token 用量与模型上下文长度，前端据此渲染用量条；模型请求失败时系统按退避策略自动重试并发出 retry 事件。raven 标签（<raven-file>、<raven-chart>、<raven-upload>）的渲染规则详见系统提示词和对应技能。

## 系统设置要点

管理员可在系统设置中调节你的运行行为，关键项包括：

- **迭代与超时**：最大迭代步数、单次查询超时。
- **上下文管理**：压缩触发阈值与保留轮数、裁剪 token 阈值与工具结果截断长度。
- **请求策略**：LLM 请求间隔、失败重试次数、限流等待与退避基数。
- **能力开关**：Web 抓取、视觉理解、OCR 是否启用；Shell 单命令超时；文件分享链接有效期。

某项能力不可用（如 web_fetch 被禁用）通常是管理员配置所致，可提示用户联系管理员。

## 系统提示词未覆盖的重要约定

**多用户隔离**：不可访问其他用户目录；不可在回复中泄露任何其他用户的用户名、目录名、文件路径或对话内容。

**系统配置**：/raven/data/config.yaml 不可修改，需由管理员手动编辑。

**资源使用**：容器为共享资源，长时高 CPU/内存任务应确认用户意图后再执行。
`

const SystemSkillAboutEn = `---
name: raven-about
description: Raven platform overview. Covers product positioning, runtime environment, agent runtime mechanics, tool inventory, memory and safety policies, business features, and usage conventions to help the model understand the platform context.
---

# Raven Platform

## What Raven Is

Raven is a **self-hosted, multi-user AI Agent platform**. Users run it on their own servers and interact with AI through a web frontend for programming, data analysis, document processing, and more.

**Target users**: engineers, programmers, finance professionals who deploy AI services on their own infrastructure. They fully own their models, data, and toolchain — no dependency on external AI services.

**Design philosophy**: Raven is an engineer's tool — restrained and efficient. Pure grayscale palette, monospace typography, dark-only theme. Information hierarchy driven by typography.

**Deployment**: Docker container, user maps their own ports.

## Runtime Environment

You run inside a Docker container (ubuntu 22.04). Pre-installed toolchain, all available via shell:

- Go, Node 22 (pnpm), Python 3 (pip/venv), gcc/g++
- git, curl, wget, make, sqlite3, unzip, uv
- apt cache retained; apt-get install for extra system packages as needed
- If RAVEN_CHINA_MIRROR=1 is set, package managers are routed to domestic mirrors

Your workspace is <root>. Do not access other users' directories or files. The container is shared — long-running, CPU/memory-heavy tasks should be confirmed with the user first; batch when appropriate.

## Your Runtime Mechanics

You run in ReAct mode: reason → call a tool → observe the result → continue reasoning until the task is done.

- **Sub-agent delegation**: complex, well-scoped independent subtasks can be delegated to the **SubAgent**. It runs in a clean context with no conversation history — ideal for boundary-clear work that can proceed in parallel with the main thread.
- **Plan mode**: for complex multi-step tasks, use the plan tools (TaskCreate/TaskUpdate/TaskGet/TaskList) to break down and track steps. Plans live only for the duration of the run and are not persisted across sessions.

## Tool Inventory

You have the following built-in tools (MCP tools are added dynamically per admin config):

- **Filesystem**: ls, read_file, write_file, edit_file, glob, grep, mkdir, mv, rm, cp
- **Terminal**: execute (streaming output; per-command timeout set by admin)
- **Web fetch**: raven_web_fetch (HTTP fetch converted to Markdown; no JS rendering; admin-gated)
- **Multimodal**: raven_visual_understand (visual analysis of images/video/audio; admin-gated + requires a visual model)
- **Unique ID**: raven_generate_uniq_id (always available; filenames under documents/, images/, videos/ must use it)
- **Plan & skills**: TaskCreate/TaskUpdate/TaskGet/TaskList; on-demand loading of selected skills' SKILL.md instructions
- **Sub-agent**: sub_agent

## Memory & Context Management

- Conversation history is persisted per round and retained across rounds.
- **Context compression**: when token usage approaches the model's context-window limit, older rounds are summarized into a "conversation summary"; the most recent rounds are kept intact. Summaries preserve file paths, URLs, commands, and raven tags verbatim.
- **Tool-result pruning**: over-long tool outputs are truncated (head and tail retained); when total tokens exceed the pruning threshold, the oldest tool rounds are dropped to fit the window.
- You may see messages prefixed "[Conversation summary]" — they replace older detail and can be treated as reliable context, no need to question their origin.

## Command Safety Review

Shell commands are validated before execution. As the MainAgent you are subject to **strict rules**:

- Blocked: sudo, chmod 777, curl|sh / wget|sh, global pip/npm installs, iptables and firewall changes, writing system files (/etc/hosts, .ssh/authorized_keys, etc.), reading credentials (.env, .aws/credentials, .ssh/id_rsa, etc.).
- Blocked destructive ops: rm -rf /, mkfs, dd if=, fork bombs, etc.
- When a command is rejected you'll get an error explaining the rule — adjust your approach (e.g., pip in a venv, npm in user scope, avoid sudo) rather than retrying the same command.

## Business Features

### Chat Modes

Users converse in **sessions**, each bound to a model and tool config. Two modes:

1. **Free mode**: user picks model, MCPs, skills per session.
2. **Persona mode**: user pre-creates a Persona with a system prompt, fixed model, MCPs, and skills. Select the persona to instantly set up for recurring task types.

### Attachments in Chat

When users upload images, documents, or other files in a conversation, the system processes them before the agent runs and injects them into the message:

- **Images** (jpg/png/gif/webp, etc.): stored as-is under <root>/temp/. Read with file tools.
- **Documents** (PDF, Word, PPT, Excel, HTML, etc.): the system automatically extracts their text content into a Markdown .md file under <root>/temp/. Read with file tools to get the document's text content.
- **Plain text files** (txt, log, json, xml, yaml, etc.): copied as-is to <root>/temp/.

Attachment paths appear in the message via <raven-upload> tags (system-generated; you do not output these). Attachments provide important context — understand them before executing the task. Attached files are located under <root>/temp/.

### Skill System

Skills are user-installed capability extensions following SKILL.md conventions. Users install from the marketplace, share with their team, or create their own. Global skills (raven-install-skill, raven-chart, raven-about) are managed by admins and visible to all. When a user asks to install a skill, invoke raven-install-skill first.

### MCP Integration

Admins can configure MCP (Model Context Protocol) endpoints to connect external services. Enabled MCP endpoints auto-register as your tools. Admins can set certain MCPs as "always on" for all users.

### File Management

Users manage their file space through the frontend file browser (directories: projects/, documents/, images/, videos/, downloads/, temp/, skills/). Upload uses a chunked API supporting large files.

### Session Sharing

Users can generate share links for sessions (public or internal). Others view shared conversations through the link. Skills can also be shared with the team.

## Frontend UI

Users navigate through the left sidebar: regular users access Dashboard, Files, Skills, Personas, and session history; admins additionally have User Management, System Settings, Model Configuration, etc. For detailed UI and operation instructions, see the raven-ui-guide skill.

## Frontend Rendering & Tags

Your Markdown output is rendered as rich content (headings, lists, tables, code highlighting, LaTeX). Replies stream via SSE in real time. Event types include reasoning, content, tool, retry, end, context, heartbeat: context events carry current token usage and the model's context length so the frontend can render a usage bar, and on model failure the system retries automatically with backoff and emits retry events. Rendering rules for raven tags (<raven-file>, <raven-chart>, <raven-upload>) are covered in the system prompt and respective skills.

## System Settings

Admins can tune your runtime behavior in System Settings. Key knobs:

- **Iteration & timeout**: max iteration steps, per-query timeout.
- **Context management**: compression trigger threshold and rounds kept, pruning token threshold and tool-result truncation lengths.
- **Request policy**: LLM request delay, retry count, rate-limit wait, backoff base.
- **Capability toggles**: web fetch, visual understanding, OCR on/off; per-command shell timeout; file share-link expiry.

If a capability is unavailable (e.g., web_fetch disabled), it is usually an admin setting — suggest the user contact their admin.

## Important Conventions NOT in the System Prompt

**Multi-user isolation**: do not access other users' directories. Never leak other users' usernames, directory names, file paths, or conversation content in replies.

**System config**: /raven/data/config.yaml cannot be modified. Changes must be made manually by an admin.

**Resource usage**: the container is shared. Long-running, CPU/memory-heavy tasks should be confirmed with the user before execution.
`

const SystemSkillUIGuide = `---
name: raven-ui-guide
description: Raven 前端界面与操作指南。详细介绍前端各页面的布局、功能与使用方式，帮助模型理解用户的操作环境，从而更好地引导用户。
---

# Raven 前端界面与操作

## 整体布局

Raven 采用左侧边栏 + 右侧主内容区布局。侧边栏可折叠为仅图标模式。

### 侧边栏（桌面端）

- **顶部**：Raven logo + 折叠/展开按钮
- **新建对话按钮**：点击进入空白对话页，开始新的 AI 交互
- **工作区导航**（折叠分组）：
  - 仪表盘：查看个人用量统计
  - 文件：管理个人工作空间文件
  - 技能：管理已安装的技能，浏览技能市场
  - 角色：创建和管理 AI 角色（预配置模型、MCP、技能的系统提示）
- **会话历史**：按时间分组（今天 / 近7天 / 近30天），支持无限滚动加载，每条会话可重命名或归档删除
- **底部用户区**：显示头像和用户名，点击弹出菜单：
  - 个人资料
  - 管理模式（仅管理员可见，切换到后台管理）
  - 退出登录

### 移动端

侧边栏默认隐藏，通过顶部工具栏的菜单按钮打开，以遮罩形式出现。

## 对话页（核心交互界面）

对话页是用户与 AI 交互的主要界面，分两种状态：

### 新建对话（/chat，无会话）

**顶部工具栏**：
- 模型/角色选择下拉框，分为「模型」和「角色」两个区域。选择一个角色会自动配置模型、MCP 工具和技能；选择一个模型会清除角色选择。
- 如果系统尚未配置任何模型，显示「无模型」警告按钮，点击后弹窗说明配置步骤。
- 移动端显示菜单按钮。

**初始状态**：中心显示 Raven logo 和「What can I help you build?」问候语。

**输入区域**（完整的 NewChatInput）为一个圆角容器，包含：
- **文本输入框**：支持多行，自动增高（上限约4行），Enter 发送（Shift+Enter 换行）。中文输入法下 Enter 不会误发送。
- **配置按钮（+ 扳手图标）**：打开浮动面板，可选择本次对话使用的 MCP 工具和技能。选中的工具/技能会显示高亮。当选择了角色时，此按钮禁用（角色已预配置工具和技能）。
- **文件上传按钮（回形针图标）**：点击选择本地文件，最多 10 个。上传使用分块方式支持大文件。图片文件会显示缩略图预览。上传完毕或失败有状态提示。可点击移除已上传文件。
- **项目目录按钮（文件夹图标）**：选择 projects/ 下的子目录作为本次对话的工作空间。选择后高亮显示。有「清空选择」选项。
- **Thinking 开关**：切换模型的深度思考模式。发送前可随意切换，发送后不可改。
- **发送/停止按钮**：向上箭头样式，高亮色。输入为空时灰色不可用。发送后变为停止按钮（圆形停止图标）。

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

**消息区域**：按时间顺序展示用户/助手/工具/摘要消息。助手消息实时流式显示（SSE 推送），包括思考内容（reasoning）和正式回复（content）。工具调用过程也会以卡片形式呈现。

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

**右键菜单**：预览、下载、重命名、删除。

**预览对话框**：文本文件滚动查看，图片/视频/音频嵌入播放，PDF 用 iframe 展示。

**拖拽上传**：支持拖拽文件到页面区域上传。

用户工作空间下有固定目录：documents/、projects/、images/、videos/、downloads/、temp/、skills/。

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

## 技能管理页（/skills）

管理用户技能，分三个标签页：

| 标签 | 内容 |
|------|------|
| 已安装（有数量） | 用户的技能列表。每项显示图标、名称、描述、Always-On 开关、安装时间。点击行打开详情侧边栏（可编辑图标和分类、查看 SKILL.md 内容、切换 Always-On、分享自定义技能、删除） |
| 市场（有数量） | 技能市场列表，每项显示安装按钮和安装次数。已安装的显示「已安装」标记。**注意**：点击安装仅复制文件并注册，依赖安装（如 pip/npm install）需在对话中完成 —— 弹窗会提供提示语，用户复制到新对话发给 AI 完成，此过程较耗时。 |
| 多用户 | 其他用户分享的技能，可安装或（自己的分享）取消分享 |

顶部有搜索框和分类筛选下拉框。右上角有同步（刷新）按钮。

## 角色管理页（/personas）

创建和管理 AI 角色（预配置的系统提示 + 模型 + MCP + 技能组合）。

**角色列表（/personas）**：顶部有「新建角色」按钮。每个角色卡片显示图标、名称、分类、角色提示摘要、MCP/技能标签、模型名。点击进入详情，右侧有编辑/删除菜单。

**角色详情/编辑**：可编辑名称、图标、分类、角色设定（textarea，最長 500 字符）、模型配置、MCP 工具（可搜索复选框列表）、技能（可搜索复选框列表）。新建时可以从模板库选择模板。

保存/取消按钮固定在顶部。

## 个人资料页（/profile）

个人设置，分为：
1. **个人信息**：头像（可点击上传）、昵称（可编辑）、用户名（只读）、角色标签、邮箱（可编辑）、注册时间。
2. **外观**：主题模式选择（浅色/深色/跟随系统）。
3. **账号安全**：修改密码（需验证当前密码，新密码至少 8 位且包含字母和数字）。
4. **关于**：项目介绍 + 退出登录。

## 管理模式

管理员通过侧边栏用户菜单的「管理模式」切换到后台管理。侧边栏导航变为管理员菜单：

### 管理仪表盘（/admin）
系统全局统计：活跃用户数、总会话数、周新增、周/日 Token、已启用模型数。包含各类趋势图表和排行榜。

### 用户管理（/admin/users）
用户列表（搜索、角色筛选、分页）。可添加/编辑/删除用户，重置密码，切换启用/禁用状态。编辑和添加通过侧边滑出面板操作。

### 系统设置（/admin/settings）
系统级配置，按分组展示（Agent、ClawHub、通用、分享、知识库、工具等）。每项设置根据类型提供不同的输入控件（文本、数字、开关、JSON 编辑器等）。修改后顶部保存按钮显示修改数量。

### 模型管理（/admin/models）
AI 模型配置列表。可添加（选择提供商、配置 API Key、选择模型名称、设置上下文长度等）、编辑、复制、删除，设置默认/压缩/多模态标记。添加时支持「保存并测试」按钮验证连接。

### MCP 管理（/admin/mcp）
MCP（Model Context Protocol）端点管理。支持三种传输类型：Stdio（npx/uvx 启动程序 + 参数和环墋变量）、SSE/HTTP（服务 URL + 请求头）。可设 Always-On 对所有用户生效。有推荐模板库可快速安装常用 MCP。支持分页。

### 技能管理（/admin/skills）
全局技能和市场技能的管理。全局技能：启用/禁用系统级技能。市场技能：发布自定义技能、从 ClawHub 导入技能、管理类别（添加/编辑/删除分类）。

### 角色模板管理（/admin/persona-templates）
管理角色模板，用户新建角色时可选择这些模板。分为模板和分类两个标签页。

### 系统信息（/admin/systemInfo）
只读系统概览：用户/模型/MCP/技能/会话/文件/分享的统计数据，数据库连接池状态、MCP 健康状态、磁盘使用情况、启用的插件列表。

## 其他页面

### 登录页（/login）
用户名、密码输入，可能触发验证码（算术题）。URL 带 ?expired=1 时显示「会话已过期」提示。

### 安装向导（/install）
系统首次启动时的初始化流程，分 5 步：选择语言 → 配置域名 → 创建管理员账号 → 配置数据库 → 配置缓存。仅系统未初始化时可访问。

## 常见问题导航指南

用户遇到问题时，指导他们访问对应页面：

- 「看不到模型」→ 找管理员在 /admin/models 添加模型，或自己在 /chat 顶部选择模型
- 「模型报错/不可用」→ 管理员在 /admin/models 检查 API Key 和连接
- 「想用某个外部工具」→ 管理员在 /admin/mcp 添加 MCP 端点
- 「想换角色/人格」→ 在 /personas 创建或选择角色，对话时在顶部下拉选择
- 「文件在哪」→ /files 页面管理所有文件
- 「对话太长了」→ 点击顶部工具栏的压缩按钮（收缩图标）
- 「想把对话发给别人看」→ 点击顶部工具栏的分享按钮生成链接
- 「Token 用多少了」→ 对话页顶部或 /dashboard 查看
- 「回复太慢了」→ 关闭 Thinking 开关可加速，或联系管理员检查系统设置中的超时配置
- 「想安装新技能」→ /skills 市场标签页。点击安装后弹窗提供依赖安装提示语，复制到新对话让 AI 完成（过程较慢，涉及 pip/npm/go 安装）
- 「修改密码」→ /profile 页面的账号安全区域
- 「忘记密码」→ 联系管理员重置（管理员在 /admin/users 操作）
- 「想设置深色模式」→ /profile 页面的外观设置
- 「API 密钥之类的放哪」→ /files 页面，点击环境变量按钮（{} 图标），设置的变量自动注入到 AI 执行的每个命令中。也可直接让 AI 写入 .profile。
`

const SystemSkillUIGuideEn = `---
name: raven-ui-guide
description: Raven frontend UI and operation guide. Details the layout, features, and usage of each frontend page, helping the model understand the user's operating environment.
---

# Raven Frontend UI & Operations

## Overall Layout

Raven uses a left sidebar + right main content layout. The sidebar is collapsible to icon-only mode.

### Sidebar (Desktop)

- **Top**: Raven logo + collapse/expand toggle button
- **New Chat button**: Navigate to a blank chat page to start a new AI interaction
- **Workspace navigation** (collapsible groups):
  - Dashboard: View personal usage statistics
  - Files: Manage personal workspace files
  - Skills: Manage installed skills, browse the skill marketplace
  - Personas: Create and manage AI personas (pre-configured system prompts with models, MCP tools, and skills)
- **Session history**: Grouped by time (Today / Past 7 days / Past 30 days), infinite scroll loading. Each session can be renamed or archived (deleted).
- **Bottom user area**: Shows avatar and username. Click opens a popover menu with:
  - Profile
  - Admin Mode (admin-only, switches to backend management)
  - Logout

### Mobile

The sidebar is hidden by default, accessible via a hamburger menu button in the top toolbar, appearing as an overlay.

## Chat Page (Core Interaction)

The chat page is the main AI interaction interface, with two states:

### New Chat (/chat, no session)

**Top toolbar**:
- Model/Persona selector dropdown with two sections: Models and Personas. Selecting a persona auto-configures the model, MCP tools, and skills. Selecting a model clears the persona selection.
- If no models are configured, a "No model" warning button appears with a dialog explaining setup steps.
- Menu button on mobile.

**Initial state**: Displays the Raven logo centered with "What can I help you build?".

**Input area** (full NewChatInput) is a rounded container with:
- **Text input**: Supports multi-line, auto-grows (max ~4 lines). Enter to send (Shift+Enter for newline). IME input (e.g. Chinese) is handled correctly — Enter while composing won't accidentally send.
- **Config button (+ wrench icon)**: Opens a floating panel to select MCP tools and skills for this conversation. Selected items show highlighted. Disabled when a persona is selected (personas pre-configure tools/skills).
- **File upload button (paperclip icon)**: Select local files, up to 10. Uses chunked upload for large files. Image files show preview thumbnails. Status indicators for upload progress/success/failure. Click to remove uploaded files.
- **Project directory button (folder icon)**: Select a subdirectory under projects/ as the workspace for this chat. Highlighted when selected. Has a "Clear selection" option.
- **Thinking toggle**: Enables/disables the model's deep thinking mode. Can be toggled freely before sending, locked after sending.
- **Send/Stop button**: Up-arrow icon, highlight color. Greyed out when input is empty. After sending, becomes a stop button (circle stop icon).

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

**Message area**: Displays messages in chronological order (user/assistant/tool/summary roles). Assistant messages stream in real time via SSE, including reasoning content and the actual reply. Tool invocations are shown as cards.

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

**Right-click menu**: Preview, Download, Rename, Delete.

**Preview dialog**: Text files in scrollable view, images/videos/audio as embedded media, PDFs in an iframe.

**Drag-and-drop**: Files can be dragged onto the page area to upload.

Fixed user workspace directories: documents/, projects/, images/, videos/, downloads/, temp/, skills/.

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

## Skills Page (/skills)

Manage user skills across three tabs:

| Tab | Content |
|-----|---------|
| Installed (with count) | User's skill list. Each item shows icon, name, description, Always-On toggle, install date. Click a row to open detail drawer (edit icon/category, view SKILL.md content, toggle Always-On, share custom skills, delete) |
| Marketplace (with count) | Available skills from the marketplace, each with install button and install count. Already installed skills show "Installed" badge. **Note**: clicking Install only copies files and registers the skill. Dependency installation (pip/npm install, etc.) must be done in a conversation — a dialog provides a prompt to copy into a new chat for the AI to complete. This can be time-consuming. |
| Team | Skills shared by other users. Install or cancel share (for own shares) |

Top bar has a search box and category filter dropdown. Sync (refresh) button in the top-right corner.

## Personas Page (/personas)

Create and manage AI personas (pre-configured system prompts + model + MCP + skills).

**Persona list (/personas)**: "New Persona" button at top. Each card shows icon, name, category, role info snippet, MCP/skill tags, model name. Click to enter detail view. Right-side menu for edit/delete.

**Persona detail/edit**: Editable fields include name, icon, category, persona settings (textarea, max 500 chars), model config, MCP tools (searchable checkbox list), skills (searchable checkbox list). When creating, can choose from template library.

Save/Cancel buttons fixed at top.

## Profile Page (/profile)

Personal settings:
1. **Personal Info**: Avatar (uploadable), nickname (editable), username (read-only), role badge, email (editable), registration date.
2. **Appearance**: Theme mode selector (Light / Dark / System).
3. **Account Security**: Change password (requires current password, new password must be at least 8 chars with letters and numbers).
4. **About**: Project description + Logout.

## Admin Mode

Administrators switch to backend management via the "Admin Mode" option in the sidebar user menu. The sidebar navigation changes to admin menus:

### Admin Dashboard (/admin)
System-wide statistics: active users, total sessions, new this week, weekly/daily tokens, enabled models. Includes trend charts and rankings.

### User Management (/admin/users)
User list (search, role filter, pagination). Add/edit/delete users, reset password, toggle enable/disable. Operations via slide-in drawer panels.

### System Settings (/admin/settings)
System configuration key-value store, grouped by domain (Agent, ClawHub, General, Sharing, Knowledge, Tools). Each setting has appropriate input controls (text, number, switch, JSON editor, etc.). Top save button shows count of modified settings.

### Model Management (/admin/models)
AI model configuration list. Add (select provider, configure API key, choose model name, set context length, etc.), edit, duplicate, delete. Set as default/compress/multimodal. "Save & Test" button to verify connectivity.

### MCP Management (/admin/mcp)
MCP (Model Context Protocol) endpoint management. Three transport types: Stdio (npx/uvx launcher with args and env vars), SSE/HTTP (service URL + request headers). Can set Always-On for all users. Recommended template library for quick MCP installation. Pagination support.

### Skill Admin (/admin/skills)
Global skills and marketplace skills management. Global skills: enable/disable system-level skills. Marketplace: publish custom skills, import from ClawHub, manage categories (add/edit/delete).

### Persona Template Admin (/admin/persona-templates)
Manage persona templates that users can choose from when creating new personas. Two tabs: Templates and Categories.

### System Info (/admin/systemInfo)
Read-only system overview: counts (users/models/MCP/skills/sessions/files/shares), database connection pool status, MCP health status, disk usage, active plugin list.

## Other Pages

### Login (/login)
Username and password input. May trigger a captcha (simple arithmetic). URL with ?expired=1 shows "Session expired" message.

### Install Wizard (/install)
Initial system setup wizard (only accessible when system is not initialized). 5 steps: language → domain → admin account → database → cache.

## Troubleshooting Guide

When users encounter issues, guide them to the appropriate page:

- "No models available" → Admin should add models at /admin/models, or the user selects a model in the /chat top dropdown
- "Model errors/unavailable" → Admin checks API key and connection at /admin/models
- "Want to use an external tool" → Admin adds MCP endpoint at /admin/mcp
- "Want to switch persona" → Create or select a persona at /personas, then choose it from the top dropdown in chat
- "Where are my files" → /files page manages all workspace files
- "Conversation is too long" → Click the compress button (shrink icon) in the top toolbar
- "Want to share a conversation" → Click the share button in the top toolbar to generate a link
- "How many tokens used" → Check the top toolbar in chat or visit /dashboard
- "Response is too slow" → Turn off the Thinking toggle to speed up, or ask admin to check timeout settings in system settings
- "Want to install new skills" → /skills page, Marketplace tab. After clicking Install, copy the dependency prompt from the dialog into a new chat for the AI to complete installation (may be slow, involving pip/npm/go)
- "Change password" → /profile page, Account Security section
- "Forgot password" → Contact admin to reset (admin operates at /admin/users)
- "Want dark mode" → /profile page, Appearance settings
- "Where to put API keys" → /files page, click the Environment Variables button ({} icon). Variables set here are auto-injected into every command the AI executes. You can also ask the AI to write to .profile directly.
`
