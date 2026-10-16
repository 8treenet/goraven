package agent

import (
	"context"
	"errors"
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

type blockedLifecycleModel struct {
	nonStreamRunnerModel
	release <-chan struct{}
}

func (m *blockedLifecycleModel) Generate(ctx context.Context, messages []*schema.Message, options ...model.Option) (*schema.Message, error) {
	select {
	case <-m.release:
		return m.nonStreamRunnerModel.Generate(ctx, messages, options...)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func lifecycleRunner(t *testing.T, chatModel iface.BaseChatModel, repo iface.MessageRepo) *MainRunner {
	t.Helper()
	main := &MainAgent{
		BaseAgent: BaseAgent{Plugins: &plugin.Plugins{}, sysCfg: repository.NewDefaultSystemConfig()},
		msgRepo:   repo,
		param: AgentParam{
			Session: &po.Session{SessionId: t.Name(), UserId: "lifecycle-test-user"},
			MsgRepo: repo, ChatModel: chatModel, SysCfg: repository.NewDefaultSystemConfig(),
		},
	}
	chatAgent, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{Name: mainAgentName, Model: chatModel})
	if err != nil {
		t.Fatal(err)
	}
	return newMainRunner(main, chatAgent, false)
}

func TestQueryRegistersBeforeReturningAndClearsOnCompletion(t *testing.T) {
	release := make(chan struct{})
	done := make(chan struct{})
	runner := lifecycleRunner(t, &blockedLifecycleModel{release: release}, &nonStreamRunnerRepo{})
	runner.OnComplete(func(*RunnerCompleteEvent) { close(done) })
	if err := runner.Query(context.Background(), "query", nil); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		close(release)
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("query did not complete")
		}
		if _, ok := GetRunner(t.Name()); ok {
			DeleteRunner(t.Name())
			t.Error("completed query remained registered")
		}
	})
	if actual, ok := GetRunner(t.Name()); !ok || actual != runner {
		t.Fatal("query returned before its runner was registered")
	}
}

type failingLifecycleHistory struct {
	nonStreamRunnerRepo
}

func (*failingLifecycleHistory) GetChatMessages(string) ([]*po.Message, error) {
	return nil, errors.New("history unavailable")
}

func TestQueryHistoryFailureDoesNotRegister(t *testing.T) {
	runner := lifecycleRunner(t, &nonStreamRunnerModel{}, &failingLifecycleHistory{})
	if err := runner.Query(context.Background(), "query", nil); err == nil {
		t.Fatal("expected synchronous history failure")
	}
	if _, ok := GetRunner(t.Name()); ok {
		DeleteRunner(t.Name())
		t.Fatal("failed query remained registered")
	}
}
