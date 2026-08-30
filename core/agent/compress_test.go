package agent

import (
	"context"
	"fmt"
	"goraven/backend/po"
	"goraven/core/iface"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type mockChatModel struct {
	response string
}

func (m *mockChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	msg := schema.AssistantMessage(m.response, nil)
	msg.ResponseMeta = &schema.ResponseMeta{
		Usage: &schema.TokenUsage{
			PromptTokens:     100,
			CompletionTokens: 50,
		},
	}
	return msg, nil
}

func (m *mockChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

func (m *mockChatModel) ModelName() string { return "mock" }
func (m *mockChatModel) Provider() string  { return "mock" }
func (m *mockChatModel) ContextLength() int {
	return 200 * 1024
}

func (m *mockChatModel) Format() iface.APIFormat {
	return iface.APIFormatOpenAI
}

type mockMsgRepo struct {
	markedRoundIds []string
	savedMessages  []*po.Message
}

func (m *mockMsgRepo) SaveChatMessage(sessionId string, msg *po.Message) error {
	m.savedMessages = append(m.savedMessages, msg)
	return nil
}

func (m *mockMsgRepo) GetChatMessages(sessionId string) ([]*po.Message, error) {
	return nil, nil
}

func (m *mockMsgRepo) AddSessionTokens(sessionId string, promptTokens, completionTokens, promptCachedTokens int) error {
	return nil
}

func (m *mockMsgRepo) SetContextTokens(sessionId string, tokens int) error {
	return nil
}

func (m *mockChatModel) UpdateSessionStatus(sessionId string, status int) error {
	return nil
}

func (m *mockMsgRepo) MarkSessionCompressed(sessionId string, roundIds []string) error {
	m.markedRoundIds = append(m.markedRoundIds, roundIds...)
	return nil
}

func (m *mockMsgRepo) UpdateSessionStatus(sessionId string, status int) error {
	return nil
}

func makeRoundMsgs(roundId string, role schema.RoleType, contents []string) []*schema.Message {
	var msgs []*schema.Message
	for i, c := range contents {
		msg := &schema.Message{Role: role, Content: c}
		msg.Extra = map[string]any{"roundId": roundId, "timestamp": int64(1000 + i), "msgId": roundId}
		msgs = append(msgs, msg)
	}
	return msgs
}

func makeSummaryMsg(content string) *schema.Message {
	msg := schema.UserMessage(content)
	msg.Extra = map[string]any{"isSummary": true, "timestamp": int64(0), "msgId": "old_summary"}
	return msg
}

func printHistory(history []*schema.Message) {
	for i, msg := range history {
		roundId, _ := msg.Extra["roundId"].(string)
		isSummary, _ := msg.Extra["isSummary"].(bool)
		prefix := ""
		if isSummary {
			prefix = "[摘要] "
		} else if roundId != "" {
			prefix = fmt.Sprintf("[轮次%s] ", roundId)
		}
		fmt.Printf("  [%d] %s%-10s %q\n", i, prefix, msg.Role, msg.Content)
	}
}

// go test -test.fullpath=true -timeout 30s -run ^TestCompress_Basic$ goraven/core/agent -v -count=1
func TestCompress_Basic(t *testing.T) {
	repo := &mockMsgRepo{}
	chatModel := &mockChatModel{response: "这是压缩后的摘要内容"}
	c := &Compress{
		chatModel:  chatModel,
		threshold:  1,
		msgRepo:    repo,
		sessionId:  "test-session",
		keepRounds: 2,
	}

	round1 := makeRoundMsgs("r1", schema.User, []string{"用户问题1"})
	round1 = append(round1, makeRoundMsgs("r1", schema.Assistant, []string{"助手回答1"})...)
	round2 := makeRoundMsgs("r2", schema.User, []string{"用户问题2"})
	round2 = append(round2, makeRoundMsgs("r2", schema.Assistant, []string{"助手回答2"})...)
	round3 := makeRoundMsgs("r3", schema.User, []string{"用户问题3"})
	round3 = append(round3, makeRoundMsgs("r3", schema.Assistant, []string{"助手回答3"})...)

	var history []*schema.Message
	history = append(history, round1...)
	history = append(history, round2...)
	history = append(history, round3...)

	result, err := c.DoCompress(context.Background(), nil, history)
	if err != nil {
		t.Fatalf("DoCompress error: %v", err)
	}

	fmt.Println("=== TestCompress_Basic ===")
	fmt.Println("压缩前:")
	printHistory(history)
	fmt.Println("压缩后:")
	printHistory(result)

	if !c.result.Compressed {
		t.Fatal("should be compressed")
	}

	// 应包含: 1个新摘要 + round2 + round3 = 5条消息
	expectedLen := 1 + len(round2) + len(round3)
	if len(result) != expectedLen {
		t.Fatalf("expected %d messages, got %d", expectedLen, len(result))
	}

	// 新摘要应在最前面
	if result[0].Content != "[对话摘要]\n这是压缩后的摘要内容" {
		t.Fatalf("first message should be new summary, got: %s", result[0].Content)
	}

	// 应标记 round1 为已压缩
	if len(repo.markedRoundIds) != 1 || repo.markedRoundIds[0] != "r1" {
		t.Fatalf("expected marked roundIds [r1], got %v", repo.markedRoundIds)
	}

	// 应保存了摘要消息到数据库
	if len(repo.savedMessages) != 1 || repo.savedMessages[0].RoleType != po.RoleTypeSummary {
		t.Fatalf("expected 1 summary message saved, got %v", repo.savedMessages)
	}
}

// go test -test.fullpath=true -timeout 30s -run ^TestCompress_WithOldSummary$ goraven/core/agent -v -count=1
func TestCompress_WithOldSummary(t *testing.T) {
	repo := &mockMsgRepo{}
	chatModel := &mockChatModel{response: "新的摘要内容"}
	c := &Compress{
		chatModel:  chatModel,
		threshold:  1,
		msgRepo:    repo,
		sessionId:  "test-session",
		keepRounds: 1,
	}

	// [旧摘要, round1, round2, round3]
	oldSummary := makeSummaryMsg("旧的摘要内容")
	round1 := makeRoundMsgs("r1", schema.User, []string{"用户问题1"})
	round1 = append(round1, makeRoundMsgs("r1", schema.Assistant, []string{"助手回答1"})...)
	round2 := makeRoundMsgs("r2", schema.User, []string{"用户问题2"})
	round2 = append(round2, makeRoundMsgs("r2", schema.Assistant, []string{"助手回答2"})...)
	round3 := makeRoundMsgs("r3", schema.User, []string{"用户问题3"})
	round3 = append(round3, makeRoundMsgs("r3", schema.Assistant, []string{"助手回答3"})...)

	var history []*schema.Message
	history = append(history, oldSummary)
	history = append(history, round1...)
	history = append(history, round2...)
	history = append(history, round3...)

	result, err := c.DoCompress(context.Background(), nil, history)
	if err != nil {
		t.Fatalf("DoCompress error: %v", err)
	}

	fmt.Println("\n=== TestCompress_WithOldSummary ===")
	fmt.Println("压缩前:")
	printHistory(history)
	fmt.Println("压缩后:")
	printHistory(result)

	if !c.result.Compressed {
		t.Fatal("should be compressed")
	}

	// 应包含: 旧摘要 + 新摘要 + round3 = 5条消息 (keepRounds=1, r1和r2被压缩)
	expectedLen := 1 + 1 + len(round3)
	if len(result) != expectedLen {
		t.Fatalf("expected %d messages, got %d", expectedLen, len(result))
	}

	// 第一条应为旧摘要
	isSummary, _ := result[0].Extra["isSummary"].(bool)
	if !isSummary {
		t.Fatal("first message should be old summary")
	}
	if result[0].Content != "旧的摘要内容" {
		t.Fatalf("first message content should be old summary, got: %s", result[0].Content)
	}

	// 第二条应为新摘要
	isSummary2, _ := result[1].Extra["isSummary"].(bool)
	if !isSummary2 {
		t.Fatal("second message should be new summary")
	}
	if result[1].Content != "[对话摘要]\n新的摘要内容" {
		t.Fatalf("second message content should be new summary, got: %s", result[1].Content)
	}

	// 应标记 r1, r2 为已压缩，旧摘要没有 roundId 不应被标记
	if len(repo.markedRoundIds) != 2 {
		t.Fatalf("expected 2 marked roundIds, got %d: %v", len(repo.markedRoundIds), repo.markedRoundIds)
	}
}

// go test -test.fullpath=true -timeout 30s -run ^TestCompress_NotEnoughRounds$ goraven/core/agent -v -count=1
func TestCompress_NotEnoughRounds(t *testing.T) {
	repo := &mockMsgRepo{}
	chatModel := &mockChatModel{response: "摘要"}
	c := &Compress{
		chatModel:  chatModel,
		threshold:  1,
		msgRepo:    repo,
		sessionId:  "test-session",
		keepRounds: 3,
	}

	// 只有2轮，keepRounds=3，不应压缩
	round1 := makeRoundMsgs("r1", schema.User, []string{"问题1"})
	round1 = append(round1, makeRoundMsgs("r1", schema.Assistant, []string{"回答1"})...)
	round2 := makeRoundMsgs("r2", schema.User, []string{"问题2"})
	round2 = append(round2, makeRoundMsgs("r2", schema.Assistant, []string{"回答2"})...)

	var history []*schema.Message
	history = append(history, round1...)
	history = append(history, round2...)

	result, err := c.DoCompress(context.Background(), nil, history)
	if err != nil {
		t.Fatalf("DoCompress error: %v", err)
	}

	fmt.Println("\n=== TestCompress_NotEnoughRounds ===")
	fmt.Println("轮次数(2) <= keepRounds(3)，不应压缩")
	fmt.Printf("result.Compressed = %v\n", c.result.Compressed)

	if c.result.Compressed {
		t.Fatal("should not be compressed when rounds <= keepRounds")
	}
	if len(repo.markedRoundIds) > 0 {
		t.Fatal("no rounds should be marked as compressed")
	}
	if len(result) != len(history) {
		t.Fatalf("history should be unchanged, expected %d got %d", len(history), len(result))
	}
}

// go test -test.fullpath=true -timeout 30s -run ^TestCompress_OldSummaryOnlySkip$ goraven/core/agent -v -count=1
func TestCompress_OldSummaryOnlySkip(t *testing.T) {
	repo := &mockMsgRepo{}
	chatModel := &mockChatModel{response: "摘要"}
	c := &Compress{
		chatModel:  chatModel,
		threshold:  1,
		msgRepo:    repo,
		sessionId:  "test-session",
		keepRounds: 3,
	}

	// 只有旧摘要 + 2轮对话，keepRounds=3，不应压缩（2轮 <= 3）
	oldSummary := makeSummaryMsg("旧摘要内容")
	round1 := makeRoundMsgs("r1", schema.User, []string{"问题1"})
	round1 = append(round1, makeRoundMsgs("r1", schema.Assistant, []string{"回答1"})...)
	round2 := makeRoundMsgs("r2", schema.User, []string{"问题2"})
	round2 = append(round2, makeRoundMsgs("r2", schema.Assistant, []string{"回答2"})...)

	var history []*schema.Message
	history = append(history, oldSummary)
	history = append(history, round1...)
	history = append(history, round2...)

	result, err := c.DoCompress(context.Background(), nil, history)
	if err != nil {
		t.Fatalf("DoCompress error: %v", err)
	}

	fmt.Println("\n=== TestCompress_OldSummaryOnlySkip ===")
	fmt.Println("旧摘要+2轮对话, keepRounds=3，不应压缩")
	fmt.Printf("result.Compressed = %v\n", c.result.Compressed)
	printHistory(result)

	if c.result.Compressed {
		t.Fatal("should not compress when rounds <= keepRounds")
	}
}

// go test -test.fullpath=true -timeout 30s -run ^TestCompress_MultipleCompress$ goraven/core/agent -v -count=1
func TestCompress_MultipleCompress(t *testing.T) {
	repo := &mockMsgRepo{}
	chatModel := &mockChatModel{response: "第二次压缩摘要"}
	c := &Compress{
		chatModel:  chatModel,
		threshold:  1,
		msgRepo:    repo,
		sessionId:  "test-session",
		keepRounds: 2,
	}

	// 模拟已经经过一次压缩的场景：[旧摘要, round2, round3, round4, round5]
	oldSummary := makeSummaryMsg("第一次的摘要内容")
	round2 := makeRoundMsgs("r2", schema.User, []string{"问题2"})
	round2 = append(round2, makeRoundMsgs("r2", schema.Assistant, []string{"回答2"})...)
	round3 := makeRoundMsgs("r3", schema.User, []string{"问题3"})
	round3 = append(round3, makeRoundMsgs("r3", schema.Assistant, []string{"回答3"})...)
	round4 := makeRoundMsgs("r4", schema.User, []string{"问题4"})
	round4 = append(round4, makeRoundMsgs("r4", schema.Assistant, []string{"回答4"})...)
	round5 := makeRoundMsgs("r5", schema.User, []string{"问题5"})
	round5 = append(round5, makeRoundMsgs("r5", schema.Assistant, []string{"回答5"})...)

	var history []*schema.Message
	history = append(history, oldSummary)
	history = append(history, round2...)
	history = append(history, round3...)
	history = append(history, round4...)
	history = append(history, round5...)

	result, err := c.DoCompress(context.Background(), nil, history)
	if err != nil {
		t.Fatalf("DoCompress error: %v", err)
	}

	fmt.Println("\n=== TestCompress_MultipleCompress ===")
	fmt.Println("压缩前:")
	printHistory(history)
	fmt.Println("压缩后:")
	printHistory(result)

	if !c.result.Compressed {
		t.Fatal("should be compressed")
	}

	// 应包含: 旧摘要 + 新摘要 + round4 + round5
	// round2, round3 被压缩
	expectedLen := 1 + 1 + len(round4) + len(round5)
	if len(result) != expectedLen {
		t.Fatalf("expected %d messages, got %d", expectedLen, len(result))
	}

	// 第一条应为旧摘要
	isSummary, _ := result[0].Extra["isSummary"].(bool)
	if !isSummary {
		t.Fatal("first message should be old summary")
	}

	// 第二条应为新摘要
	isSummary2, _ := result[1].Extra["isSummary"].(bool)
	if !isSummary2 {
		t.Fatal("second message should be new summary")
	}

	// r2, r3 应被标记为已压缩
	if len(repo.markedRoundIds) != 2 {
		t.Fatalf("expected 2 marked roundIds, got %d: %v", len(repo.markedRoundIds), repo.markedRoundIds)
	}
}

// go test -test.fullpath=true -timeout 30s -run ^TestCompress_KeepOneRound$ goraven/core/agent -v -count=1
func TestCompress_KeepOneRound(t *testing.T) {
	repo := &mockMsgRepo{}
	chatModel := &mockChatModel{response: "上海天气的摘要"}
	c := &Compress{
		chatModel:  chatModel,
		threshold:  1,
		msgRepo:    repo,
		sessionId:  "51314",
		keepRounds: 1,
	}

	// 模拟: 3轮上海天气 + 1轮北京天气, keepRounds=1
	// 压缩前: [上海1, 上海2, 上海3, 北京]
	// 压缩后: [摘要, 北京]
	round1 := makeRoundMsgs("r1", schema.User, []string{"上海天气如何？"})
	round1 = append(round1, makeRoundMsgs("r1", schema.Assistant, []string{"上海晴天28度"})...)
	round2 := makeRoundMsgs("r2", schema.User, []string{"上海天气如何？"})
	round2 = append(round2, makeRoundMsgs("r2", schema.Assistant, []string{"上海晴天28度"})...)
	round3 := makeRoundMsgs("r3", schema.User, []string{"上海天气如何？"})
	round3 = append(round3, makeRoundMsgs("r3", schema.Assistant, []string{"上海晴天28度"})...)
	round4 := makeRoundMsgs("r4", schema.User, []string{"北京天气如何？"})
	round4 = append(round4, makeRoundMsgs("r4", schema.Assistant, []string{"北京晴天28度"})...)

	var history []*schema.Message
	history = append(history, round1...)
	history = append(history, round2...)
	history = append(history, round3...)
	history = append(history, round4...)

	result, err := c.DoCompress(context.Background(), nil, history)
	if err != nil {
		t.Fatalf("DoCompress error: %v", err)
	}

	fmt.Println("\n=== TestCompress_KeepOneRound ===")
	fmt.Println("压缩前:")
	printHistory(history)
	fmt.Println("压缩后:")
	printHistory(result)

	if !c.result.Compressed {
		t.Fatal("should be compressed")
	}

	// 应包含: 新摘要 + round4 = 3条消息
	expectedLen := 1 + len(round4)
	if len(result) != expectedLen {
		t.Fatalf("expected %d messages, got %d", expectedLen, len(result))
	}

	// 第一条应为摘要，第二条开始是北京天气
	if result[0].Content != "[对话摘要]\n上海天气的摘要" {
		t.Fatalf("first should be new summary, got: %s", result[0].Content)
	}
	if result[1].Content != "北京天气如何？" {
		t.Fatalf("second should be 北京 question, got: %s", result[1].Content)
	}

	// r1, r2, r3 都应被标记为已压缩
	if len(repo.markedRoundIds) != 3 {
		t.Fatalf("expected 3 marked roundIds (r1,r2,r3), got %d: %v", len(repo.markedRoundIds), repo.markedRoundIds)
	}
	expectedIds := map[string]bool{"r1": true, "r2": true, "r3": true}
	for _, rid := range repo.markedRoundIds {
		if !expectedIds[rid] {
			t.Fatalf("unexpected marked roundId: %s", rid)
		}
	}

	// 摘要的时间戳应在被压缩消息之后，保留消息之前
	savedSummary := repo.savedMessages[0]
	fmt.Printf("摘要保存时间戳: %d\n", savedSummary.Timestamp)
	if savedSummary.RoleType != po.RoleTypeSummary {
		t.Fatalf("saved message should be summary type, got: %s", savedSummary.RoleType)
	}
}
