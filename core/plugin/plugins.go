package plugin

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"
)

// Plugins holds a per-agent set of instantiated plugins.
// Each agent gets its own Plugins via CreatePlugins().
type Plugins struct {
	plugins        []Plugin
	lifecycleHooks []AgentLifecycleHook
	roundHooks     []RoundHook
	toolHooks      []ToolHook
	sseHooks       []SSEHook
}

// ---- Agent Lifecycle ----

func (p *Plugins) FireAgentBeforeCreate(ctx *AgentCreateContext) error {
	for _, h := range p.lifecycleHooks {
		if err := h.BeforeAgentCreate(ctx); err != nil {
			return fmt.Errorf("plugin %s: %w", h.Name(), err)
		}
	}
	return nil
}

// ---- Round ----

func (p *Plugins) FireBeforeRound(ctx *RoundContext) error {
	for _, h := range p.roundHooks {
		if err := h.BeforeRound(ctx); err != nil {
			return fmt.Errorf("plugin %s: %w", h.Name(), err)
		}
	}
	return nil
}

func (p *Plugins) FireAfterRound(ctx *RoundContext) error {
	for _, h := range p.roundHooks {
		if err := h.AfterRound(ctx); err != nil {
			return fmt.Errorf("plugin %s: %w", h.Name(), err)
		}
	}
	return nil
}

// ---- Tool ----

func (p *Plugins) FireBeforeTool(ctx context.Context, toolName string, args string) (string, bool) {
	currentArgs := args
	for _, h := range p.toolHooks {
		modified, skip := h.BeforeTool(ctx, toolName, currentArgs)
		if skip {
			return modified, true
		}
		currentArgs = modified
	}
	return currentArgs, false
}

func (p *Plugins) FireAfterTool(ctx context.Context, toolName string, args string, result string, execErr error) string {
	currentResult := result
	for _, h := range p.toolHooks {
		currentResult = h.AfterTool(ctx, toolName, args, currentResult, execErr)
	}
	return currentResult
}

func (p *Plugins) HasToolHooks() bool {
	return len(p.toolHooks) > 0
}

func (p *Plugins) NewToolInterceptorMiddleware() compose.ToolMiddleware {
	a := &toolInterceptorAdapter{plugins: p}
	return compose.ToolMiddleware{
		Invokable:          a.wrapInvokable,
		Streamable:         a.wrapStreamable,
		EnhancedInvokable:  a.wrapEnhancedInvokable,
		EnhancedStreamable: a.wrapEnhancedStreamable,
	}
}

// ---- SSE ----

func (p *Plugins) FireSSEHook(ctx context.Context, event *SSEEventData) *SSEEventData {
	current := event
	for _, h := range p.sseHooks {
		current = h.OnSSEEvent(ctx, current)
		if current == nil {
			return nil
		}
	}
	return current
}
