package audit

import (
	"context"
	"fmt"
	"log"

	"goraven/core/plugin"
)

func Register() {
	plugin.Register(func() plugin.Plugin { return &AuditPlugin{} })
}

type AuditPlugin struct {
	logs []string
}

func (p *AuditPlugin) Name() string    { return "goraven/audit" }
func (p *AuditPlugin) Version() string { return "1.0.0" }

// ---- RoundHook ----

func (p *AuditPlugin) BeforeRound(ctx *plugin.RoundContext) error {
	p.logs = append(p.logs, fmt.Sprintf("[BeforeRound] session=%s round=%s query=%q", ctx.SessionID, ctx.RoundID, ctx.Query))
	return nil
}

func (p *AuditPlugin) AfterRound(ctx *plugin.RoundContext) error {
	replyLen := 0
	if ctx.Reply != nil {
		replyLen = len(*ctx.Reply)
	}
	reasoningLen := 0
	if ctx.ReasoningContent != nil {
		reasoningLen = len(*ctx.ReasoningContent)
	}
	p.logs = append(p.logs, fmt.Sprintf("[AfterRound] reply_len=%d reasoning_len=%d stopped=%v", replyLen, reasoningLen, ctx.Stopped))

	// 打印所有收集的日志
	for _, l := range p.logs {
		log.Println("[audit]", l)
	}
	p.logs = nil
	return nil
}

// ---- ToolHook ----

func (p *AuditPlugin) BeforeTool(ctx context.Context, toolName string, args string) (string, bool) {
	p.logs = append(p.logs, fmt.Sprintf("[BeforeTool] tool=%s args_len=%d", toolName, len(args)))
	return args, false
}

func (p *AuditPlugin) AfterTool(ctx context.Context, toolName string, args string, result string, execErr error) string {
	if execErr != nil {
		p.logs = append(p.logs, fmt.Sprintf("[AfterTool] tool=%s err=%v", toolName, execErr))
	} else {
		p.logs = append(p.logs, fmt.Sprintf("[AfterTool] tool=%s result_len=%d", toolName, len(result)))
	}
	return result
}

// ---- SSEHook ----

func (p *AuditPlugin) OnSSEEvent(ctx context.Context, event *plugin.SSEEventData) *plugin.SSEEventData {
	p.logs = append(p.logs, fmt.Sprintf("[OnSSEEvent] type=%s", event.Type))
	return event
}

// ---- AgentLifecycleHook ----

func (p *AuditPlugin) BeforeAgentCreate(ctx *plugin.AgentCreateContext) error {
	p.logs = append(p.logs, fmt.Sprintf("[BeforeAgentCreate] session=%s user=%s", ctx.SessionID, ctx.UserID))
	return nil
}
