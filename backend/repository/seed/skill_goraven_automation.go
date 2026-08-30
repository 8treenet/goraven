package seed

const SystemSkillGoRavenAutomation = `---
name: goraven-automation
description: 当用户提出定时任务、自动化或任何周期性、延迟执行需求时调用，如"每天9点自动…"、"每隔30分钟…"、"下周一提醒我"、"两小时后执行"。
---

# 自动化任务创建

用户希望**在未来某个时间点、或按某种节奏，无需他在场的情况下自动执行一项工作**。通过五个内置工具完成：先查当前时间锚定计划，再创建任务；已有任务可查询、查看详情与修改。到点后调度器会自动新建会话执行需求。

| 工具 | 用途 |
|------|------|
| ` + "`goraven_get_current_time`" + ` | 获取当前日期时间和星期。计算任何绝对时间前必须调用，禁止凭空猜测 |
| ` + "`goraven_create_automation_task`" + ` | 创建自动化任务 |
| ` + "`goraven_list_automation_tasks`" + ` | 查询当前用户启用中的任务（分页），用于查找 task_id |
| ` + "`goraven_get_automation_task`" + ` | 按 task_id 返回完整详情（含需求与状态），修改前用它获取当前字段 |
| ` + "`goraven_update_automation_task`" + ` | 全量替换更新任务；改计划会重置下次执行时间；已完成任务不可改 |

## 触发场景

包括但不限于：
- "每天/每周/工作日 X 点帮我做 …"
- "每隔 N 分钟/小时检查一次 …"
- "明天下午 3 点 / 下周一 / 两小时后 …"
- "以后每次收盘后…""定期备份…"
- 用户明说"定时、自动、周期性、轮询、到点提醒"等字眼

## 咨询类提问

用户问"如何创建""能定时执行吗"是在咨询，不是下达任务，禁止照搬本文的表格、工具名和参数。用大白话介绍能力并举一两个例子（如"每天 9 点汇总待办""每 30 分钟检查一次"），再用提问引导说出具体需求：什么时间或节奏执行？到点做什么？结果怎么交付？需求明确后转入「创建流程」。机制类疑问（任务在哪看、错过是否补跑、执行时用什么模型和配置）按「执行机制」作答，同样用大白话。

## 创建流程

### 第一步：明确需求

未来会话执行的指令必须完整。至少澄清四件事：

1. **做什么** —— 完整的执行指令（见 Requirement 写法）
2. **何时执行** —— 绝对时间还是相对频率？一次还是重复？
3. **结果交付** —— 文字回复？图表标签？写入哪个文件？
4. **异常处理** —— 数据缺失、文件不存在时跳过还是说明原因？

执行计划存在歧义时**必须先确认再创建**：
- 相对表述换算："每小时"=每 60 分钟；"工作日"通常指周一至周五，需与用户核对
- 星期编号复述：0=周日，1~6=周一至周六；创建前向用户复述换算出的具体日期
- 单次 vs 周期："下个月提醒我"是一次性还是每月一次？问了再建
- 间隔粒度最小为 5 分钟，不支持秒级任务

### 第二步：锚定当前时间

创建任何任务前先调用 ` + "`goraven_get_current_time`" + `：

- 所有相对时间都基于它换算为绝对值："明天9点"、"两小时后"、"下周一"
- 响应字段可直接使用：` + "`datetime`" + `（RFC3339）、` + "`weekday`" + `（星期）
- 可选传参 ` + "`timezone`" + `（IANA 名称）；不传即服务器本地时区，调度同样按本地时区执行

### 第三步：选择执行类型

| exec_type | 含义 | 必填参数 |
|---|---|---|
| 1 | 单次固定时间 | ` + "`run_at`" + `（未来时间） |
| 2 | 按间隔循环 | ` + "`interval_minutes`" + ` |
| 3 | 每天固定时刻 | ` + "`fixed_time`" + ` |
| 4 | 每周固定时刻 | ` + "`weekday`" + ` + ` + "`fixed_time`" + ` |

参数格式与取值范围以工具 schema 为准；参数组合必须与类型匹配：每天不要传 weekday，每周必须带 weekday。

### 第四步：撰写 Requirement（最关键）

触发时会**新建一个空白会话**执行，它看不到本次对话的历史、附件和你说过的话。因此：

- **禁止指代**："上面提到的文件""我们讨论的方案"这类表述一律不允许出现
- **信息自包含**：目标、步骤、涉及的路径（<根路径> 开头的绝对路径）、数据来源全部写进 requirement
- **写明交付方式**：回复中嵌入 <goraven-chart> 图表标签？纯文字总结？把结果写入哪个文件？
- **写明容错**：如"文件不存在时仅回复提示，不做其他操作"

示例 —— 用户说："每天早上9点汇总一下我文档目录里 todos.csv 的待办事项，画个图"

- exec_type = 3，fixed_time = "09:00"
- requirement 内容：

` + "```" + `
读取 <根路径>/documents/todos.csv。
1. 统计各状态待办数量，使用 <goraven-chart> 柱状图嵌入回复
2. 列出今日到期的事项原文
3. 若文件不存在，仅回复"todos.csv 不存在"，不执行其他操作
` + "```" + `

### 第五步：创建并复述结果

创建成功后响应含 ` + "`task_id`" + `、` + "`title`" + `、` + "`next_run_at`" + `（首次执行时间）。把 next_run_at 明确转述给用户确认，并告知可随时到控制台的「自动化任务」中查看该任务的执行状态与历史结果。

**会话上下文继承**：任务创建时自动快照当前会话的模型、角色、项目目录、MCP 和技能配置，触发会话沿用同一套配置。不要询问用户"用哪个模型/MCP"，也不要在 requirement 里交代这些信息；创建成功后把沿用的配置（模型/角色/项目目录/MCP/技能）简要转述给用户。

## 修改已有任务

用户要调整已有任务（改需求、改执行计划等）时：

1. 用 ` + "`goraven_list_automation_tasks`" + ` 找到 task_id（仅返回启用中的任务；停用任务需从用户或控制台获得 ID）
2. 用 ` + "`goraven_get_automation_task`" + ` 获取完整当前字段
3. 与用户确认要改的部分，再调用 ` + "`goraven_update_automation_task`" + ` **全量回显**所有业务字段（标题、需求、exec_type 及对应计划参数），仅应用修改
4. 修改任一计划字段会重置下次执行时间；仅改标题或需求则保持原执行节奏。复述响应中的 next_run_at
5. 已完成任务不可修改；停用任务可以修改

修改 requirement 同样遵守"自包含"原则：触发会话看不到修改时的这段对话。

## 执行机制（回答用户疑问时使用）

- 触发产生的会话不在侧边栏显示；在控制台「自动化任务」的任务详情中可打开对应会话查看过程与结果
- 执行成功才记录一条执行记录（保留最近若干条）；失败的信息去触发产生的会话里看
- 任务执行沿用**创建时**的会话配置（模型、角色、项目目录、MCP、技能），之后当前会话更换配置不影响已建任务；任务配置创建后不可修改，需要更换时删除重建
- 间隔型从**上次真正跑完的时刻**起算间隔（跑 5 分钟+间隔 30 分钟 → 约 35 分钟后下次），不会堆积
- 每天/每周锚定钟表时间；服务重启期间错过的周期不补跑；间隔型最多补跑一次；单次任务会在服务恢复后尽快补执行

## 规则

- 时间数值一律来自 goraven_get_current_time 的推算结果，不接受估算的时间
- 必须有用户的明确意图才能创建任务，禁止为了演示擅自创建
- 单次任务失败不会重试，requirement 中考虑好兜底方式
- 咨询类提问用大白话引导，禁止向用户输出工具名、参数名或内部表格
`

const SystemSkillGoRavenAutomationEn = `---
name: goraven-automation
description: Use when users request scheduled tasks, automation, or recurring/delayed execution, such as "every day at 9am automatically...", "check every 30 minutes", "remind me next Monday", "run in two hours".
---

# Automation Task Creation

The user wants work performed **at a future time or on a recurring rhythm, without being present**. Five built-in tools accomplish this: anchor the schedule with a time query first, then create the task; existing tasks can be listed, inspected and updated. When due, the scheduler spawns a fresh session to run the requirement.

| Tool | Purpose |
|------|---------|
| ` + "`goraven_get_current_time`" + ` | Get the current date, time and weekday. Must be called before computing any absolute time; never guess |
| ` + "`goraven_create_automation_task`" + ` | Create the automation task |
| ` + "`goraven_list_automation_tasks`" + ` | List the user's enabled tasks (paginated); use to find task IDs |
| ` + "`goraven_get_automation_task`" + ` | Full detail by task_id (requirement and status); fetch current fields before updating |
| ` + "`goraven_update_automation_task`" + ` | Update a task by full replacement; schedule changes reset next run; done tasks cannot be updated |

## Trigger Scenarios

Includes but not limited to:
- "Every day / every week / on weekdays at X do ..."
- "Check every N minutes/hours ..."
- "Tomorrow at 3pm / next Monday / in two hours ..."
- "After every market close ...", "periodically back up ..."
- Explicit words such as "scheduled, automatic, recurring, polling, reminder"

## Consultation Questions

Asking "how do I create one" or "can you run things on a schedule" is a consultation, not a task request — never recite this document's tables, tool names or parameters. Explain the capability in plain words with one or two examples ("summarize todos every morning at 9", "check every 30 minutes"), then elicit the need with questions: when or how often? what to do when triggered? how to deliver results? Once the need is concrete, switch to the Creation Flow. Mechanics questions (where tasks live, missed runs, which model and configuration a run uses) are answered from Execution Mechanics, also in plain words.

## Creation Flow

### Step 1: Clarify Requirements

The instruction executed by the future session must be complete. Clarify at least four things:

1. **What to do** — the full execution instruction (see Writing the Requirement)
2. **When to run** — absolute time or relative frequency? Once or repeating?
3. **Delivery format** — text reply? chart tag? write results into which file?
4. **Error handling** — skip silently or report when data is missing or a file does not exist?

**Confirm before creating whenever the schedule is ambiguous:**
- Relative phrasing conversion: "hourly" = every 60 minutes; "weekdays" usually means Monday through Friday, verify with the user
- Weekday mapping restatement: 0 = Sunday, 1~6 = Monday through Saturday; restate the computed concrete date before creating
- Once vs recurring: "remind me next month" — one time or every month? Ask before creating
- The finest granularity is 5 minutes; second-level schedules are not supported

### Step 2: Anchor Current Time

Call ` + "`goraven_get_current_time`" + ` before creating any task:

- Convert all relative times to absolute values with it: "tomorrow 9am", "in two hours", "next Monday"
- Use response fields directly: ` + "`datetime`" + ` (RFC3339), ` + "`weekday`" + `
- Optional param ` + "`timezone`" + ` (IANA name); when omitted the server local timezone is used, and scheduling also runs on local time

### Step 3: Pick Execution Type

| exec_type | Meaning | Required params |
|---|---|---|
| 1 | Once at a fixed time | ` + "`run_at`" + ` (in the future) |
| 2 | Recurring interval | ` + "`interval_minutes`" + ` |
| 3 | Daily fixed time | ` + "`fixed_time`" + ` |
| 4 | Weekly fixed time | ` + "`weekday`" + ` plus ` + "`fixed_time`" + ` |

Formats and value ranges follow the tool schemas; parameter combinations must match the type: daily must not carry weekday, weekly must carry it.

### Step 4: Write the Requirement (most critical)

A **brand-new blank session** executes the task when triggered. It cannot see this conversation's history, attachments, or anything you said here. Therefore:

- **No references**: phrases like "the file mentioned above" or "our discussed plan" are forbidden
- **Self-contained info**: goal, steps, involved paths (absolute paths starting with <root path>), data sources all go into requirement
- **State delivery explicitly**: embed <goraven-chart> tags in the reply? plain text summary? write results into which file?
- **State fallback behavior**: e.g. "if the file does not exist, reply with a notice only and take no further action"

Example — user says: "Summarize todos.csv from my documents directory every morning at 9 and draw a chart"

- exec_type = 3, fixed_time = "09:00"
- requirement content:

` + "```" + `
Read <root path>/documents/todos.csv.
1. Count items by status and embed a bar chart via <goraven-chart> in the reply
2. List the original text of items due today
3. If the file does not exist, reply "todos.csv not found" only and take no other action
` + "```" + `

### Step 5: Create and Restate the Result

A successful creation returns ` + "`task_id`" + `, ` + "`title`" + `, ` + "`next_run_at`" + ` (first execution time). Relay next_run_at clearly back to the user for confirmation, and tell them the task's execution status and past results can be checked anytime in the console's Automation Tasks section.

**Session context inheritance**: on creation the task snapshots the current session's model, persona, project directory, MCP servers and skills; the triggered session reuses exactly that configuration. Do not ask the user which model/MCP to use, and do not mention these inside requirement. After creation, briefly tell the user which configuration (model/persona/project directory/MCP/skills) the task will run with.

## Updating an Existing Task

When the user wants to adjust an existing task (change the requirement, the schedule, etc.):

1. Find the task_id via ` + "`goraven_list_automation_tasks`" + ` (enabled tasks only; for disabled ones get the ID from the user or the console)
2. Fetch the full current fields via ` + "`goraven_get_automation_task`" + `
3. Confirm the intended changes with the user, then call ` + "`goraven_update_automation_task`" + ` echoing ALL business fields (title, requirement, exec_type and its schedule params) with only your modifications applied
4. Any schedule change resets the next run time; title/requirement-only changes keep the original rhythm. Relay the next_run_at from the response
5. Done tasks cannot be updated; disabled ones can

A modified requirement follows the same self-contained rule: the triggered session cannot see this conversation.

## Execution Mechanics (use when answering user questions)

- Sessions spawned by triggers are hidden from the sidebar; open them from the task detail in the console's Automation Tasks section to review process and results
- An execution record is written only on success (recent entries are kept); failure details live in the triggered session
- Runs reuse the session configuration captured at creation (model, persona, project directory, MCP, skills); later changes to the current session don't affect existing tasks. A task's configuration cannot be edited after creation — delete and recreate to change it
- Interval tasks count from the moment the previous run actually finished (5-minute run + 30-minute interval -> next run after ~35 minutes), never piling up
- Daily/weekly types anchor to clock time; periods missed during downtime are skipped, interval tasks catch up at most once, once-tasks execute as soon as possible after recovery

## Rules

- All time values must derive from goraven_get_current_time results; estimated times are unacceptable
- A clear user intent is required before creating any task; never create one just for demonstration
- Once-type tasks never retry on failure; consider fallback behavior inside requirement
- Consultation questions are answered in plain words; never expose tool names, parameter names or internal tables
`
