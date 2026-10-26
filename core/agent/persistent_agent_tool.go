package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const persistentAgentToolParamRequest = "request"
const persistentAgentToolParamNew = "new"

type PersistentAgentTool struct {
	agent           adk.Agent
	history         []adk.Message
	mu              sync.Mutex
	enableStreaming bool
}

func NewPersistentAgentTool(enableStreaming bool, agent adk.Agent) *PersistentAgentTool {
	return &PersistentAgentTool{
		enableStreaming: enableStreaming,
		agent:           agent,
	}
}

func (t *PersistentAgentTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	name := t.agent.Name(ctx)
	if name == "" {
		return nil, fmt.Errorf("persistent agent tool requires a non-empty agent Name")
	}
	desc := t.agent.Description(ctx)
	if desc == "" {
		return nil, fmt.Errorf("persistent agent tool requires a non-empty agent Description")
	}

	return &schema.ToolInfo{
		Name: name,
		Desc: desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			persistentAgentToolParamRequest: {
				Desc:     "task description to be processed by the sub-agent",
				Required: true,
				Type:     schema.String,
			},
			persistentAgentToolParamNew: {
				Desc:     "if true, start a new conversation context, discarding previous history",
				Required: false,
				Type:     schema.Boolean,
			},
		}),
	}, nil
}

func (t *PersistentAgentTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type request struct {
		Request string `json:"request"`
		New     bool   `json:"new"`
	}

	req := &request{}
	err := sonic.UnmarshalString(argumentsInJSON, req)
	if err != nil {
		return "", fmt.Errorf("persistent agent tool: failed to parse arguments: %w", err)
	}

	t.mu.Lock()
	if req.New {
		t.history = nil
	}
	messages := make([]adk.Message, len(t.history))
	copy(messages, t.history)
	messages = append(messages, schema.UserMessage(req.Request))
	t.mu.Unlock()

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           t.agent,
		EnableStreaming: t.enableStreaming,
	})

	iter := runner.Run(ctx, messages)

	var lastContent string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			return "", fmt.Errorf("persistent agent tool: agent execution error: %w", event.Err)
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			mv := event.Output.MessageOutput
			msg, msgErr := mv.GetMessage()
			if msgErr != nil {
				continue
			}
			if msg != nil && mv.Role == schema.Assistant {
				lastContent = msg.Content
			}
		}
	}

	t.mu.Lock()
	t.history = append(t.history, schema.UserMessage(req.Request))
	if lastContent != "" {
		t.history = append(t.history, schema.AssistantMessage(lastContent, nil))
	}
	t.mu.Unlock()

	return lastContent, nil
}
