package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"goraven/config"
	"goraven/util"

	"github.com/8treenet/freedom"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// SystemResult 系统代理执行结果
type SystemResult struct {
	Status  string   `json:"status"`
	Summary string   `json:"summary"`
	Details []string `json:"details"`
}

func newSystemRunner(sysAgent *SystemAgent, runnerAgent *adk.ChatModelAgent) *SystemRunner {
	runner := adk.NewRunner(context.Background(), adk.RunnerConfig{
		Agent:           runnerAgent,
		EnableStreaming: false,
	})

	return &SystemRunner{
		sysAgent: sysAgent,
		runner:   runner,
	}
}

type SystemRunner struct {
	sysAgent *SystemAgent
	runner   *adk.Runner
}

func (sr *SystemRunner) Run(ctx context.Context, query string) (result *SystemResult, err error) {
	// 一次性运行：使用系统代理模型配置的 header 名与新的运行 ID，安装流程内多轮工具调用共享
	ctx = util.WithConversationHeader(ctx, sr.sysAgent.chatModel.GetConversationHeaderKey(), util.UUID())
	runCtx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	messages := []adk.Message{schema.UserMessage(query)}
	iter := sr.runner.Run(runCtx, messages)

	var replyContent string
	var runErr error
	var promptTokens, completionTokens, promptCachedTokens int
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		fmt.Println(event.AgentName)
		if event.Err != nil {
			freedom.Logger().Debug("SystemRunner event error:", event.Err)
			runErr = event.Err
			continue
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}

		mv := event.Output.MessageOutput
		if mv.IsStreaming {
			for {
				chunk, err := mv.MessageStream.Recv()
				if err != nil {
					break
				}
				if chunk.Content != "" {
					replyContent += chunk.Content
				}
				if chunk.ResponseMeta != nil && chunk.ResponseMeta.Usage != nil {
					promptTokens = max(promptTokens, chunk.ResponseMeta.Usage.PromptTokens)
					completionTokens = max(completionTokens, chunk.ResponseMeta.Usage.CompletionTokens)
					promptCachedTokens = max(promptCachedTokens, chunk.ResponseMeta.Usage.PromptTokenDetails.CachedTokens)
				}
			}
		} else {
			msg, err := mv.GetMessage()
			if err != nil || msg == nil {
				continue
			}
			jdata, _ := json.Marshal(msg)
			fmt.Println(string(jdata))
			if msg.Role == schema.Assistant && msg.Content != "" {
				replyContent += msg.Content
			}
			if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
				promptTokens += msg.ResponseMeta.Usage.PromptTokens
				completionTokens += msg.ResponseMeta.Usage.CompletionTokens
				promptCachedTokens += msg.ResponseMeta.Usage.PromptTokenDetails.CachedTokens
			}
		}
	}

	if runErr != nil {
		err = runErr
		return
	}

	if sr.sysAgent.dailyStatsRepo != nil {
		sr.sysAgent.dailyStatsRepo.AddDailyStats(sr.sysAgent.userId, promptTokens, completionTokens, promptCachedTokens)
	}
	result = parseSystemResult(replyContent)
	return
}

var systemResultRe = regexp.MustCompile(`<system_result>\s*(\{[\s\S]*?\})\s*</system_result>`)

func parseSystemResult(replyContent string) *SystemResult {
	matches := systemResultRe.FindStringSubmatch(replyContent)
	if len(matches) < 2 {
		summary := "系统代理输出结果为空"
		if config.Get().GetLanguage() == "en" {
			summary = "system agent output is empty"
		}
		return &SystemResult{Status: "unknown", Summary: summary}
	}
	var result SystemResult
	if err := json.Unmarshal([]byte(matches[1]), &result); err != nil {
		summary := "系统代理输出格式解析失败: " + err.Error()
		if config.Get().GetLanguage() == "en" {
			summary = "failed to parse system agent output: " + err.Error()
		}
		return &SystemResult{Status: "parse_error", Summary: summary}
	}
	return &result
}
