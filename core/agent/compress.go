package agent

import (
	"context"
	"fmt"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/config"
	"goraven/core/iface"
	"goraven/util"
	"strings"
	"unicode/utf8"

	"github.com/8treenet/freedom"
	"github.com/cloudwego/eino/schema"
)

type Compress struct {
	chatModel      iface.BaseChatModel
	threshold      int
	msgRepo        iface.MessageRepo
	dailyStatsRepo iface.DailyStatsRepo
	sessionId      string
	userId         string
	keepRounds     int
	result         *CompressResult
}

type CompressResult struct {
	Compressed         bool
	History            []*schema.Message
	PromptTokens       int
	CompletionTokens   int
	PromptCachedTokens int
	SummaryContent     string
}

func NewCompress(chatModel iface.BaseChatModel, msgRepo iface.MessageRepo, dailyStatsRepo iface.DailyStatsRepo, sessionId, userId string, sysCfg *repository.SystemConfig) *Compress {
	percent := sysCfg.CompressThresholdPercent
	if percent <= 50 {
		percent = 50
	}

	contextLen := chatModel.ContextLength()
	if contextLen <= 128*1024 {
		contextLen = 128 * 1024
	}
	threshold := contextLen * percent / 100
	freedom.Logger().Debugf("compress threshold:%v", threshold)

	keepRounds := sysCfg.CompressKeepRounds
	if keepRounds <= 0 {
		keepRounds = 3
	}

	return &Compress{
		chatModel:      chatModel,
		threshold:      threshold,
		msgRepo:        msgRepo,
		dailyStatsRepo: dailyStatsRepo,
		sessionId:      sessionId,
		userId:         userId,
		keepRounds:     keepRounds,
	}
}

type roundGroup struct {
	roundId   string
	messages  []*schema.Message
	isSummary bool
	maxTs     int64
}

func (c *Compress) groupByRounds(history []*schema.Message) []*roundGroup {
	roundMap := make(map[string]*roundGroup)
	var roundOrder []string
	var noRoundGroup *roundGroup

	for _, msg := range history {
		if msg.Role == schema.System {
			continue
		}

		roundId, _ := msg.Extra["roundId"].(string)
		ts, _ := msg.Extra["timestamp"].(int64)
		if roundId == "" {
			if noRoundGroup == nil {
				noRoundGroup = &roundGroup{roundId: "", isSummary: true}
			}
			noRoundGroup.messages = append(noRoundGroup.messages, msg)
			if ts > noRoundGroup.maxTs {
				noRoundGroup.maxTs = ts
			}
			continue
		}

		if _, ok := roundMap[roundId]; !ok {
			roundMap[roundId] = &roundGroup{roundId: roundId}
			roundOrder = append(roundOrder, roundId)
		}
		rg := roundMap[roundId]
		rg.messages = append(rg.messages, msg)
		if ts > rg.maxTs {
			rg.maxTs = ts
		}
	}

	var rounds []*roundGroup
	if noRoundGroup != nil && len(noRoundGroup.messages) > 0 {
		rounds = append(rounds, noRoundGroup)
	}
	for _, rid := range roundOrder {
		rounds = append(rounds, roundMap[rid])
	}
	return rounds
}

func (c *Compress) estimateTokens(messages []*schema.Message) int {
	total := 0
	for _, msg := range messages {
		text := msg.Content
		text += msg.ReasoningContent
		for _, part := range msg.UserInputMultiContent {
			if part.Type == schema.ChatMessagePartTypeText {
				text += part.Text
			}
		}
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				text += tc.Function.Name + tc.Function.Arguments
			}
		}
		if msg.ToolName != "" {
			text += msg.ToolName
		}

		byteLen := len(text)
		runeLen := utf8.RuneCountInString(text)
		// 多字节字符（中文等，约 3 字节/字，~1.5 token/字）与 ASCII（1 字节/字，~0.25 token/字）
		// 密度差异大，不能共用单一系数：字节主导沿用 bytes/4，多字节主导按 rune*1.5 估算
		if byteLen > runeLen*2 {
			total += runeLen * 3 / 2
		} else {
			total += (byteLen + 3) / 4
		}
	}
	return total
}

func (c *Compress) ShouldCompress(history []*schema.Message) bool {
	return c.estimateTokens(history) > c.threshold
}

func (c *Compress) DoCompress(ctx context.Context, runner *MainRunner, history []*schema.Message) ([]*schema.Message, error) {
	if !c.ShouldCompress(history) {
		c.result = &CompressResult{Compressed: false, History: history}
		return history, nil
	}

	rounds := c.groupByRounds(history)

	roundCount := len(rounds)
	if rounds[0].isSummary {
		roundCount--
	}
	if roundCount <= c.keepRounds {
		c.result = &CompressResult{Compressed: false, History: history}
		return history, nil
	}

	compressCount := roundCount - c.keepRounds
	if runner != nil {
		runner.sendSSEEvent(&SSEEvent{
			Type:    SSEEventTypeReasoning,
			Content: getCompressI18n().reasoningMsg,
		})
	}

	startIdx := 0
	if rounds[0].isSummary {
		startIdx = 1
	}

	var compressMessages []*schema.Message
	var compressedRoundIds []string
	var lastCompressedMsgTs int64
	for i := startIdx; i < startIdx+compressCount; i++ {
		if rounds[i].roundId != "" {
			compressedRoundIds = append(compressedRoundIds, rounds[i].roundId)
		}
		compressMessages = append(compressMessages, rounds[i].messages...)
		if rounds[i].maxTs > lastCompressedMsgTs {
			lastCompressedMsgTs = rounds[i].maxTs
		}
	}

	if len(compressMessages) == 0 {
		c.result = &CompressResult{Compressed: false, History: history}
		return history, nil
	}

	compressText := c.buildCompressText(compressMessages)

	resp, err := c.chatModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage(getCompressSystemPrompt()),
		schema.UserMessage(compressText),
	})
	if err != nil {
		return nil, err
	}

	c.result = &CompressResult{
		Compressed:     true,
		SummaryContent: resp.Content,
	}
	if resp.ResponseMeta != nil && resp.ResponseMeta.Usage != nil {
		c.result.PromptTokens = resp.ResponseMeta.Usage.PromptTokens
		c.result.CompletionTokens = resp.ResponseMeta.Usage.CompletionTokens
		c.result.PromptCachedTokens = resp.ResponseMeta.Usage.PromptTokenDetails.CachedTokens
	}

	var keptMessages []*schema.Message
	for i := startIdx + compressCount; i < len(rounds); i++ {
		keptMessages = append(keptMessages, rounds[i].messages...)
	}

	newSummaryMsg := schema.UserMessage(getCompressI18n().summaryPrefix + resp.Content)
	newSummaryMsg.Extra = map[string]any{"isSummary": true}

	var newHistory []*schema.Message
	if rounds[0].isSummary {
		newHistory = rounds[0].messages
	}
	newHistory = append(newHistory, newSummaryMsg)
	newHistory = append(newHistory, keptMessages...)
	c.result.History = newHistory

	freedom.Logger().Debugf("Compress: %d rounds (%d messages) -> 1 summary + %d rounds (%d messages kept), keepRounds=%d, promptTokens=%d, completionTokens=%d",
		compressCount, len(compressMessages), c.keepRounds, len(keptMessages), c.keepRounds, c.result.PromptTokens, c.result.CompletionTokens)

	c.saveSummary(compressedRoundIds, lastCompressedMsgTs)
	return newHistory, nil
}

// ForceCompress 手动压缩全部历史消息，无任何条件限制
// 将所有非系统消息压缩为一份摘要，返回仅含摘要的新历史
func (c *Compress) ForceCompress(ctx context.Context, history []*schema.Message) ([]*schema.Message, error) {
	var allMessages []*schema.Message
	var allRoundIds []string
	var lastCompressedMsgTs int64

	for _, msg := range history {
		if msg.Role == schema.System {
			continue
		}
		if isSummary, _ := msg.Extra["isSummary"].(bool); isSummary {
			continue
		}
		allMessages = append(allMessages, msg)
		roundId, _ := msg.Extra["roundId"].(string)
		if roundId != "" {
			found := false
			for _, rid := range allRoundIds {
				if rid == roundId {
					found = true
					break
				}
			}
			if !found {
				allRoundIds = append(allRoundIds, roundId)
			}
		}
		ts, _ := msg.Extra["timestamp"].(int64)
		if ts > lastCompressedMsgTs {
			lastCompressedMsgTs = ts
		}
	}

	if len(allMessages) == 0 {
		c.result = &CompressResult{Compressed: false, History: history}
		return history, nil
	}

	compressText := c.buildCompressText(allMessages)
	resp, err := c.chatModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage(getCompressSystemPrompt()),
		schema.UserMessage(compressText),
	})
	if err != nil {
		return nil, err
	}

	c.result = &CompressResult{
		Compressed:     true,
		SummaryContent: resp.Content,
	}
	if resp.ResponseMeta != nil && resp.ResponseMeta.Usage != nil {
		c.result.PromptTokens = resp.ResponseMeta.Usage.PromptTokens
		c.result.CompletionTokens = resp.ResponseMeta.Usage.CompletionTokens
		c.result.PromptCachedTokens = resp.ResponseMeta.Usage.PromptTokenDetails.CachedTokens
	}

	newSummaryMsg := schema.UserMessage(getCompressI18n().summaryPrefix + resp.Content)
	newSummaryMsg.Extra = map[string]any{"isSummary": true}
	newHistory := []*schema.Message{newSummaryMsg}
	c.result.History = newHistory

	freedom.Logger().Debugf("ForceCompress: %d messages -> 1 summary, promptTokens=%d, completionTokens=%d",
		len(allMessages), c.result.PromptTokens, c.result.CompletionTokens)

	c.saveSummary(allRoundIds, lastCompressedMsgTs)
	return newHistory, nil
}

// buildCompressText 将消息列表格式化为待压缩的文本
func (c *Compress) buildCompressText(messages []*schema.Message) string {
	var b strings.Builder
	i18n := getCompressI18n()
	b.WriteString(i18n.historyPrompt)
	for _, msg := range messages {
		switch msg.Role {
		case schema.User:
			b.WriteString(fmt.Sprintf(i18n.userLabel+": %s\n\n", msg.Content))
		case schema.Assistant:
			if len(msg.ToolCalls) > 0 {
				var toolInfos []string
				for _, tc := range msg.ToolCalls {
					toolInfos = append(toolInfos, fmt.Sprintf("%s(%s)", tc.Function.Name, tc.Function.Arguments))
				}
				if msg.Content != "" {
					b.WriteString(fmt.Sprintf(i18n.assistantLabel+": %s\n", msg.Content))
				}
				b.WriteString(fmt.Sprintf(i18n.toolCallLabel+": %s\n\n", strings.Join(toolInfos, ", ")))
			} else {
				b.WriteString(fmt.Sprintf(i18n.assistantLabel+": %s\n\n", msg.Content))
			}
		case schema.Tool:
			toolName := msg.ToolName
			if toolName == "" {
				toolName = i18n.defaultTool
			}
			b.WriteString(fmt.Sprintf(i18n.toolResultFmt+": %s\n\n", toolName, msg.Content))
		}
	}
	return b.String()
}

func (c *Compress) saveSummary(roundIds []string, lastCompressedMsgTs int64) {
	if c.msgRepo == nil || !c.result.Compressed {
		return
	}

	c.msgRepo.MarkSessionCompressed(c.sessionId, roundIds)

	summaryTs := lastCompressedMsgTs + 1
	if summaryTs <= 0 {
		summaryTs = util.Millisecond()
	}

	summaryMsg := &po.Message{
		SessionId:               c.sessionId,
		Timestamp:               summaryTs,
		Content:                 c.result.SummaryContent,
		RoleType:                po.RoleTypeSummary,
		PromptTokensCount:       c.result.PromptTokens,
		CompletionTokensCount:   c.result.CompletionTokens,
		PromptCachedTokensCount: c.result.PromptCachedTokens,
	}
	c.msgRepo.SaveChatMessage(c.sessionId, summaryMsg)
	c.msgRepo.AddSessionTokens(c.sessionId, c.result.PromptTokens, c.result.CompletionTokens, c.result.PromptCachedTokens)
	c.msgRepo.SetContextTokens(c.sessionId, 0)
	if c.dailyStatsRepo != nil {
		c.dailyStatsRepo.AddDailyStats(c.userId, c.result.PromptTokens, c.result.CompletionTokens, c.result.PromptCachedTokens)
	}
}

func (c *Compress) GetResult() *CompressResult {
	return c.result
}

// ════════════════════════════════════════════════════════════════════════════
// 压缩相关 i18n 字符串
// ════════════════════════════════════════════════════════════════════════════

type compressI18n struct {
	historyPrompt  string
	userLabel      string
	assistantLabel string
	toolCallLabel  string
	defaultTool    string
	toolResultFmt  string
	summaryPrefix  string
	reasoningMsg   string
}

func getCompressI18n() compressI18n {
	if config.Get().GetLanguage() == "zh" {
		return compressI18n{
			historyPrompt:  "以下是需要压缩的对话历史：\n\n",
			userLabel:      "[用户]",
			assistantLabel: "[助手]",
			toolCallLabel:  "[助手调用工具]",
			defaultTool:    "工具",
			toolResultFmt:  "[%s返回]",
			summaryPrefix:  "[对话摘要]\n",
			reasoningMsg:   "正在整理对话历史...\n\n",
		}
	}
	return compressI18n{
		historyPrompt:  "Below is the conversation history to compress:\n\n",
		userLabel:      "[User]",
		assistantLabel: "[Assistant]",
		toolCallLabel:  "[Assistant tool call]",
		defaultTool:    "Tool",
		toolResultFmt:  "[%s result]",
		summaryPrefix:  "[Conversation summary]\n",
		reasoningMsg:   "Organizing conversation history...\n\n",
	}
}
