package mock

import (
	"testing"
	"time"

	"goraven/core/agent"
)

// TestStreamMockDurationBudget 校验 mock 脚本的“预算”总时长 >= 4 分钟。
// 本测试只读取常量并求和，不会真正启动 goroutine，避免 CI 阻塞 4 分钟。
func TestStreamMockDurationBudget(t *testing.T) {
	steps := scriptSteps()

	const toolSleepCap = 30 * time.Second
	const stepDelay = StreamMockStepDelayMs * time.Millisecond

	var total time.Duration
	var toolCount, contentCount, reasoningCount, retryCount, endCount int

	for _, s := range steps {
		if s.postWork > toolSleepCap {
			t.Fatalf("step postWork %v exceeds tool-sleep cap %v", s.postWork, toolSleepCap)
		}
		total += s.preSleep + s.postWork
		switch s.kind {
		case "content":
			contentCount++
			// 估算：内容文本长度 / ChunkSize × ChunkDelay
			chunks := (len([]rune(s.content)) + StreamMockChunkSize - 1) / StreamMockChunkSize
			if chunks > 0 {
				total += time.Duration(chunks-1) * time.Duration(StreamMockChunkDelayMs) * time.Millisecond
			}
		case "tool":
			toolCount++
		case "reasoning":
			reasoningCount++
		case "retry":
			retryCount++
		case "end":
			endCount++
		}
	}

	t.Logf("budget=%v | reasoning=%d tool=%d content=%d retry=%d end=%d",
		total.Truncate(time.Second), reasoningCount, toolCount, contentCount, retryCount, endCount)

	if endCount != 1 {
		t.Errorf("expected exactly 1 end event, got %d", endCount)
	}
	if contentCount < 2 {
		t.Errorf("expected at least 2 content blocks, got %d", contentCount)
	}
	if toolCount < 2 {
		t.Errorf("expected at least 2 tool events, got %d", toolCount)
	}
	if total < 4*time.Minute {
		t.Errorf("total budget %v is less than 4 minutes", total)
	}
	_ = stepDelay
}

// TestStreamMockEventDiversity 校验 mock 覆盖了所有 4 种用户关心的 SSE 事件类型。
func TestStreamMockEventDiversity(t *testing.T) {
	steps := scriptSteps()
	seen := map[string]bool{}
	for _, s := range steps {
		seen[s.kind] = true
	}
	for _, kind := range []string{"reasoning", "content", "tool", "retry", "end"} {
		if !seen[kind] {
			t.Errorf("script does not include event kind %q", kind)
		}
	}
}

// TestStreamMockHeartbeat 校验 BuildStreamMock 在脚本运行期间会按 ~30s 节奏
// 发出 heartbeat（前端俗称 ping）事件。
//
// 为了不让测试跑满 4 分钟，单独起一个短脚本：只发一个 tool 事件并 sleep ~65s，
// 期间应至少观察到 2 个 heartbeat。
func TestStreamMockHeartbeat(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping heartbeat test in short mode")
	}

	ch := make(chan *agent.SSEEvent, 100)
	stop := make(chan struct{})

	go func() {
		defer close(stop)
		// 一个会持续 65s 的最小脚本：tool + 65s sleep
		ch <- &agent.SSEEvent{
			Type: agent.SSEEventTypeTool,
			Tool: &agent.SSEEventTool{Name: "ls", DisplayName: "文件系统", Icon: "📁", Action: "scanning"},
		}
		time.Sleep(65 * time.Second)
		ch <- &agent.SSEEvent{Type: agent.SSEEventTypeEnd}
		close(ch)
	}()
	go runHeartbeat(ch, stop)

	heartbeats := 0
	deadline := time.After(70 * time.Second)
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				if heartbeats < 2 {
					t.Errorf("expected at least 2 heartbeats within 65s, got %d", heartbeats)
				}
				return
			}
			if ev.Type == agent.SSEEventTypeHeartbeat {
				heartbeats++
				t.Logf("heartbeat received at t=%v", time.Since(time.Now()).String())
			}
		case <-deadline:
			t.Fatalf("test timed out, only got %d heartbeats", heartbeats)
		}
	}
}
