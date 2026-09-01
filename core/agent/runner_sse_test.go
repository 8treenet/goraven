package agent

import (
	"sync"
	"testing"
)

func TestHoldRunnerCAS(t *testing.T) {
	if ok := HoldRunner("cas-session-1"); !ok {
		t.Errorf("first HoldRunner should succeed")
	}
	if ok := HoldRunner("cas-session-1"); ok {
		t.Errorf("second HoldRunner for same session should fail")
	}
	ClearRunnerHold("cas-session-1")
	if ok := HoldRunner("cas-session-1"); !ok {
		t.Errorf("HoldRunner should succeed after ClearRunnerHold")
	}
	ClearRunnerHold("cas-session-1")
}

func TestSendSSEConcurrentWithClose(t *testing.T) {
	runner := &MainRunner{}
	runner.sseChan = make(chan *SSEEvent, 1)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.sendSSE(&SSEEvent{Type: SSEEventTypeContent})
		}()
	}
	runner.closeSSEChan()
	wg.Wait()

	// 重复 close 不应 panic
	runner.closeSSEChan()
}
