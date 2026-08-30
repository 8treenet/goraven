package agent

import (
	"context"
	"errors"
	"fmt"
	"goraven/core/iface"
	"strings"

	"github.com/8treenet/freedom"
	"github.com/cloudwego/eino/schema"
)

const (
	sessionTitleMaxRunes      = 30
	sessionTitleInputMaxRunes = 2000
)

// SessionTitler 生成会话标题的独立组件。
// 不依赖 MainAgent；token 用量会按聊天主流程口径累加到 session 与 daily_stats。
type SessionTitler struct {
	chatModel      iface.BaseChatModel
	msgRepo        iface.MessageRepo
	dailyStatsRepo iface.DailyStatsRepo
	sessionId      string
	userId         string
}

// NewSessionTitler 构造 SessionTitler。
// chatModel 通常复用聊天侧的压缩模型；msgRepo / dailyStatsRepo 可为 nil（仅生成不写统计）。
func NewSessionTitler(
	chatModel iface.BaseChatModel,
	msgRepo iface.MessageRepo,
	dailyStatsRepo iface.DailyStatsRepo,
	sessionId, userId string,
) *SessionTitler {
	return &SessionTitler{
		chatModel:      chatModel,
		msgRepo:        msgRepo,
		dailyStatsRepo: dailyStatsRepo,
		sessionId:      sessionId,
		userId:         userId,
	}
}

// Generate 根据首轮 user/assistant 内容生成短标题；不写库标题，写库由调用方完成。
// 同步累加 token 统计到 session 与 user_daily_stats。
func (t *SessionTitler) Generate(ctx context.Context, userContent, assistantReply string) (string, error) {
	if t == nil || t.chatModel == nil {
		return "", errors.New("session titler model is nil")
	}
	if strings.TrimSpace(userContent) == "" && strings.TrimSpace(assistantReply) == "" {
		return "", errors.New("empty conversation for title")
	}

	i18n := getSessionTitleI18n()
	user := truncateRunesForSessionTitle(userContent, sessionTitleInputMaxRunes)
	reply := truncateRunesForSessionTitle(assistantReply, sessionTitleInputMaxRunes)

	resp, err := t.chatModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage(i18n.systemPrompt),
		schema.UserMessage(fmt.Sprintf(i18n.userPrompt, user, reply)),
	})
	if err != nil {
		return "", err
	}

	title := sanitizeSessionTitle(resp.Content)
	if title == "" {
		return "", errors.New("empty title generated")
	}

	t.recordTokenUsage(resp)
	freedom.Logger().Debugf("session title generated: sessionId=%s title=%s", t.sessionId, title)
	return title, nil
}

func (t *SessionTitler) recordTokenUsage(resp *schema.Message) {
	if resp == nil || resp.ResponseMeta == nil || resp.ResponseMeta.Usage == nil {
		return
	}
	usage := resp.ResponseMeta.Usage
	prompt := usage.PromptTokens
	completion := usage.CompletionTokens
	cached := usage.PromptTokenDetails.CachedTokens
	if prompt == 0 && completion == 0 {
		return
	}
	if t.msgRepo != nil && t.sessionId != "" {
		if err := t.msgRepo.AddSessionTokens(t.sessionId, prompt, completion, cached); err != nil {
			freedom.Logger().Warnf("session title AddSessionTokens sessionId=%s err=%v", t.sessionId, err)
		}
	}
	if t.dailyStatsRepo != nil && t.userId != "" {
		if err := t.dailyStatsRepo.AddDailyStats(t.userId, prompt, completion, cached); err != nil {
			freedom.Logger().Warnf("session title AddDailyStats userId=%s err=%v", t.userId, err)
		}
	}
}

func truncateRunesForSessionTitle(s string, n int) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n])
}

func sanitizeSessionTitle(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexAny(s, "\r\n"); idx >= 0 {
		s = s[:idx]
	}
	s = strings.Trim(s, "\"'`“”‘’《》〈〉「」 \t")
	s = strings.TrimSpace(s)
	rs := []rune(s)
	if len(rs) > sessionTitleMaxRunes {
		s = string(rs[:sessionTitleMaxRunes])
	}
	return s
}

type sessionTitleI18n struct {
	systemPrompt string
	userPrompt   string
}

func getSessionTitleI18n() sessionTitleI18n {
	if isChineseLanguage() {
		return sessionTitleI18n{
			systemPrompt: `你是一个会话标题生成助手。根据用户与助手的首次对话内容，生成一个简短的会话标题。
要求：
1. 不超过 20 个汉字（或等长的英文短语）
2. 只输出标题本身，不要加任何前缀、引号、标点、解释或换行
3. 概括对话主题，名词性短语为佳`,
			userPrompt: `用户消息：
%s

助手回复：
%s

标题：`,
		}
	}
	return sessionTitleI18n{
		systemPrompt: `You generate concise session titles. Based on the first user-assistant exchange, output a single short title.
Rules:
1. No more than 40 characters
2. Output only the title itself; no prefix, quotes, punctuation, explanation, or newline
3. Summarize the topic as a noun phrase`,
		userPrompt: `User message:
%s

Assistant reply:
%s

Title:`,
	}
}
