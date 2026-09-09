package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/core/iface"
	"goraven/core/plugin"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type recordingDailyStatsRepo struct {
	toolNames []string
}

func (r *recordingDailyStatsRepo) AddDailyStats(string, int, int, int) error {
	return nil
}

func (r *recordingDailyStatsRepo) AddToolDailyStats(_ string, _ string, toolName string) error {
	r.toolNames = append(r.toolNames, toolName)
	return nil
}

func TestNonStreamLoopDoesNotEmitSubAgentToolEvent(t *testing.T) {
	stats := &recordingDailyStatsRepo{}
	runner := &MainRunner{
		mainAgent: &MainAgent{
			BaseAgent: BaseAgent{Plugins: &plugin.Plugins{}},
			param: AgentParam{
				Session:        &po.Session{UserId: "user-1"},
				DailyStatsRepo: stats,
			},
		},
	}
	toolCalls := []schema.ToolCall{{
		ID: "call-1",
		Function: schema.FunctionCall{
			Name:      "some_tool",
			Arguments: "{}",
		},
	}}
	mv := &adk.MessageVariant{
		Message: schema.AssistantMessage("", toolCalls),
		Role:    schema.Assistant,
	}

	runner.nonStreamLoop(subAgentName, mv, &[]string{})

	if len(stats.toolNames) != 0 {
		t.Fatalf("sub-agent tool event should not be emitted, got %v", stats.toolNames)
	}
}

type nonStreamRunnerModel struct {
	iface.ConversationHeaderKeyHolder
	iface.VisualSupportHolder
}

func (m *nonStreamRunnerModel) Generate(context.Context, []*schema.Message, ...model.Option) (*schema.Message, error) {
	msg := schema.AssistantMessage("reply", nil)
	msg.ResponseMeta = &schema.ResponseMeta{
		Usage: &schema.TokenUsage{
			PromptTokens: 11,
			PromptTokenDetails: schema.PromptTokenDetails{
				CachedTokens: 7,
			},
			CompletionTokens: 5,
		},
	}
	return msg, nil
}

func (m *nonStreamRunnerModel) Stream(context.Context, []*schema.Message, ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

func (m *nonStreamRunnerModel) ModelName() string       { return "non-stream-test" }
func (m *nonStreamRunnerModel) Provider() string        { return "test" }
func (m *nonStreamRunnerModel) ContextLength() int      { return 1000 }
func (m *nonStreamRunnerModel) Format() iface.APIFormat { return iface.APIFormatOpenAI }

type nonStreamRunnerRepo struct {
	cachedTokens int
	savedMsg     *po.Message
	done         chan struct{}
}

func (r *nonStreamRunnerRepo) SaveChatMessage(_ string, m *po.Message) error {
	if m.RoleType == po.RoleTypeAssistant {
		r.savedMsg = m
	}
	return nil
}
func (r *nonStreamRunnerRepo) GetChatMessages(string) ([]*po.Message, error) {
	return nil, nil
}
func (r *nonStreamRunnerRepo) AddSessionTokens(_ string, _, _, cached int) error {
	r.cachedTokens = cached
	return nil
}
func (r *nonStreamRunnerRepo) SetContextTokens(string, int) error           { return nil }
func (r *nonStreamRunnerRepo) UpdateSessionStatus(string, int) error        { return nil }
func (r *nonStreamRunnerRepo) MarkSessionCompressed(string, []string) error { return nil }

func TestNonStreamRunnerPersistsCachedPromptTokens(t *testing.T) {
	repo := &nonStreamRunnerRepo{done: make(chan struct{})}
	model := &nonStreamRunnerModel{}
	main := &MainAgent{
		BaseAgent: BaseAgent{
			Plugins: &plugin.Plugins{},
			sysCfg:  repository.NewDefaultSystemConfig(),
		},
		msgRepo: repo,
		param: AgentParam{
			Session:   &po.Session{SessionId: "session-1", UserId: "user-1"},
			MsgRepo:   repo,
			ChatModel: model,
			SysCfg:    repository.NewDefaultSystemConfig(),
		},
	}
	chatAgent, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:  mainAgentName,
		Model: model,
	})
	if err != nil {
		t.Fatalf("create chat model agent: %v", err)
	}
	runner := newMainRunner(main, chatAgent, false)
	runner.OnComplete(func(*RunnerCompleteEvent) { close(repo.done) })

	if err := runner.Query(context.Background(), "query", nil); err != nil {
		t.Fatalf("query: %v", err)
	}
	select {
	case <-repo.done:
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not complete")
	}

	if repo.cachedTokens != 7 {
		t.Fatalf("cached prompt tokens = %d, want 7", repo.cachedTokens)
	}
	if repo.savedMsg == nil {
		t.Fatal("assistant message was not saved")
	}
	if repo.savedMsg.PromptCachedTokensCount != 7 {
		t.Fatalf("saved PromptCachedTokensCount = %d, want 7", repo.savedMsg.PromptCachedTokensCount)
	}
}

type mediaSaveRepo struct {
	saved *po.Message
}

func (r *mediaSaveRepo) SaveChatMessage(_ string, m *po.Message) error { r.saved = m; return nil }
func (r *mediaSaveRepo) GetChatMessages(string) ([]*po.Message, error) { return nil, nil }
func (r *mediaSaveRepo) AddSessionTokens(string, int, int, int) error { return nil }
func (r *mediaSaveRepo) SetContextTokens(string, int) error           { return nil }
func (r *mediaSaveRepo) UpdateSessionStatus(string, int) error        { return nil }
func (r *mediaSaveRepo) MarkSessionCompressed(string, []string) error { return nil }

func TestSaveQueryPersistsMedia(t *testing.T) {
	repo := &mediaSaveRepo{}
	runner := &MainRunner{
		mainAgent: &MainAgent{
			msgRepo: repo,
			param: AgentParam{
				Session: &po.Session{SessionId: "s1", UserId: "u1"},
				MsgRepo: repo,
			},
		},
		RoundId: "r1",
	}
	if err := runner.saveQuery("hello", []MediaItem{
		{Type: "image", Path: "/temp/a.jpg", URL: "https://x/api/hfs/public/a.jpg"},
	}); err != nil {
		t.Fatalf("saveQuery: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("message not saved")
	}
	if repo.saved.Content != "hello" {
		t.Fatalf("content = %q, want hello", repo.saved.Content)
	}
	if repo.saved.Media == "" {
		t.Fatal("media not persisted")
	}
	var items []MediaItem
	if err := json.Unmarshal([]byte(repo.saved.Media), &items); err != nil || len(items) != 1 || items[0].Type != "image" {
		t.Fatalf("media json mismatch: %s err=%v", repo.saved.Media, err)
	}
}

func TestSaveQueryWithoutMedia(t *testing.T) {
	repo := &mediaSaveRepo{}
	runner := &MainRunner{
		mainAgent: &MainAgent{
			msgRepo: repo,
			param: AgentParam{
				Session: &po.Session{SessionId: "s1", UserId: "u1"},
				MsgRepo: repo,
			},
		},
		RoundId: "r1",
	}
	if err := runner.saveQuery("hello", nil); err != nil {
		t.Fatalf("saveQuery: %v", err)
	}
	if repo.saved == nil || repo.saved.Media != "" {
		t.Fatalf("media should be empty: %+v", repo.saved)
	}
}
