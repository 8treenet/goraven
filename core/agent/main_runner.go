package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"goraven/backend/po"
	"goraven/config"
	"goraven/core/plugin"
	"goraven/util"
	"io"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	"github.com/8treenet/freedom"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func newMainRunner(main *MainAgent, runnerAgent *adk.ChatModelAgent, enableStreaming bool) *MainRunner {
	result := &MainRunner{
		mainAgent: main,
		RoundId:   util.UUID(),
	}

	result.runner = adk.NewRunner(context.Background(), adk.RunnerConfig{
		Agent:           runnerAgent,
		EnableStreaming: enableStreaming,
	})
	return result
}

type collectedMsg struct {
	roleType         string
	content          string
	reasoningContent string
	ext              string
	toolCallIds      []string
	toolCallsInfo    string
}

// activeRunners 全局运行器映射表，跨请求持久化
// Runner 结束时 self-cleanup（见 Query defer）
var activeRunners sync.Map
var runnerHold sync.Map

func RegisterRunnerHold(sessionId string) {
	runnerHold.Store(sessionId, true)
}

// ClearRunnerHold releases a hold without registering a runner. Call this when
// agent construction fails after RegisterRunnerHold so GetRunner does not block
// for the full 30s waiting for a runner that will never arrive.
func ClearRunnerHold(sessionId string) {
	runnerHold.Delete(sessionId)
}

// RegisterRunner 将 runner 注册到全局映射
func RegisterRunner(sessionId string, runner *MainRunner) {
	activeRunners.Store(sessionId, runner)
	runnerHold.Delete(sessionId)
}

// GetRunner 从全局映射获取活跃运行器
// 如果 runner 处于 hold 状态（StartChat goroutine 仍在构建），最多等待 30s。
// 超时后删除 hold 作为安全阀，避免后续 GetRunner 反复阻塞 30s。
// 正常路径下 hold 已由 RegisterRunner 释放，此处的 Delete 是 no-op。
func GetRunner(sessionId string) (*MainRunner, bool) {
	for i := 0; i < 60; i++ {
		_, ok := runnerHold.Load(sessionId)
		if !ok {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	runnerHold.Delete(sessionId)
	val, ok := activeRunners.Load(sessionId)
	if !ok {
		return nil, false
	}
	return val.(*MainRunner), true
}

// DeleteRunner 从全局映射删除运行器
func DeleteRunner(sessionId string) {
	activeRunners.Delete(sessionId)
}

type MainRunner struct {
	RoundId              string
	mainAgent            *MainAgent
	runner               *adk.Runner
	history              []adk.Message
	sseChan              chan *SSEEvent
	saveReplyContent     string
	saveReasoningContent string
	// 0未开始拉取，1开始拉取
	fetchStatus int
	stop        bool
	Mutex       sync.Mutex
	onComplete  func(*RunnerCompleteEvent)
	cancelCtx   context.CancelFunc
}

// RunnerCompleteEvent OnComplete 回调入参，暴露助手最终回复内容与运行结果状态：
// Terminated 表示被终止（用户停止/超时/限额），Err 为本轮模型或工具链错误信息（可为空）。
type RunnerCompleteEvent struct {
	Reply      string
	Err        string
	Terminated bool
}

func (runner *MainRunner) Terminat() {
	defer runner.Mutex.Unlock()
	runner.Mutex.Lock()
	runner.stop = true
	if runner.cancelCtx != nil {
		runner.cancelCtx()
	}
}

func (runner *MainRunner) IsStopped() bool {
	defer runner.Mutex.Unlock()
	runner.Mutex.Lock()
	return runner.stop
}

func (runner *MainRunner) sendSSEEvent(event *SSEEvent) {
	if event.Type == SSEEventTypeReasoning {
		runner.flushReasoningContent(event.Content)
	}
	if event.Type == SSEEventTypeContent {
		runner.flushReplyContent(event.Content)
	}
	runner.dispatchSSE(event)
}

// sendSSEEventNoFlush sends event to SSE channel without accumulating to internal reply/reasoning state.
// Used for sub-agent content that should be displayed live but not persisted to DB.
func (runner *MainRunner) sendSSEEventNoFlush(event *SSEEvent) {
	runner.dispatchSSE(event)
}

func (runner *MainRunner) dispatchSSE(event *SSEEvent) {
	if runner.IsStopped() && event.Type != SSEEventTypeEnd {
		return
	}

	sseData := &plugin.SSEEventData{
		Type:    event.Type,
		Content: event.Content,
	}
	if event.Tool != nil {
		sseData.Tool = &plugin.SSEEventToolData{
			Name:        event.Tool.Name,
			DisplayName: event.Tool.DisplayName,
			Icon:        event.Tool.Icon,
			Action:      event.Tool.Action,
		}
	}
	if event.Retry != nil {
		sseData.Retry = &plugin.SSERetryInfoData{
			MaxRetries: event.Retry.MaxRetries,
			Attempt:    event.Retry.Attempt,
			Error:      event.Retry.Error,
		}
	}
	if event.Context != nil {
		sseData.Context = &plugin.SSEContextData{
			Tokens: event.Context.Tokens,
			Limit:  event.Context.Limit,
		}
	}

	filtered := runner.mainAgent.Plugins.FireSSEHook(context.Background(), sseData)
	if filtered != nil {
		event.Type = filtered.Type
		event.Content = filtered.Content
		event.Tool = nil
		event.Retry = nil
		event.Context = nil
		if filtered.Tool != nil {
			event.Tool = &SSEEventTool{
				Name:        filtered.Tool.Name,
				DisplayName: filtered.Tool.DisplayName,
				Icon:        filtered.Tool.Icon,
				Action:      filtered.Tool.Action,
			}
		}
		if filtered.Retry != nil {
			event.Retry = &SSERetryInfo{
				MaxRetries: filtered.Retry.MaxRetries,
				Attempt:    filtered.Retry.Attempt,
				Error:      filtered.Retry.Error,
			}
		}
		if filtered.Context != nil {
			event.Context = &SSEContextInfo{
				Tokens: filtered.Context.Tokens,
				Limit:  filtered.Context.Limit,
			}
		}
	}

	if runner.GetFetchStatus() != 1 && event.Type != SSEEventTypeEnd {
		return
	}

	select {
	case runner.sseChan <- event:
	case <-time.After(500 * time.Millisecond):
		freedom.Logger().Debug("send sse event timeout")
	}
}

func (runner *MainRunner) sendSSEToolEvent(name, arguments string) (eventTool string) {
	if runner.mainAgent.param.DailyStatsRepo != nil {
		runner.mainAgent.param.DailyStatsRepo.AddToolDailyStats(runner.mainAgent.param.UserId(), "tool", name)
	}

	tevent := runner.buildSSEEventTool(name, arguments)
	if tevent == nil {
		return ""
	}
	runner.sendSSEEvent(&SSEEvent{
		Type: SSEEventTypeTool,
		Tool: tevent,
	})
	eventToolBytes, _ := json.Marshal(tevent)
	eventTool = string(eventToolBytes)
	return
}

func (runner *MainRunner) flushReasoningContent(reasoningContent string) {
	defer runner.Mutex.Unlock()
	runner.Mutex.Lock()
	runner.saveReasoningContent += reasoningContent
}

func (runner *MainRunner) flushReplyContent(replyContent string) {
	defer runner.Mutex.Unlock()
	runner.Mutex.Lock()
	runner.saveReplyContent += replyContent
}

func (runner *MainRunner) GetReasoningContent() (reasoningContent string) {
	defer runner.Mutex.Unlock()
	runner.Mutex.Lock()
	reasoningContent = runner.saveReasoningContent
	return
}

func (runner *MainRunner) GetReplyContent() (replyContent string) {
	defer runner.Mutex.Unlock()
	runner.Mutex.Lock()
	replyContent = runner.saveReplyContent
	return
}

func (runner *MainRunner) GetFetchStatus() (status int) {
	defer runner.Mutex.Unlock()
	runner.Mutex.Lock()
	status = runner.fetchStatus
	return
}

func (runner *MainRunner) StopFetchStatus() {
	defer runner.Mutex.Unlock()
	runner.Mutex.Lock()
	runner.fetchStatus = 0
}

func (runner *MainRunner) StartFetch() <-chan *SSEEvent {
	defer runner.Mutex.Unlock()
	runner.Mutex.Lock()
	runner.fetchStatus = 1
	return runner.sseChan
}

func (runner *MainRunner) OnComplete(onComplete func(*RunnerCompleteEvent)) {
	runner.onComplete = onComplete
}

func (runner *MainRunner) startHeartbeat() {
	go func() {
		time.Sleep(15 * time.Second)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			r, ok := GetRunner(runner.mainAgent.param.SessionId())
			if !ok || r.GetFetchStatus() == 0 || r.IsStopped() {
				return
			}
			r.sendSSEEventNoFlush(&SSEEvent{Type: SSEEventTypeHeartbeat})
		}
	}()
}

func (runner *MainRunner) Query(runctx context.Context, content string) (err error) {
	if runner.IsStopped() {
		return errors.New("runner status not empty")
	}

	runctx = util.WithConversationHeader(runctx, runner.mainAgent.param.ChatModel.GetConversationHeaderKey(), runner.mainAgent.param.SessionId())

	runner.Mutex.Lock()
	runner.sseChan = make(chan *SSEEvent, 20000)
	runner.stop = false
	runner.fetchStatus = 0
	runner.Mutex.Unlock()

	runner.history, err = runner.getHistory()
	if err != nil {
		return
	}

	err = runner.saveQuery(content)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(runctx, time.Duration(runner.mainAgent.sysCfg.MainAgentTimeoutMinutes)*time.Minute)
	runner.cancelCtx = cancel

	runner.startHeartbeat()

	go func() {
		// 提前声明供 defer 闭包捕获：OnComplete 需要感知本轮是否报错/被终止
		var replyErrContent string
		var terminated bool
		defer cancel()
		defer func() {
			if r := recover(); r != nil {
				freedom.Logger().Errorf("MainRunner Query panic: %v\n%s", r, debug.Stack())
			}
			runner.Terminat()
			runner.sendSSEEvent(&SSEEvent{Type: SSEEventTypeEnd})
			close(runner.sseChan)
			DeleteRunner(runner.mainAgent.param.SessionId())
			runner.mainAgent.msgRepo.UpdateSessionStatus(runner.mainAgent.param.SessionId(), 0)

			if cb := runner.onComplete; cb != nil {
				cb(&RunnerCompleteEvent{Reply: runner.GetReplyContent(), Err: replyErrContent, Terminated: terminated})
			}
		}()

		var promptTokens, completionTokens, promptCachedTokens int
		var lastPromptTokens int
		var collectedMsgs []collectedMsg
		startTime := time.Now()

		roundCtx := &plugin.RoundContext{
			SessionID: runner.mainAgent.param.SessionId(),
			UserID:    runner.mainAgent.param.UserId(),
			RoundID:   runner.RoundId,
			Query:     content,
		}

		if runner.mainAgent.compress != nil {
			compressedHistory, err := runner.mainAgent.compress.DoCompress(context.Background(), runner, runner.history)
			if err != nil {
				freedom.Logger().Debugf("Compress error: %v", err)
			} else {
				runner.history = compressedHistory
			}
		}

		roundCtx.Messages = runner.history
		beforeRoundFailed := false
		if err := runner.mainAgent.Plugins.FireBeforeRound(roundCtx); err != nil {
			replyErrContent = err.Error()
			runner.sendSSEEvent(&SSEEvent{
				Type:    SSEEventTypeContent,
				Content: replyErrContent,
			})
			beforeRoundFailed = true
		}

		if !beforeRoundFailed {
			collectedMsgs, promptTokens, completionTokens, promptCachedTokens, lastPromptTokens, terminated, replyErrContent = runner.loop(ctx, content)
		}

		duration := int(time.Since(startTime).Milliseconds())
		adkMsg := schema.AssistantMessage(runner.GetReplyContent(), nil)
		adkMsg.ReasoningContent = runner.GetReasoningContent()
		adkMsg.ResponseMeta = &schema.ResponseMeta{
			Usage: &schema.TokenUsage{
				PromptTokens: promptTokens,
				PromptTokenDetails: schema.PromptTokenDetails{
					CachedTokens: promptCachedTokens,
				},
				CompletionTokens: completionTokens,
			},
		}

		reply := runner.GetReplyContent()
		reasoning := runner.GetReasoningContent()
		roundCtx.Reply = &reply
		roundCtx.ReasoningContent = &reasoning
		roundCtx.Stopped = terminated
		if err := runner.mainAgent.Plugins.FireAfterRound(roundCtx); err != nil {
			freedom.Logger().Errorf("plugin AfterRound: %v", err)
		}

		timestamp := util.Millisecond() + 2
		if terminated {
			collectedMsgs = filterIncompleteMessages(collectedMsgs)
		}
		runner.saveIntermediateMessages(&timestamp, collectedMsgs)
		runner.saveReply(&timestamp, replyErrContent, adkMsg, duration, terminated, lastPromptTokens)

		if runner.mainAgent.param.DailyStatsRepo != nil {
			runner.mainAgent.param.DailyStatsRepo.AddDailyStats(runner.mainAgent.param.UserId(), promptTokens, completionTokens, promptCachedTokens)
		}
	}()
	return
}

func (runner *MainRunner) loop(ctx context.Context, content string) (msgs []collectedMsg, promptTokens, completionTokens, promptCachedTokens int, lastPromptTokens int, terminated bool, replyErrContent string) {
	var toolNames []string
	iter := runner.runner.Run(ctx, append(runner.history, schema.UserMessage(content)))

	for {
		if runner.IsStopped() {
			terminated = true
			break
		}
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			replyErrContent = event.Err.Error()
			runner.sendSSEEvent(&SSEEvent{
				Type:    SSEEventTypeContent,
				Content: replyErrContent,
			})
			continue
		}

		freedom.Logger().Debugf("%s event", event.AgentName)
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}

		mv := event.Output.MessageOutput

		var msgPromptTokens, msgCompletionTokens, msgPromptCachedTokens int
		var collected collectedMsg
		var stop bool
		if mv.IsStreaming {
			collected, msgPromptTokens, msgCompletionTokens, msgPromptCachedTokens, stop = runner.streamLoop(event.AgentName, mv, &toolNames, &replyErrContent)
		} else {
			collected, msgPromptTokens, msgCompletionTokens, msgPromptCachedTokens = runner.nonStreamLoop(event.AgentName, mv, &toolNames)
		}
		if stop {
			terminated = true
			break
		}
		if replyErrContent != "" {
			runner.sendSSEEvent(&SSEEvent{
				Type:    SSEEventTypeContent,
				Content: replyErrContent,
			})
		}

		if collected.roleType != "" {
			msgs = append(msgs, collected)
		}

		promptTokens += msgPromptTokens
		completionTokens += msgCompletionTokens
		promptCachedTokens += msgPromptCachedTokens
		lastPromptTokens = msgPromptTokens

		if msgPromptTokens > 0 {
			runner.sendSSEEventNoFlush(&SSEEvent{
				Type: SSEEventTypeContext,
				Context: &SSEContextInfo{
					Tokens: msgPromptTokens,
					Limit:  runner.mainAgent.param.ChatModel.ContextLength(),
				},
			})
		}

		// 运行时 Token 限额检查：本轮累计 + 今日已用 超过日限额时终止会话
		if limit := runner.mainAgent.param.DailyTokenLimit; limit > 0 {
			if runner.mainAgent.param.DailyTokenUsed+promptTokens+completionTokens >= limit*1_000_000 {
				var limitExceededMsg string
				if config.Get().GetLanguage() == "en" {
					limitExceededMsg = fmt.Sprintf("\n⚠️ Daily token limit reached (%dM). This session has ended. Please contact your administrator to increase the quota or try again tomorrow.", limit)
				} else {
					limitExceededMsg = fmt.Sprintf("\n⚠️ 今日 Token 用量已达上限（%dM），会话已自动结束。如需继续使用，请联系管理员调整额度或明日再试。", limit)
				}
				runner.sendSSEEvent(&SSEEvent{
					Type:    SSEEventTypeContent,
					Content: limitExceededMsg,
				})
				terminated = true
				break
			}
		}

		delayMs := runner.mainAgent.sysCfg.LLMRequestDelayMs
		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}
	return
}

func (runner *MainRunner) streamLoop(agentName string, mv *adk.MessageVariant, toolNames *[]string, replyErrContent *string) (collectedMsg, int, int, int, bool) {
	stream := mv.MessageStream
	var streamContent string
	var streamReasoningContent string
	var streamToolCallID string
	var streamToolName string
	var eventTool string
	streamToolCalls := make(map[int]*po.ToolCallData)
	sentToolEvents := make(map[string]bool)
	// deferredIndices 记录本轮需要延迟到流式参数完整后再发送 SSE 事件的
	// tool call 索引（按 streamToolCalls 的 key），按出现顺序入队。
	// 同一 turn 内同名工具被多次调用（如两次 TaskUpdate 分别完成/删除不同任务）
	// 时按索引区分，确保每一次调用都独立发出事件、各自携带完整参数。
	deferredIndices := make([]int, 0)
	deferredSeenIdx := make(map[int]bool)
	var promptTokens, completionTokens, promptCachedTokens int
	isMainAgent := agentName == mainAgentName
	stop := false
	for {
		if runner.IsStopped() {
			stop = true
			break
		}
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			freedom.Logger().Debug("stream.Recv err:", err)
			*replyErrContent = err.Error()
			break
		}

		// Tool result stream: accumulate but don't emit SSE events
		if mv.Role == schema.Tool {
			streamContent += chunk.Content
			if chunk.ToolCallID != "" {
				streamToolCallID = chunk.ToolCallID
			}
			if chunk.ToolName != "" {
				streamToolName = chunk.ToolName
			}
			continue
		}

		if chunk.Content != "" {
			streamContent += chunk.Content
		}
		if chunk.ReasoningContent != "" {
			streamReasoningContent += chunk.ReasoningContent
		}

		if len(chunk.ToolCalls) > 0 {
			for _, tc := range chunk.ToolCalls {
				if tc.Function.Name != "" {
					if isMainAgent {
						*toolNames = append(*toolNames, tc.Function.Name)
						// 依赖参数渲染展示的工具（execute/TaskCreate/TaskUpdate 等），
						// 流式下参数尚未完整，延迟到流结束（参数齐备）后再发送，
						// 避免基于残缺参数误判命令/任务类型。
						// 依赖参数渲染展示的工具（execute/TaskUpdate 等），流式下参数尚未完整，
						// 延迟到流结束（参数齐备）后再发送，避免基于残缺参数误判命令/任务类型。
						// 按 tool call 索引收集而非按工具名，避免同 turn 同名工具被多次调用
						// 时只发一次事件、后续调用丢失 SSE 展示（如先完成 #2 再删除 #4）。
						if isDeferredToolEvent(tc.Function.Name) && tc.Index != nil {
							idx := *tc.Index
							if !deferredSeenIdx[idx] {
								deferredSeenIdx[idx] = true
								deferredIndices = append(deferredIndices, idx)
							}
						} else if !sentToolEvents[tc.Function.Name] {
							sentToolEvents[tc.Function.Name] = true
							eventTool = runner.sendSSEToolEvent(tc.Function.Name, tc.Function.Arguments)
						}
					}
				}
				if isMainAgent && tc.Index != nil {
					idx := *tc.Index
					existing, ok := streamToolCalls[idx]
					if !ok {
						streamToolCalls[idx] = &po.ToolCallData{
							ID:        tc.ID,
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						}
					} else {
						if tc.ID != "" {
							existing.ID = tc.ID
						}
						if tc.Function.Name != "" {
							existing.Name = tc.Function.Name
						}
						existing.Arguments += tc.Function.Arguments
					}
				}
			}
		}

		// Real-time SSE for reasoning and content.
		// Only the main agent's final answer (no tool calls) is flushed to DB state;
		// tool-call-phase content and sub-agent content use NoFlush.
		if chunk.ReasoningContent != "" {
			runner.sendSSEEventNoFlush(&SSEEvent{Type: SSEEventTypeReasoning, Content: chunk.ReasoningContent})
		}
		if chunk.Content != "" {
			if isMainAgent && len(streamToolCalls) == 0 {
				runner.sendSSEEvent(&SSEEvent{Type: SSEEventTypeContent, Content: chunk.Content})
			} else if isMainAgent {
				runner.sendSSEEventNoFlush(&SSEEvent{Type: SSEEventTypeContent, Content: chunk.Content})
			} else {
				runner.sendSSEEventNoFlush(&SSEEvent{Type: SSEEventTypeReasoning, Content: chunk.Content})
			}
		}

		if chunk.ResponseMeta != nil && chunk.ResponseMeta.Usage != nil {
			promptTokens = chunk.ResponseMeta.Usage.PromptTokens
			completionTokens = chunk.ResponseMeta.Usage.CompletionTokens
			promptCachedTokens = chunk.ResponseMeta.Usage.PromptTokenDetails.CachedTokens
		}
	}
	stream.Close()
	if stop {
		return collectedMsg{}, promptTokens, completionTokens, promptCachedTokens, stop
	}

	// 延迟工具事件至此：此时流式参数已完整，可基于完整参数精确分发展示
	// （execute 按命令首 token、TaskUpdate 带 taskId/status/subject 等）。
	// 按 tool call 索引顺序逐个发送，每次调用各自携带自己的完整参数，
	// 同 turn 同名工具被多次调用时每一次都会发出独立的 SSE 事件。
	for _, idx := range deferredIndices {
		tc := streamToolCalls[idx]
		if tc == nil {
			continue
		}
		eventTool = runner.sendSSEToolEvent(tc.Name, tc.Arguments)
	}

	// Only flush reasoning for main agent's final answer (no tool calls).
	// Tool-call-phase reasoning stays only in the intermediate message saved by saveIntermediateMessages.
	if isMainAgent && len(streamToolCalls) == 0 && streamReasoningContent != "" {
		runner.flushReasoningContent(streamReasoningContent)
	}

	if streamContent != "" && mv.Role == schema.Tool {
		freedom.Logger().Debugf("tool result stream %s", streamContent)
	}

	var result collectedMsg
	if mv.Role == schema.Tool {
		if streamContent != "" || streamToolCallID != "" {
			extData, _ := json.Marshal(po.ToolExt{ToolCallID: streamToolCallID, ToolName: streamToolName})
			var toolCallIds []string
			if streamToolCallID != "" {
				toolCallIds = []string{streamToolCallID}
			}
			result = collectedMsg{
				roleType:      po.RoleTypeTool,
				content:       streamContent,
				ext:           string(extData),
				toolCallIds:   toolCallIds,
				toolCallsInfo: eventTool,
			}
		}
	} else if len(streamToolCalls) > 0 {
		indices := make([]int, 0, len(streamToolCalls))
		for idx := range streamToolCalls {
			indices = append(indices, idx)
		}
		sort.Slice(indices, func(i, j int) bool { return indices[i] < indices[j] })
		var toolCalls []po.ToolCallData
		var toolCallIds []string
		for _, idx := range indices {
			toolCalls = append(toolCalls, *streamToolCalls[idx])
			if streamToolCalls[idx].ID != "" {
				toolCallIds = append(toolCallIds, streamToolCalls[idx].ID)
			}
		}
		extData, _ := json.Marshal(po.AssistantExt{ToolCalls: toolCalls})
		result = collectedMsg{
			roleType:         po.RoleTypeAssistant,
			content:          streamContent,
			reasoningContent: streamReasoningContent,
			ext:              string(extData),
			toolCallIds:      toolCallIds,
			toolCallsInfo:    eventTool,
		}
	}
	if !isMainAgent {
		return collectedMsg{}, promptTokens, completionTokens, promptCachedTokens, stop
	}

	return result, promptTokens, completionTokens, promptCachedTokens, stop
}

func (runner *MainRunner) nonStreamLoop(agentName string, mv *adk.MessageVariant, toolNames *[]string) (collectedMsg, int, int, int) {
	isMainAgent := agentName == mainAgentName
	msg, err := mv.GetMessage()
	if err != nil || msg == nil {
		return collectedMsg{}, 0, 0, 0
	}

	if mv.Role == schema.Tool {
		if msg.Content != "" || msg.ToolCallID != "" {
			extData, _ := json.Marshal(po.ToolExt{ToolCallID: msg.ToolCallID, ToolName: msg.ToolName})
			var toolCallIds []string
			if msg.ToolCallID != "" {
				toolCallIds = []string{msg.ToolCallID}
			}
			return collectedMsg{
				roleType:    po.RoleTypeTool,
				content:     msg.Content,
				ext:         string(extData),
				toolCallIds: toolCallIds,
			}, 0, 0, 0
		}
		return collectedMsg{}, 0, 0, 0
	}

	if isMainAgent && len(msg.ToolCalls) > 0 {
		for _, tc := range msg.ToolCalls {
			if tc.Function.Name != "" {
				*toolNames = append(*toolNames, tc.Function.Name)
			}
		}
	}

	if msg.ReasoningContent != "" {
		runner.sendSSEEventNoFlush(&SSEEvent{Type: SSEEventTypeReasoning, Content: msg.ReasoningContent})
	}

	if isMainAgent && msg.Content != "" && len(msg.ToolCalls) == 0 {
		runner.sendSSEEvent(&SSEEvent{Type: SSEEventTypeContent, Content: msg.Content})
	} else if isMainAgent && msg.Content != "" {
		runner.sendSSEEventNoFlush(&SSEEvent{Type: SSEEventTypeContent, Content: msg.Content})
	} else if msg.Content != "" {
		runner.sendSSEEventNoFlush(&SSEEvent{Type: SSEEventTypeReasoning, Content: msg.Content})
	}

	// Flush reasoning for main agent's final answer (no tool calls)
	if isMainAgent && len(msg.ToolCalls) == 0 && msg.ReasoningContent != "" {
		runner.flushReasoningContent(msg.ReasoningContent)
	}

	var promptTokens, completionTokens, promptCachedTokens int
	if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
		promptTokens = msg.ResponseMeta.Usage.PromptTokens
		completionTokens = msg.ResponseMeta.Usage.CompletionTokens
		promptCachedTokens = msg.ResponseMeta.Usage.PromptTokenDetails.CachedTokens
	}

	var result collectedMsg
	var eventTool string
	if isMainAgent {
		for _, tc := range msg.ToolCalls {
			if tc.Function.Name != "" {
				eventTool = runner.sendSSEToolEvent(tc.Function.Name, tc.Function.Arguments)
			}
		}
	}

	if len(msg.ToolCalls) > 0 {
		var toolCalls []po.ToolCallData
		var toolCallIds []string
		for _, tc := range msg.ToolCalls {
			toolCalls = append(toolCalls, po.ToolCallData{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			})
			if tc.ID != "" {
				toolCallIds = append(toolCallIds, tc.ID)
			}
		}
		extData, _ := json.Marshal(po.AssistantExt{ToolCalls: toolCalls})
		result = collectedMsg{
			roleType:         po.RoleTypeAssistant,
			content:          msg.Content,
			reasoningContent: msg.ReasoningContent,
			ext:              string(extData),
			toolCallIds:      toolCallIds,
			toolCallsInfo:    eventTool,
		}
	}

	if !isMainAgent {
		return collectedMsg{}, 0, 0, 0
	}
	return result, promptTokens, completionTokens, promptCachedTokens
}

func (runner *MainRunner) saveQuery(content string) error {
	userMsg := &po.Message{
		SessionId: runner.mainAgent.param.SessionId(),
		Timestamp: util.Millisecond() + 2,
		Content:   content,
		RoleType:  po.RoleTypeUser,
		Created:   time.Now(),
		Updated:   time.Now(),
		RoundId:   runner.RoundId,
	}
	err := runner.mainAgent.msgRepo.SaveChatMessage(runner.mainAgent.param.SessionId(), userMsg)
	if err != nil {
		freedom.Logger().Errorf("SaveChatMessage %v", err)
		return err
	}
	err = runner.mainAgent.msgRepo.UpdateSessionStatus(runner.mainAgent.param.SessionId(), 1)
	return err
}

func (runner *MainRunner) saveReply(timestamp *int64, replyErrContent string, replyContent adk.Message, duration int, terminated bool, contextTokens int) {
	if terminated && replyContent.Content == "" && replyErrContent == "" {
		return
	}
	contextState := uint8(0)
	if terminated {
		contextState = 1
	}
	replyMsg := &po.Message{
		SessionId:               runner.mainAgent.param.SessionId(),
		Timestamp:               *timestamp + 1,
		Content:                 replyContent.Content,
		ReasoningContent:        replyContent.ReasoningContent,
		RoleType:                po.RoleTypeAssistant,
		ContextState:            contextState,
		PromptTokensCount:       replyContent.ResponseMeta.Usage.PromptTokens,
		CompletionTokensCount:   replyContent.ResponseMeta.Usage.CompletionTokens,
		PromptCachedTokensCount: replyContent.ResponseMeta.Usage.PromptTokenDetails.CachedTokens,
		Duration:                duration,
		AsstError:               replyErrContent,
		Created:                 time.Now(),
		Updated:                 time.Now(),
		RoundId:                 runner.RoundId,
	}
	if replyContent.Content == "" && replyErrContent != "" {
		replyMsg.Content = replyErrContent
	}
	err := runner.mainAgent.msgRepo.SaveChatMessage(runner.mainAgent.param.SessionId(), replyMsg)
	if err != nil {
		freedom.Logger().Errorf("SaveChatMessage %v", err)
		return
	}
	runner.mainAgent.msgRepo.AddSessionTokens(runner.mainAgent.param.SessionId(), replyContent.ResponseMeta.Usage.PromptTokens, replyContent.ResponseMeta.Usage.CompletionTokens, replyContent.ResponseMeta.Usage.PromptTokenDetails.CachedTokens)
	runner.mainAgent.msgRepo.SetContextTokens(runner.mainAgent.param.SessionId(), contextTokens)
}

func (runner *MainRunner) saveIntermediateMessages(timestamp *int64, msgs []collectedMsg) {
	sessionId := runner.mainAgent.param.SessionId()
	for _, m := range msgs {
		if m.content == "" && m.ext == "" {
			continue
		}
		*timestamp = *timestamp + 1
		poMsg := &po.Message{
			SessionId:        sessionId,
			Timestamp:        *timestamp,
			Content:          m.content,
			ReasoningContent: m.reasoningContent,
			RoleType:         m.roleType,
			Tool:             1,
			Ext:              m.ext,
			Created:          time.Now(),
			Updated:          time.Now(),
			RoundId:          runner.RoundId,
			ToolCallsInfo:    m.toolCallsInfo,
		}
		err := runner.mainAgent.msgRepo.SaveChatMessage(sessionId, poMsg)
		if err != nil {
			freedom.Logger().Errorf("SaveChatMessage intermediate %v", err)
		}
	}
}

func (runner *MainRunner) getHistory() (result []adk.Message, e error) {
	list, err := runner.mainAgent.msgRepo.GetChatMessages(runner.mainAgent.param.SessionId())
	if err != nil {
		e = err
		return
	}

	return BuildHistoryFromMessages(list), nil
}

func (runner *MainRunner) buildSSEEventTool(name, arguments string) *SSEEventTool {
	// 优先走参数展示解析器注册表（execute/TaskUpdate 等），
	// 命中后由 resolver 基于参数 + MainAgent 运行期状态生成展示；
	// 未命中再回退到 toolRegistry 静态项。
	if resolver, ok := toolDisplayResolvers[name]; ok {
		if display, matched := resolver(arguments, runner.mainAgent); matched {
			return buildSSEEventToolFromDisplay(name, display)
		}
	}
	if display, ok := toolRegistry[name]; ok {
		return buildSSEEventToolFromDisplay(name, display)
	}

	if mcp, ok := runner.mainAgent.GetMCPToolDisplay(name); ok {
		if runner.mainAgent.param.DailyStatsRepo != nil {
			runner.mainAgent.param.DailyStatsRepo.AddToolDailyStats(runner.mainAgent.param.UserId(), "mcp", mcp.Name)
		}
		displayName := mcp.DisplayName
		if displayName == "" {
			displayName = mcp.Name
		}
		useChinese := isChineseLanguage()
		action := "Executing " + displayName
		if useChinese {
			action = "正在执行 " + displayName
		}
		return &SSEEventTool{
			Name:        name,
			DisplayName: "MCP",
			Icon:        "🔌",
			Action:      action,
		}
	}

	return nil
}

func filterIncompleteMessages(msgs []collectedMsg) []collectedMsg {
	pending := make(map[string]bool)
	lastComplete := 0

	for i, msg := range msgs {
		if msg.roleType == po.RoleTypeAssistant {
			for _, id := range msg.toolCallIds {
				pending[id] = true
			}
		} else if msg.roleType == po.RoleTypeTool {
			for _, id := range msg.toolCallIds {
				delete(pending, id)
			}
		}

		if len(pending) == 0 {
			lastComplete = i + 1
		}
	}

	if lastComplete < len(msgs) {
		return msgs[:lastComplete]
	}
	return msgs
}
