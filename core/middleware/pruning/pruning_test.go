/*
 * Copyright 2026 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pruning

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func newTestMiddleware(cfg *Config) adk.ChatModelAgentMiddleware {
	if cfg.TokenCounter == nil {
		cfg.TokenCounter = defaultTokenCounter
	}
	return &pruningMiddleware{config: cfg}
}

func makeRound(callID, toolName, args, content string, ts int64) *toolRound {
	asst := createAssistantToolCall(callID, toolName, args, ts)
	tool := createToolMessage(content, callID, ts)
	return &toolRound{
		assistantMsg: asst,
		assistantIdx: 0,
		toolMsgs:     []*schema.Message{tool},
		toolIndices:  []int{1},
		timestamp:    ts,
	}
}

// ============================================================================
// 1. 未超阈值：不做任何操作
// ============================================================================

func TestPruneRounds_BelowThreshold(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      100000,
		MaxToolResultLength: 100,
		HeadLength:          40,
		TailLength:          40,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, strings.Repeat("x", 200), now-10000),
		makeRound("call_2", "T2", `{}`, "short", now),
	}

	pruned := mw.pruneRounds(rounds, 0)
	assert.Len(t, pruned, 2)
	for _, r := range pruned {
		for _, tm := range r.toolMsgs {
			assert.NotContains(t, tm.Content, "[System:")
		}
	}
}

// ============================================================================
// 2. 超阈值，截断即满足：不丢弃轮次
// ============================================================================

func TestPruneRounds_TruncationSuffices(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      500,
		MaxToolResultLength: 100,
		HeadLength:          40,
		TailLength:          40,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	longContent := strings.Repeat("abcdefgh", 500) // 4000 chars, ~1000 tokens

	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, longContent, now-100000),
		makeRound("call_2", "T2", `{}`, "short", now),
	}

	pruned := mw.pruneRounds(rounds, 0)
	assert.Len(t, pruned, 2, "截断后满足阈值，两轮都应保留")
	assert.Contains(t, pruned[0].toolMsgs[0].Content, "[System:", "超长内容应被截断")
	assert.Equal(t, "short", pruned[1].toolMsgs[0].Content, "短内容保留原文")
}

// ============================================================================
// 3. 超阈值，截断不够，需丢弃旧轮次
// ============================================================================

func TestPruneRounds_TruncationNotEnough_DropsOldRounds(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      40,
		MaxToolResultLength: 100,
		HeadLength:          40,
		TailLength:          40,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	longContent := strings.Repeat("a", 500) // ~125 tokens before truncation, ~45 after

	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, longContent, now-100000), // 截断后~45 tokens
		makeRound("call_2", "T2", `{}`, "hello", now-1000),       // ~2 tokens
		makeRound("call_3", "T3", `{}`, "ok", now),               // ~1 token
	}

	pruned := mw.pruneRounds(rounds, 0)

	// call_1: 截断后~45 tokens, 加上 call_2+call_3 ~48 > 40 → 丢弃
	for _, r := range pruned {
		assert.NotEqual(t, "call_1", r.assistantMsg.ToolCalls[0].ID, "call_1 应被丢弃")
	}
	// call_2 + call_3 保留
	assert.Len(t, pruned, 2)
}

// ============================================================================
// 4. 阈值极小，保留最后1轮
// ============================================================================

func TestPruneRounds_KeepsAtLeastOne(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      1,
		MaxToolResultLength: 100,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, "r1", now-3000),
		makeRound("call_2", "T2", `{}`, "r2", now-2000),
		makeRound("call_3", "T3", `{}`, "r3", now),
	}

	pruned := mw.pruneRounds(rounds, 0)
	assert.Len(t, pruned, 1, "应保留至少1轮")
	assert.Equal(t, "call_3", pruned[0].assistantMsg.ToolCalls[0].ID)
}

// ============================================================================
// 5. 部分截断部分丢弃：混合场景
// ============================================================================

func TestPruneRounds_MixedTruncateAndDrop(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      70,
		MaxToolResultLength: 100,
		HeadLength:          40,
		TailLength:          40,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	longContent := strings.Repeat("a", 500)

	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, "short", now-100000),    // 短，不截断
		makeRound("call_2", "T2", `{}`, longContent, now-50000), // 超长
		makeRound("call_3", "T3", `{}`, "result", now),
	}

	pruned := mw.pruneRounds(rounds, 0)

	// call_1: 短内容，不截断但直接丢弃（最旧）
	for _, r := range pruned {
		assert.NotEqual(t, "call_1", r.assistantMsg.ToolCalls[0].ID, "call_1 应被丢弃")
	}

	// call_2: 超长，截断后满足阈值 → 保留但截断
	var found2 *toolRound
	for _, r := range pruned {
		if r.assistantMsg.ToolCalls[0].ID == "call_2" {
			found2 = r
		}
	}
	assert.NotNil(t, found2, "call_2 应保留")
	assert.Contains(t, found2.toolMsgs[0].Content, "[System:", "call_2 应被截断")

	assert.Len(t, pruned, 2)
}

// ============================================================================
// 6. 短内容不被截断（仅超长才触发）
// ============================================================================

func TestPruneRounds_ShortContentNotTruncated(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      10,
		MaxToolResultLength: 100,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, "hello world", now),
	}

	pruned := mw.pruneRounds(rounds, 0)
	assert.Len(t, pruned, 1)
	assert.Equal(t, "hello world", pruned[0].toolMsgs[0].Content)
}

// ============================================================================
// 7. 消息对完整性：toolCall 与 tool 响应成对
// ============================================================================

func TestMessagePairIntegrity(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      10,
		MaxToolResultLength: 100,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	messages := []adk.Message{
		schema.SystemMessage("system"),
		schema.UserMessage("Q1"),
		createAssistantToolCall("call_1", "T1", `{}`, now-100),
		createToolMessage("Result 1", "call_1", now-100),
		schema.UserMessage("Q2"),
		createAssistantToolCall("call_2", "T2", `{}`, now),
		createToolMessage("Result 2", "call_2", now),
	}

	state := &adk.ChatModelAgentState{Messages: messages}
	_, newState, err := mw.BeforeModelRewriteState(context.Background(), state, &adk.ModelContext{})
	assert.NoError(t, err)

	for i, msg := range newState.Messages {
		if msg.Role == schema.Assistant && len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				matched := false
				for j := i + 1; j < len(newState.Messages); j++ {
					if newState.Messages[j].Role == schema.Assistant && len(newState.Messages[j].ToolCalls) > 0 {
						break
					}
					if newState.Messages[j].Role == schema.Tool && newState.Messages[j].ToolCallID == tc.ID {
						matched = true
						break
					}
				}
				assert.True(t, matched, fmt.Sprintf("toolCall %s 必须有对应的 tool 响应", tc.ID))
			}
		}
	}
}

// ============================================================================
// 8. 并行 tool calling 轮次完整性
// ============================================================================

func TestParallelToolCallingRoundIntegrity(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      10,
		MaxToolResultLength: 100000,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	round2Asst := createAssistantMultiToolCall([]toolCallSpec{
		{"call_2a", "GetWeather", `{"loc":"北京"}`},
		{"call_2b", "GetWeather", `{"loc":"上海"}`},
	}, now-100000)

	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, "r1", now-200000),
		{
			assistantMsg: round2Asst,
			assistantIdx: 2,
			toolMsgs: []*schema.Message{
				createToolMessage("晴 25℃", "call_2a", now-100000),
				createToolMessage("多云 22℃", "call_2b", now-100000),
			},
			toolIndices: []int{3, 4},
			timestamp:   now - 100000,
		},
		makeRound("call_3", "T3", `{}`, "r3", now),
	}

	// 设 baseTokens 使丢弃 round1 后刚好满足阈值
	round2Tokens := estimateRoundTokens(rounds[1])
	round3Tokens := estimateRoundTokens(rounds[2])
	baseTokens := cfg.TokenThreshold - (round2Tokens + round3Tokens)

	pruned := mw.pruneRounds(rounds, baseTokens)

	// round1 被丢弃
	for _, r := range pruned {
		assert.NotEqual(t, "call_1", r.assistantMsg.ToolCalls[0].ID, "call_1 应被丢弃")
	}

	// round2 保留且完整
	var found2 *toolRound
	for _, r := range pruned {
		if len(r.assistantMsg.ToolCalls) >= 2 && r.assistantMsg.ToolCalls[0].ID == "call_2a" {
			found2 = r
			break
		}
	}
	assert.NotNil(t, found2)
	assert.Len(t, found2.assistantMsg.ToolCalls, 2)
	assert.Len(t, found2.toolMsgs, 2)
}

// ============================================================================
// extractToolRounds
// ============================================================================

func TestExtractToolRounds(t *testing.T) {
	t.Run("single tool call", func(t *testing.T) {
		now := time.Now().UnixMilli()
		messages := []adk.Message{
			schema.SystemMessage("system"),
			schema.UserMessage("user1"),
			createAssistantToolCall("call_1", "T1", `{}`, now),
			createToolMessage("r1", "call_1", now),
			schema.UserMessage("user2"),
			createAssistantToolCall("call_2", "T2", `{}`, now),
			createToolMessage("r2", "call_2", now),
		}

		rounds := extractToolRounds(messages)
		assert.Len(t, rounds, 2)
		assert.Equal(t, "call_1", rounds[0].assistantMsg.ToolCalls[0].ID)
		assert.Len(t, rounds[0].toolMsgs, 1)
		assert.Equal(t, "call_2", rounds[1].assistantMsg.ToolCalls[0].ID)
		assert.Len(t, rounds[1].toolMsgs, 1)
	})

	t.Run("parallel tool calls", func(t *testing.T) {
		now := time.Now().UnixMilli()
		asst := createAssistantMultiToolCall([]toolCallSpec{
			{"call_A", "GetWeather", `{"loc":"北京"}`},
			{"call_B", "GetWeather", `{"loc":"上海"}`},
		}, now)
		messages := []adk.Message{
			schema.SystemMessage("system"),
			schema.UserMessage("query"),
			asst,
			createToolMessage("晴 25℃", "call_A", now),
			createToolMessage("多云 22℃", "call_B", now),
		}

		rounds := extractToolRounds(messages)
		assert.Len(t, rounds, 1, "多 toolCall 聚合为 1 个 round")
		assert.Len(t, rounds[0].toolMsgs, 2)
		assert.Equal(t, "call_A", rounds[0].toolMsgs[0].ToolCallID)
		assert.Equal(t, "call_B", rounds[0].toolMsgs[1].ToolCallID)
	})

	t.Run("no tool rounds", func(t *testing.T) {
		messages := []adk.Message{
			schema.SystemMessage("system"),
			schema.UserMessage("user"),
			schema.AssistantMessage("response", nil),
		}

		rounds := extractToolRounds(messages)
		assert.Len(t, rounds, 0)
	})
}

// ============================================================================
// truncateContent
// ============================================================================

func TestTruncateContent(t *testing.T) {
	t.Run("short content unchanged", func(t *testing.T) {
		result := truncateContent("hello", 10, 10)
		assert.Equal(t, "hello", result)
	})

	t.Run("long content truncated", func(t *testing.T) {
		result := truncateContent(strings.Repeat("a", 1000), 100, 100)
		assert.Contains(t, result, "[System: 因超出上下文限制")
		assert.True(t, len(result) < 1000)
	})
}

// ============================================================================
// estimateStrTokens
// ============================================================================

func TestEstimateStrTokens(t *testing.T) {
	assert.Equal(t, int64(1), estimateStrTokens("ab"))
	assert.Equal(t, int64(2), estimateStrTokens("hello"))
	assert.Equal(t, int64(3), estimateStrTokens("hello world"))
	assert.Equal(t, int64(6), estimateStrTokens("你好世界")) // 4 runes * 3 / 2
}

// ============================================================================
// 9. 非零 baseTokens
// ============================================================================

func TestPruneRounds_WithBaseTokens(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      50,
		MaxToolResultLength: 100,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, "result1", now-2000),
		makeRound("call_2", "T2", `{}`, "result2", now),
	}

	baseTokens := int64(45)
	pruned := mw.pruneRounds(rounds, baseTokens)
	assert.Len(t, pruned, 1, "baseTokens(45) + 保留1轮 > 50, 应丢弃旧轮")
	assert.Equal(t, "call_2", pruned[0].assistantMsg.ToolCalls[0].ID)
}

// ============================================================================
// 10. 长内容轮次截断后仍超阈值 → 丢弃
// ============================================================================

func TestPruneRounds_LongRoundStillOverAfterTruncation(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      10,
		MaxToolResultLength: 100,
		HeadLength:          40,
		TailLength:          40,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	longContent := strings.Repeat("a", 500)

	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, longContent, now-100000), // 长内容，截断后~40tokens > 10
		makeRound("call_2", "T2", `{}`, "ok", now-50000),         // 短内容
		makeRound("call_3", "T3", `{}`, "last", now),             // 最后1轮
	}

	pruned := mw.pruneRounds(rounds, 0)
	assert.Len(t, pruned, 2, "call_1 截断后仍超阈值应被丢弃, 保留 call_2 call_3")
	for _, r := range pruned {
		assert.NotEqual(t, "call_1", r.assistantMsg.ToolCalls[0].ID, "call_1 应被丢弃")
	}
}

// ============================================================================
// 11. 单轮始终保留
// ============================================================================

func TestPruneRounds_SingleRoundAlwaysKept(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      1,
		MaxToolResultLength: 2000,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, "very long result", now),
	}

	pruned := mw.pruneRounds(rounds, 0)
	assert.Len(t, pruned, 1, "单轮始终保留")
	assert.Equal(t, "very long result", pruned[0].toolMsgs[0].Content, "单轮内容不被修改")
}

// ============================================================================
// 12. 截断后 token 反而增加时正确丢弃
// ============================================================================

func TestPruneRounds_TruncationIncreasesTokens_DropsRound(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      30,
		MaxToolResultLength: 100,
		HeadLength:          40,
		TailLength:          40,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	now := time.Now().UnixMilli()
	barelyLong := strings.Repeat("a", 101)

	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, barelyLong, now-1000),
		makeRound("call_2", "T2", `{}`, "short", now),
	}

	round1Tokens := estimateRoundTokens(rounds[0]) // ~26 原始
	round2Tokens := estimateRoundTokens(rounds[1]) // ~2
	t.Logf("round1 before truncation: %d tokens", round1Tokens)
	t.Logf("threshold: %d", cfg.TokenThreshold)
	t.Logf("total before: %d", round1Tokens+round2Tokens)

	pruned := mw.pruneRounds(rounds, 0)
	t.Logf("pruned count: %d", len(pruned))

	assert.Len(t, pruned, 1, "截断后token反而增加，应丢弃该轮")
	assert.Equal(t, "call_2", pruned[0].assistantMsg.ToolCalls[0].ID)
}

// ============================================================================
// 13. 恰好等于阈值时不剪枝
// ============================================================================

func TestPruneRounds_ExactlyAtThreshold(t *testing.T) {
	now := time.Now().UnixMilli()
	rounds := []*toolRound{
		makeRound("call_1", "T1", `{}`, "hello", now-1000),
		makeRound("call_2", "T2", `{}`, "world", now),
	}

	totalTokens := estimateRoundTokens(rounds[0]) + estimateRoundTokens(rounds[1])

	cfg := &Config{
		TokenThreshold:      totalTokens,
		MaxToolResultLength: 100,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	pruned := mw.pruneRounds(rounds, 0)
	assert.Len(t, pruned, 2, "恰好等于阈值不应剪枝")
}

// ============================================================================
// 14. 空轮次列表
// ============================================================================

func TestPruneRounds_EmptyRounds(t *testing.T) {
	cfg := &Config{
		TokenThreshold:      1,
		MaxToolResultLength: 100,
	}
	mw := newTestMiddleware(cfg).(*pruningMiddleware)

	pruned := mw.pruneRounds(nil, 0)
	assert.Len(t, pruned, 0)

	pruned = mw.pruneRounds([]*toolRound{}, 0)
	assert.Len(t, pruned, 0)
}

// ============================================================================
// 辅助函数
// ============================================================================

func createAssistantToolCall(callID, toolName, arguments string, timestamp int64) *schema.Message {
	msg := schema.AssistantMessage("", []schema.ToolCall{
		{
			ID:   callID,
			Type: "function",
			Function: schema.FunctionCall{
				Name:      toolName,
				Arguments: arguments,
			},
		},
	})
	msg.Extra = map[string]any{"timestamp": timestamp}
	return msg
}

type toolCallSpec struct {
	id   string
	name string
	args string
}

func createAssistantMultiToolCall(specs []toolCallSpec, timestamp int64) *schema.Message {
	toolCalls := make([]schema.ToolCall, 0, len(specs))
	for _, s := range specs {
		toolCalls = append(toolCalls, schema.ToolCall{
			ID:   s.id,
			Type: "function",
			Function: schema.FunctionCall{
				Name:      s.name,
				Arguments: s.args,
			},
		})
	}
	msg := schema.AssistantMessage("", toolCalls)
	msg.Extra = map[string]any{"timestamp": timestamp}
	return msg
}

func createToolMessage(content, toolCallID string, timestamp int64) *schema.Message {
	msg := schema.ToolMessage(content, toolCallID)
	msg.Extra = map[string]any{"timestamp": timestamp}
	return msg
}
