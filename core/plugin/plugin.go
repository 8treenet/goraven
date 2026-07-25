package plugin

import "context"

type Plugin interface {
	Name() string
	Version() string
}

type RoundContext struct {
	SessionID        string
	RoundID          string
	Query            string
	Reply            *string
	ReasoningContent *string
	Stopped          bool
}

type SSEEventData struct {
	Type string
}

type AgentCreateContext struct {
	SessionID string
	UserID    string
}

type RoundHook interface {
	BeforeRound(ctx *RoundContext) error
	AfterRound(ctx *RoundContext) error
}

type ToolHook interface {
	BeforeTool(ctx context.Context, toolName string, args string) (string, bool)
	AfterTool(ctx context.Context, toolName string, args string, result string, execErr error) string
}

type SSEHook interface {
	OnSSEEvent(ctx context.Context, event *SSEEventData) *SSEEventData
}

type AgentLifecycleHook interface {
	BeforeAgentCreate(ctx *AgentCreateContext) error
}

type PluginInfo struct {
	Name    string
	Version string
}

type PluginFactory func() Plugin

var factories []PluginFactory

func Register(factory PluginFactory) {
	factories = append(factories, factory)
}

func GetAllPluginInfo() []PluginInfo {
	var result []PluginInfo
	for _, f := range factories {
		p := f()
		result = append(result, PluginInfo{
			Name:    p.Name(),
			Version: p.Version(),
		})
	}
	return result
}
