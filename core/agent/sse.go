package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"goraven/config"
	"goraven/core/tools"
)

const (
	SSEEventTypeReasoning = "reasoning"
	SSEEventTypeContent   = "content"
	SSEEventTypeTool      = "tool"
	SSEEventTypeEnd       = "end"
	SSEEventTypeRetry     = "retry"
	SSEEventTypeContext   = "context"
	SSEEventTypeHeartbeat = "heartbeat"
)

// SSEEventTool 工具事件的展示信息
type SSEEventTool struct {
	Name        string `json:"name"`        // 工具名称
	DisplayName string `json:"displayName"` // 本地化展示名称
	Icon        string `json:"icon"`        // emoji 图标
	Action      string `json:"action"`      // 本地化动作描述
}

// SSEContextInfo 上下文更新事件的负载
type SSEContextInfo struct {
	Tokens int `json:"tokens"` // 当前 prompt tokens
	Limit  int `json:"limit"`  // 模型最大上下文长度
}

// SSERetryInfo 模型重试事件的元数据
type SSERetryInfo struct {
	MaxRetries int    `json:"maxRetries"`      // 最大重试次数
	Attempt    int    `json:"attempt"`         // 当前重试次数（从1开始）
	Error      string `json:"error,omitempty"` // 触发的错误信息
}

// SSEEvent SSE 流推送事件，前端根据 Type 读取对应的负载字段
type SSEEvent struct {
	Type    string          `json:"type"`
	Content string          `json:"content,omitempty"`
	Tool    *SSEEventTool   `json:"tool,omitempty"`
	Retry   *SSERetryInfo   `json:"retry,omitempty"`
	Context *SSEContextInfo `json:"context,omitempty"`
}

type MCPToolDisplayGetter interface {
	GetMCPToolDisplay(toolName string) (tools.MCP, bool)
}

type ToolDisplay struct {
	NameZh   string
	NameEn   string
	ActionZh string
	ActionEn string
	Icon     string
}

// toolDisplayResolver 在固定 toolRegistry 之外提供按工具参数解析展示的能力。
// arguments 为 LLM 工具调用的 JSON 参数；main 为当前 MainAgent，供 resolver
// 访问 plan task backend 等运行期状态（如 TaskUpdate 通过 taskId 反查 subject）。
// 返回 (display, true) 命中后立即采用；返回 false 则回退到 toolRegistry 静态项。
type toolDisplayResolver func(arguments string, main *MainAgent) (ToolDisplay, bool)

// toolDisplayResolvers 工具参数展示解析器注册表，运行期不变，由 init 注入。
var toolDisplayResolvers map[string]toolDisplayResolver

// deferredToolEvents 标记需要在流式参数完整后再发送 SSE 工具事件的工具名。
// LLM 在流式输出 tool call 时参数会分片到达，提前解析会得到残缺/错误展示，
// 因此对依赖参数渲染展示的工具（execute、TaskCreate、TaskUpdate 等）需要延迟触发。
var deferredToolEvents map[string]bool

func init() {
	toolDisplayResolvers = map[string]toolDisplayResolver{}
	deferredToolEvents = map[string]bool{}

	// execute 按命令首 token 细分展示（curl/git/npm/python/go 等），
	// 未命中命令表时回退到 toolRegistry["execute"]（终端）默认。
	registerToolDisplayResolver("execute", func(arguments string, _ *MainAgent) (ToolDisplay, bool) {
		if display, ok := resolveShellCommandDisplay(extractArgumentValue(arguments, "command")); ok {
			return display, true
		}
		return ToolDisplay{}, false
	})

	// TaskUpdate 把 taskId / status / activeForm 带入动作描述，让用户看到
	// LLM 正在更新哪一项任务、切换到什么状态、当前进行时的具体动作。
	registerToolDisplayResolver("TaskUpdate", resolveTaskUpdateDisplay)

	// skill 按技能名称展示 "正在加载 {name}"，未命中时回退到 toolRegistry 静态项。
	registerToolDisplayResolver("skill", func(arguments string, _ *MainAgent) (ToolDisplay, bool) {
		base, ok := toolRegistry["skill"]
		if !ok {
			return ToolDisplay{}, false
		}
		skillName := extractArgumentValue(arguments, "skill")
		if skillName == "" {
			return base, true
		}
		return ToolDisplay{
			NameZh:   base.NameZh,
			NameEn:   base.NameEn,
			ActionZh: "正在加载 " + skillName,
			ActionEn: "Loading " + skillName,
			Icon:     base.Icon,
		}, true
	})

	// 依赖参数渲染的工具需延迟到流式参数完整后再发送 SSE 事件。
	registerDeferredToolEvent("execute")
	registerDeferredToolEvent("TaskUpdate")
	registerDeferredToolEvent("skill")
}

func registerToolDisplayResolver(name string, resolver toolDisplayResolver) {
	toolDisplayResolvers[name] = resolver
}

func registerDeferredToolEvent(name string) {
	deferredToolEvents[name] = true
}

func isDeferredToolEvent(name string) bool {
	return deferredToolEvents[name]
}

// resolveTaskUpdateDisplay 在 TaskUpdate 静态展示之上附带 taskId / status / subject：
//   - in_progress 且带 activeForm 时，直接以 activeForm 作为动作描述
//     （它是 LLM 提供的现在进行时形式，最贴合当前正在做的事）；
//   - 带 taskId 时通过 main.planTaskBackend 反查任务 subject，
//     命中后动作描述成为 "正在完成任务 #7：实现登录" 等可读形式；
//     查不到 subject 则回退到纯 #id 形式；
//   - 无 taskId 时回退到 base 静态展示。
func resolveTaskUpdateDisplay(arguments string, main *MainAgent) (ToolDisplay, bool) {
	base, ok := toolRegistry["TaskUpdate"]
	if !ok {
		return ToolDisplay{}, false
	}
	taskID := extractArgumentValue(arguments, "taskId")
	status := extractArgumentValue(arguments, "status")
	activeForm := extractArgumentValue(arguments, "activeForm")

	// in_progress 时的 activeForm 是供给加载动画展示的现在进行时形式，
	// 优先于静态 Action，让用户看到正在做的具体动作而不是泛化的"正在更新任务"。
	if status == "in_progress" && activeForm != "" {
		return ToolDisplay{
			NameZh:   base.NameZh,
			NameEn:   base.NameEn,
			ActionZh: activeForm,
			ActionEn: activeForm,
			Icon:     base.Icon,
		}, true
	}

	if taskID == "" {
		return base, true
	}

	// subject 不在 TaskUpdate 入参里，从 MainAgent 上缓存的 plan task backend 反查已创建的任务；
	// main 或 backend 为 nil 时静默降级到 #id 形式（测试 / 未启用 plan task 中间件时都可能为 nil）。
	subject := ""
	if main != nil && main.planTaskBackend != nil {
		subject, _ = main.planTaskBackend.lookupTaskSubject(planTaskBaseDir, taskID)
	}

	var zhPrefix, enPrefix string
	switch status {
	case "completed":
		zhPrefix, enPrefix = "完成任务", "Complete task"
	case "deleted":
		zhPrefix, enPrefix = "删除任务", "Delete task"
	case "in_progress":
		zhPrefix, enPrefix = "开始任务", "Start task"
	case "pending":
		zhPrefix, enPrefix = "挂起任务", "Pending task"
	default:
		zhPrefix, enPrefix = "更新任务", "Update task"
	}

	zhAct, enAct := buildTaskActionLabel(zhPrefix, enPrefix, taskID, subject)
	return ToolDisplay{
		NameZh:   base.NameZh,
		NameEn:   base.NameEn,
		ActionZh: zhAct,
		ActionEn: enAct,
		Icon:     base.Icon,
	}, true
}

// buildTaskActionLabel 拼装 TaskUpdate 的动作描述标签，
// 在 "正在完成任务 #7" 基础上追加可读 subject（如 "正在完成任务 #7：实现登录"）。
// subject 为空时仅保留 #id 形式，保证解析器降级可用。
func buildTaskActionLabel(zhPrefix, enPrefix, taskID, subject string) (zh, en string) {
	zh = zhPrefix + " #" + taskID
	en = enPrefix + " #" + taskID
	if subject == "" {
		return zh, en
	}
	zh = zh + "：" + subject
	en = en + ": " + subject
	return zh, en
}

var toolRegistry = map[string]ToolDisplay{
	"ls": {
		NameZh:   "文件系统",
		NameEn:   "File System",
		ActionZh: "正在列出目录",
		ActionEn: "Listing directory",
		Icon:     "📁",
	},
	"read_file": {
		NameZh:   "文件系统",
		NameEn:   "File System",
		ActionZh: "正在读取文件",
		ActionEn: "Reading file",
		Icon:     "📄",
	},
	"write_file": {
		NameZh:   "文件系统",
		NameEn:   "File System",
		ActionZh: "正在写入文件",
		ActionEn: "Writing file",
		Icon:     "✏️",
	},
	"edit_file": {
		NameZh:   "文件系统",
		NameEn:   "File System",
		ActionZh: "正在编辑文件",
		ActionEn: "Editing file",
		Icon:     "📝",
	},
	"mkdir": {
		NameZh:   "文件系统",
		NameEn:   "File System",
		ActionZh: "正在创建目录",
		ActionEn: "Creating directory",
		Icon:     "📂",
	},
	"mv": {
		NameZh:   "文件系统",
		NameEn:   "File System",
		ActionZh: "正在移动文件",
		ActionEn: "Moving file",
		Icon:     "↗️",
	},
	"rm": {
		NameZh:   "文件系统",
		NameEn:   "File System",
		ActionZh: "正在删除文件",
		ActionEn: "Deleting file",
		Icon:     "🗑️",
	},
	"cp": {
		NameZh:   "文件系统",
		NameEn:   "File System",
		ActionZh: "正在拷贝文件",
		ActionEn: "Copying file",
		Icon:     "📋",
	},
	"glob": {
		NameZh:   "文件系统",
		NameEn:   "File System",
		ActionZh: "正在搜索文件",
		ActionEn: "Searching files",
		Icon:     "🔍",
	},
	"grep": {
		NameZh:   "文件系统",
		NameEn:   "File System",
		ActionZh: "正在搜索内容",
		ActionEn: "Searching content",
		Icon:     "🔎",
	},
	"execute": {
		NameZh:   "终端",
		NameEn:   "Terminal",
		ActionZh: "正在执行命令",
		ActionEn: "Executing command",
		Icon:     "💻",
	},
	"TaskCreate": {
		NameZh:   "任务管理",
		NameEn:   "Task Management",
		ActionZh: "正在创建任务",
		ActionEn: "Creating task",
		Icon:     "➕",
	},
	"TaskUpdate": {
		NameZh:   "任务管理",
		NameEn:   "Task Management",
		ActionZh: "正在更新任务",
		ActionEn: "Updating task",
		Icon:     "🔄",
	},
	"TaskGet": {
		NameZh:   "任务管理",
		NameEn:   "Task Management",
		ActionZh: "正在获取任务",
		ActionEn: "Getting task",
		Icon:     "📋",
	},
	"TaskList": {
		NameZh:   "任务管理",
		NameEn:   "Task Management",
		ActionZh: "正在列出任务",
		ActionEn: "Listing tasks",
		Icon:     "📑",
	},
	"skill": {
		NameZh:   "技能加载",
		NameEn:   "Skill Load",
		ActionZh: "正在加载技能",
		ActionEn: "Loading skills",
		Icon:     "⚡",
	},
	"goraven_visual_understand": {
		NameZh:   "多模态识别",
		NameEn:   "Visual Understand",
		ActionZh: "正在分析文件内容",
		ActionEn: "Analyzing file content",
		Icon:     "👁️",
	},
	"goraven_doc_parse": {
		NameZh:   "文档解析",
		NameEn:   "Doc Parse",
		ActionZh: "正在解析文档",
		ActionEn: "Parsing document",
		Icon:     "📑",
	},
	"goraven_web_fetch": {
		NameZh:   "网页获取",
		NameEn:   "Web Fetch",
		ActionZh: "正在获取网页内容",
		ActionEn: "Fetching web page",
		Icon:     "🌐",
	},
	"goraven_create_automation_task": {
		NameZh:   "自动化任务",
		NameEn:   "Automation Task",
		ActionZh: "正在创建自动化任务",
		ActionEn: "Creating automation task",
		Icon:     "⏰",
	},
	"goraven_list_automation_tasks": {
		NameZh:   "自动化任务",
		NameEn:   "Automation Task",
		ActionZh: "正在查询自动化任务",
		ActionEn: "Listing automation tasks",
		Icon:     "🗓️",
	},
	"goraven_get_automation_task": {
		NameZh:   "自动化任务",
		NameEn:   "Automation Task",
		ActionZh: "正在查询自动化任务详情",
		ActionEn: "Getting automation task",
		Icon:     "⏰",
	},
	"goraven_update_automation_task": {
		NameZh:   "自动化任务",
		NameEn:   "Automation Task",
		ActionZh: "正在更新自动化任务",
		ActionEn: "Updating automation task",
		Icon:     "⏰",
	},
	"goraven_get_current_time": {
		NameZh:   "时间查询",
		NameEn:   "Get Current Time",
		ActionZh: "正在获取当前时间",
		ActionEn: "Getting current time",
		Icon:     "🕐",
	},
	"sub_agent": {
		NameZh:   "子代理",
		NameEn:   "Sub Agent",
		ActionZh: "正在委托子代理执行任务",
		ActionEn: "Delegating task to sub-agent",
		Icon:     "🤖",
	},
}

// resolveShellCommandDisplay 取命令行首个空白分隔 token，按 switch 匹配细分展示。
// 命中返回对应展示与 true；未命中返回零值与 false（由调用方回退到 toolRegistry["execute"] 终端默认）。
// ToolDisplay 字段顺序：NameZh, NameEn, ActionZh, ActionEn, Icon。
func resolveShellCommandDisplay(command string) (ToolDisplay, bool) {
	command = strings.TrimSpace(command)
	if command == "" {
		return ToolDisplay{}, false
	}
	token := command
	if i := strings.IndexAny(command, " \t\n"); i >= 0 {
		token = command[:i]
	}
	var d ToolDisplay
	switch token {
	// 网络
	case "curl", "wget":
		d = ToolDisplay{"网络请求", "HTTP Request", "正在发送网络请求", "Sending network request", "📡"}
	case "ping":
		d = ToolDisplay{"网络请求", "Network", "正在检测连通性", "Pinging host", "📡"}
	// 版本控制
	case "git":
		d = ToolDisplay{"版本控制", "Git", "正在执行 Git 命令", "Running git command", "🔀"}
	// 语言 / 包管理
	case "python", "python3", "pip", "uv":
		d = ToolDisplay{"Python", "Python", "正在运行 Python", "Running Python", "🐍"}
	case "npm", "pnpm", "yarn", "npx":
		d = ToolDisplay{"包管理", "Package Manager", "正在执行包管理", "Running package manager", "📦"}
	case "go":
		d = ToolDisplay{"Go", "Go", "正在运行 Go", "Running Go", "🐹"}
	case "cargo":
		d = ToolDisplay{"Rust", "Rust", "正在运行 Cargo", "Running Cargo", "🦀"}
	// 进程
	case "ps":
		d = ToolDisplay{"进程", "Processes", "正在查看进程", "Listing processes", "📊"}
	case "top", "htop":
		d = ToolDisplay{"进程", "Processes", "正在监控进程", "Monitoring processes", "📊"}
	case "kill", "killall":
		d = ToolDisplay{"进程", "Processes", "正在终止进程", "Killing process", "📊"}
	// 磁盘
	case "df":
		d = ToolDisplay{"磁盘", "Disk Usage", "正在查看磁盘容量", "Checking disk capacity", "💾"}
	case "du":
		d = ToolDisplay{"磁盘", "Disk Usage", "正在统计目录大小", "Calculating directory size", "💾"}
	// 文件查看
	case "cat", "less", "head", "tail":
		d = ToolDisplay{"文件查看", "File Viewer", "正在查看文件内容", "Viewing file content", "📄"}
	// 文件操作
	case "cp":
		d = ToolDisplay{"文件操作", "File Operation", "正在拷贝文件", "Copying file", "📁"}
	case "mv":
		d = ToolDisplay{"文件操作", "File Operation", "正在移动文件", "Moving file", "📁"}
	case "rm":
		d = ToolDisplay{"文件操作", "File Operation", "正在删除文件", "Deleting file", "📁"}
	case "mkdir":
		d = ToolDisplay{"文件操作", "File Operation", "正在创建目录", "Creating directory", "📁"}
	case "touch":
		d = ToolDisplay{"文件操作", "File Operation", "正在创建文件", "Creating file", "📁"}
	// 文件系统
	case "ls":
		d = ToolDisplay{"文件系统", "File System", "正在列出目录", "Listing directory", "📁"}
	case "find":
		d = ToolDisplay{"文件系统", "File System", "正在查找文件", "Finding files", "📁"}
	case "which":
		d = ToolDisplay{"文件系统", "File System", "正在查找命令路径", "Locating command", "📁"}
	// 文本处理
	case "grep":
		d = ToolDisplay{"文本处理", "Text Processing", "正在搜索内容", "Searching content", "🔍"}
	case "awk":
		d = ToolDisplay{"文本处理", "Text Processing", "正在处理文本", "Processing text", "🔍"}
	case "sed":
		d = ToolDisplay{"文本处理", "Text Processing", "正在编辑文本", "Editing text", "🔍"}
	// 压缩解压
	case "tar":
		d = ToolDisplay{"压缩解压", "Archive", "正在打包解包", "Archiving", "📦"}
	case "gzip", "gunzip":
		d = ToolDisplay{"压缩解压", "Archive", "正在压缩解压", "Compressing/decompressing", "📦"}
	case "zip", "unzip":
		d = ToolDisplay{"压缩解压", "Archive", "正在压缩解压", "Compressing/decompressing", "📦"}
	// 权限管理
	case "chmod":
		d = ToolDisplay{"权限管理", "Permissions", "正在修改权限", "Changing permissions", "🔐"}
	case "chown":
		d = ToolDisplay{"权限管理", "Permissions", "正在修改所有者", "Changing owner", "🔐"}
	default:
		return ToolDisplay{}, false
	}
	return d, true
}

func GetLanguage() string {
	return config.Get().GetLanguage()
}

func buildSSEEventToolFromDisplay(name string, display ToolDisplay) *SSEEventTool {
	useChinese := isChineseLanguage()
	toolName := display.NameEn
	action := display.ActionEn
	if useChinese {
		toolName = display.NameZh
		action = display.ActionZh
	}

	toolContent := toolName
	return &SSEEventTool{
		Name:        name,
		DisplayName: toolContent,
		Icon:        display.Icon,
		Action:      action,
	}
}

func extractArgumentValue(argumentsJSON, key string) string {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argumentsJSON), &args); err != nil {
		return ""
	}
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func isChineseLanguage() bool {
	return GetLanguage() == "zh"
}
