package pruning

import (
	"context"
	"fmt"
	"goraven/backend/repository"
	"strings"
	"unicode/utf8"

	"github.com/8treenet/freedom"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

const (
	msgPrunedFlag = "_pruning_mw_processed"
)

// Config is the configuration for pruning middleware.
type Config struct {
	// TokenThreshold is the threshold for total message tokens.
	// When total tokens exceed this threshold, pruning starts.
	// Default: 64000
	TokenThreshold int64

	// TokenCounter is used to count tokens in messages.
	// If nil, uses default token counter (character count / 4).
	TokenCounter func(ctx context.Context, msgs []adk.Message, tools []*schema.ToolInfo) (int64, error)

	// MaxToolResultLength is the maximum length for tool result content.
	// Tool results exceeding this length will be truncated (head + tail).
	// Default: 2000
	MaxToolResultLength int

	// HeadLength is the number of characters to keep from the head when truncating.
	// Default: 1000
	HeadLength int

	// TailLength is the number of characters to keep from the tail when truncating.
	// Default: 1000
	TailLength int
}

type pruningMiddleware struct {
	adk.BaseChatModelAgentMiddleware

	config *Config
}

// New creates a new pruning middleware.
// sysCfg provides default values from system settings; conf overrides if provided.
func New(sysCfg *repository.SystemConfig, conf ...*Config) (adk.ChatModelAgentMiddleware, error) {
	cfg := &Config{
		TokenThreshold:      int64(sysCfg.PruningTokenThreshold * 1024),
		TokenCounter:        defaultTokenCounter,
		MaxToolResultLength: sysCfg.PruningMaxToolResultLength,
		HeadLength:          sysCfg.PruningHeadTruncateLength,
		TailLength:          sysCfg.PruningTailTruncateLength,
	}
	if cfg.TokenThreshold < 50*1024 {
		cfg.TokenThreshold = 50 * 1024
	}
	if len(conf) > 0 {
		cfg = conf[0]
	}
	return &pruningMiddleware{config: cfg}, nil
}

func (p *pruningMiddleware) BeforeModelRewriteState(ctx context.Context, state *adk.ChatModelAgentState, mc *adk.ModelContext) (
	context.Context, *adk.ChatModelAgentState, error) {

	originalMsgCount := len(state.Messages)
	rounds := extractToolRounds(state.Messages)

	if len(rounds) == 0 {
		freedom.Logger().Debugf("[PRUNING] 消息数=%d, 无工具轮次, 跳过\n", originalMsgCount)
		return ctx, state, nil
	}

	roundMsgIdx := make(map[int]bool, len(rounds)*2)
	for _, r := range rounds {
		roundMsgIdx[r.assistantIdx] = true
		for _, ti := range r.toolIndices {
			roundMsgIdx[ti] = true
		}
	}
	nonRoundMsgs := make([]adk.Message, 0, len(state.Messages))
	for i, m := range state.Messages {
		if !roundMsgIdx[i] {
			nonRoundMsgs = append(nonRoundMsgs, m)
		}
	}
	baseTokens, err := p.config.TokenCounter(ctx, nonRoundMsgs, mc.Tools)
	if err != nil {
		return ctx, state, err
	}

	originalRoundCount := len(rounds)
	roundTokens := int64(0)
	for _, r := range rounds {
		roundTokens += estimateRoundTokens(r)
	}

	prunedRounds := p.pruneRounds(rounds, baseTokens)

	state.Messages = reconstructMessages(state.Messages, prunedRounds)
	removedRounds := originalRoundCount - len(prunedRounds)

	freedom.Logger().Debugf("[PRUNING] 消息: %d->%d | 工具轮次: %d->%d | Token: %d(基数%d+轮次%d) | 阈值: %d | 删除轮次: %d\n",
		originalMsgCount, len(state.Messages),
		originalRoundCount, len(prunedRounds),
		baseTokens+roundTokens, baseTokens, roundTokens,
		p.config.TokenThreshold,
		removedRounds)

	return ctx, state, nil
}

// toolRound 代表一轮工具调用：一个 assistant 消息（含 N 个 tool call）
// 加上它的所有 tool 响应。以 round 为最小剪枝单位可保证 toolCall 与 tool 响应的配对完整性。
type toolRound struct {
	assistantMsg *schema.Message
	assistantIdx int
	// toolMsgs 与 toolIndices 按 assistantMsg.ToolCalls 的顺序依次对应。
	// 若某个 toolCall 在后续消息里找不到响应，则该位置被跳过（长度可能 < len(ToolCalls)）。
	toolMsgs    []*schema.Message
	toolIndices []int
	timestamp   int64
}

// extractToolRounds 把消息序列按 assistant 消息聚合成 round。
// 每个含 ToolCalls 的 assistant 消息产生一个 round，之后的 tool 消息按 ToolCallID 匹配归入。
func extractToolRounds(messages []adk.Message) []*toolRound {
	var rounds []*toolRound

	for i := 0; i < len(messages); i++ {
		msg := messages[i]
		if msg == nil || msg.Role != schema.Assistant || len(msg.ToolCalls) == 0 {
			continue
		}

		round := &toolRound{
			assistantMsg: msg,
			assistantIdx: i,
			timestamp:    getTimestamp(msg),
		}

		for _, tc := range msg.ToolCalls {
			for j := i + 1; j < len(messages); j++ {
				toolMsg := messages[j]
				if toolMsg == nil {
					continue
				}
				// 下一个 assistant with toolCalls 之前的 tool 消息都属于当前 round
				if toolMsg.Role == schema.Assistant && len(toolMsg.ToolCalls) > 0 {
					break
				}
				if toolMsg.Role == schema.Tool && toolMsg.ToolCallID == tc.ID {
					round.toolMsgs = append(round.toolMsgs, toolMsg)
					round.toolIndices = append(round.toolIndices, j)
					break
				}
			}
		}

		rounds = append(rounds, round)
	}

	return rounds
}

func (p *pruningMiddleware) pruneRounds(rounds []*toolRound, baseTokens int64) []*toolRound {
	if len(rounds) == 0 {
		return rounds
	}

	var totalTokens int64
	for _, r := range rounds {
		totalTokens += estimateRoundTokens(r)
	}

	if baseTokens+totalTokens <= p.config.TokenThreshold {
		return rounds
	}

	// 从旧到新逐轮处理，始终保留至少1轮。
	for i := 0; i < len(rounds)-1; i++ {
		r := rounds[i]
		if r == nil {
			continue
		}

		oldTokens := estimateRoundTokens(r)

		// ① 截断超长工具结果
		for _, toolMsg := range r.toolMsgs {
			if utf8.RuneCountInString(toolMsg.Content) > p.config.MaxToolResultLength {
				toolMsg.Content = truncateContent(
					toolMsg.Content,
					p.config.HeadLength,
					p.config.TailLength,
				)
			}
		}

		newTokens := estimateRoundTokens(r)
		totalTokens = totalTokens - oldTokens + newTokens
		if baseTokens+totalTokens <= p.config.TokenThreshold {
			break
		}

		// ② 仍超阈值 → 丢弃整个轮次
		totalTokens -= newTokens
		rounds[i] = nil
		if baseTokens+totalTokens <= p.config.TokenThreshold {
			break
		}
	}

	final := make([]*toolRound, 0, len(rounds))
	for _, r := range rounds {
		if r != nil {
			final = append(final, r)
		}
	}

	return final
}

func truncateContent(content string, headLen, tailLen int) string {
	runes := []rune(content)
	if len(runes) <= headLen+tailLen {
		return content
	}

	head := string(runes[:headLen])
	tail := string(runes[len(runes)-tailLen:])
	middleLen := len(runes) - headLen - tailLen

	return fmt.Sprintf("%s\n\n[System: 因超出上下文限制，中间 %d 个字符已被剪枝隐藏]\n\n... %s", head, middleLen, tail)
}

// reconstructMessages 按原顺序重建消息：
//   - 保留所有非 tool 且非（含 toolCalls 的 assistant）消息；
//   - 对含 toolCalls 的 assistant 和 tool 消息，只保留出现在 rounds 里的索引。
//
// 以 round 为整体保留/剔除，保证 toolCall 与 tool 响应成对。
func reconstructMessages(original []adk.Message, rounds []*toolRound) []adk.Message {
	keepIndices := make(map[int]bool, len(rounds)*2)
	for _, r := range rounds {
		keepIndices[r.assistantIdx] = true
		for _, ti := range r.toolIndices {
			keepIndices[ti] = true
		}
	}

	result := make([]adk.Message, 0, len(original))
	for i, msg := range original {
		if msg == nil {
			continue
		}
		isToolMsg := msg.Role == schema.Tool
		isAsstWithCalls := msg.Role == schema.Assistant && len(msg.ToolCalls) > 0
		if keepIndices[i] || (!isToolMsg && !isAsstWithCalls) {
			result = append(result, msg)
		}
	}

	return result
}

func getTimestamp(msg *schema.Message) int64 {
	if msg.Extra == nil {
		return 0
	}
	if ts, ok := msg.Extra["timestamp"].(int64); ok {
		return ts
	}
	return 0
}

func getMsgPrunedFlag(msg *schema.Message) bool {
	if msg.Extra == nil {
		return false
	}
	v, ok := msg.Extra[msgPrunedFlag].(bool)
	return ok && v
}

func setMsgPrunedFlag(msg *schema.Message) {
	if msg.Extra == nil {
		msg.Extra = make(map[string]any)
	}
	msg.Extra[msgPrunedFlag] = true
}

// estimateStrTokens 估算一段文本的 token 数，对中英文混合内容做自适应。
// 英文等 ASCII 主导文本沿用 bytes/4 的经验公式，中文等多字节主导文本按 rune*1.5 估算。
func estimateStrTokens(s string) int64 {
	byteLen := len(s)
	runeLen := utf8.RuneCountInString(s)
	if byteLen > runeLen*2 {
		return int64(runeLen * 3 / 2)
	}
	return int64((byteLen + 3) / 4)
}

func estimateRoundTokens(r *toolRound) int64 {
	var tokens int64
	tokens += estimateStrTokens(r.assistantMsg.Content)
	tokens += estimateStrTokens(r.assistantMsg.ReasoningContent)
	for _, tc := range r.assistantMsg.ToolCalls {
		tokens += estimateStrTokens(tc.Function.Name)
		tokens += estimateStrTokens(tc.Function.Arguments)
	}
	for _, tm := range r.toolMsgs {
		tokens += estimateStrTokens(tm.Content)
	}
	return tokens
}

func defaultTokenCounter(_ context.Context, msgs []adk.Message, _ []*schema.ToolInfo) (int64, error) {
	var tokens int64
	for _, msg := range msgs {
		if msg == nil {
			continue
		}

		var sb strings.Builder
		sb.WriteString(string(msg.Role))
		sb.WriteString("\n")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
		if msg.ReasoningContent != "" {
			sb.WriteString(msg.ReasoningContent)
			sb.WriteString("\n")
		}

		if msg.Role == schema.Assistant && len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				sb.WriteString(tc.Function.Name)
				sb.WriteString("\n")
				sb.WriteString(tc.Function.Arguments)
			}
		}

		tokens += estimateStrTokens(sb.String())
	}
	return tokens, nil
}
