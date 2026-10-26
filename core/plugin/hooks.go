package plugin

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// ======================= Round Hook =======================

// RoundHook intercepts the lifecycle of a single conversation round (one user query → agent reply).
type RoundHook interface {
	Plugin
	// BeforeRound is called after history is loaded and compressed, but before the LLM is invoked.
	// Plugins can modify ctx.Messages to inject system prompts or alter history.
	BeforeRound(ctx *RoundContext) error
	// AfterRound is called after the LLM returns but before the reply is persisted.
	// Plugins can read the reply and reasoning content, e.g. for auditing or content filtering.
	AfterRound(ctx *RoundContext) error
}

// RoundContext carries the full context of a single query round.
type RoundContext struct {
	SessionID        string         // session identifier
	UserID           string         // user identifier
	RoundID          string         // unique round identifier
	Query            string         // user input for this round (read-only in BeforeRound, nil in AfterRound)
	Messages         []*schema.Message // conversation history (BeforeRound: readable and modifiable; AfterRound: read-only)
	Reply            *string        // agent reply content (AfterRound: readable and modifiable)
	ReasoningContent *string        // full reasoning/thinking chain (AfterRound: read-only)
	Stopped          bool           // whether the round was terminated
}

// ======================= Agent Lifecycle Hook =======================

// AgentLifecycleHook intercepts agent creation.
type AgentLifecycleHook interface {
	Plugin
	// BeforeAgentCreate is called after the agent object is built but before NewRunner.
	// Plugins can inject tools, MCP configurations, or middleware via the context.
	BeforeAgentCreate(ctx *AgentCreateContext) error
}

// AgentCreateContext carries the context during agent creation.
// Plugins can inject tools and middleware using the AddTool and AddMiddleware callbacks.
type AgentCreateContext struct {
	SessionID     string         // session identifier
	UserID        string         // user identifier
	AddTool       func(tool any) // call to append a custom tool to the agent
	AddMiddleware func(mw any)   // call to append a middleware to the agent
}

// ======================= Tool Hook =======================

// ToolHook intercepts tool execution.
type ToolHook interface {
	Plugin
	// BeforeTool is called before a tool executes.
	// modifiedArgs can replace the original arguments.
	// skip=true short-circuits execution and returns modifiedArgs as the tool result.
	BeforeTool(ctx context.Context, toolName string, args string) (modifiedArgs string, skip bool)
	// AfterTool is called after a tool completes. The result can be modified.
	AfterTool(ctx context.Context, toolName string, args string, result string, execErr error) (modifiedResult string)
}

// ======================= SSE Hook =======================

// SSEHook intercepts Server-Sent Events pushed to the frontend.
type SSEHook interface {
	Plugin
	// OnSSEEvent is called for each SSE event.
	// Return nil to filter out (suppress) this event from reaching the frontend.
	// Return a modified event to replace the original.
	OnSSEEvent(ctx context.Context, event *SSEEventData) *SSEEventData
}

// SSEEventData 单个 SSE 事件的数据，前端根据 Type 读取对应的负载字段
type SSEEventData struct {
	Type    string            // "reasoning", "content", "tool", "end", "retry", "context"
	Content string            // 事件内容文本（reasoning/content/end 事件使用）
	Tool    *SSEEventToolData // 工具展示信息（仅 "tool" 类型事件使用）
	Retry   *SSERetryInfoData // 模型重试信息（仅 "retry" 类型事件使用）
	Context *SSEContextData   // 上下文信息（仅 "context" 类型事件使用）
}

// SSEEventToolData 工具事件的展示元数据
type SSEEventToolData struct {
	Name        string // 工具名称
	DisplayName string // 本地化展示名称
	Icon        string // emoji 图标
	Action      string // 本地化动作描述
}

// SSERetryInfoData 模型重试事件的元数据
type SSERetryInfoData struct {
	MaxRetries int    // 最大重试次数
	Attempt    int    // 当前重试次数（从1开始）
	Error      string // 触发的错误信息
}

// SSEContextData 上下文更新事件的数据
type SSEContextData struct {
	Tokens int // 当前 prompt tokens
	Limit  int // 模型最大上下文长度
}
