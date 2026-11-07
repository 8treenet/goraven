package local_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"goraven/core/sandbox/local"
	"goraven/core/sandbox/types"

	"github.com/cloudwego/eino/adk/filesystem"
)

// processState 读取 /proc/<pid>/stat 的进程状态字段。
// 返回 ("", false) 表示进程已被回收（/proc 条目消失）。
func processState(pid int) (string, bool) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return "", false
	}
	// comm 字段可能包含空格与括号，取最后一个 ')' 之后的内容。
	text := string(data)
	idx := strings.LastIndex(text, ")")
	if idx < 0 || idx+2 >= len(text) {
		return "", true
	}
	fields := strings.Fields(text[idx+2:])
	if len(fields) == 0 {
		return "", true
	}
	return fields[0], true
}

// assertChildReaped 断言被终止的子进程已被父进程回收；超时后仍留在
// /proc 中的 "Z" 状态即僵尸进程（<defunct>）。
func assertChildReaped(t *testing.T, pidFile string) {
	t.Helper()
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("读取子进程 pid 失败（命令可能未启动）: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("解析子进程 pid 失败: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	state := ""
	alive := true
	for time.Now().Before(deadline) {
		state, alive = processState(pid)
		if !alive || state == "Z" || state == "X" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if !alive {
		return // 已被 Wait 回收，符合预期
	}
	if state == "Z" {
		t.Fatalf("子进程 pid=%d 在终止后成为僵尸进程（state=Z）：cmd.Wait() 未在非正常分支调用", pid)
	}
	t.Fatalf("子进程 pid=%d 在终止后仍存活（state=%s）：未被终止/回收", pid, state)
}

func drain(t *testing.T, sr interface {
	Recv() (*filesystem.ExecuteResponse, error)
}) {
	t.Helper()
	for {
		if _, err := sr.Recv(); err != nil {
			return
		}
	}
}

// TestExecuteStreamingTimeoutReapsChildProcess 覆盖 issue #26：
// 命令超时被 SIGKILL 后，父进程必须回收子进程，否则会留下僵尸进程（<defunct>），
// 持续占用 PID 表。干净上游上此测试会观察到僵尸状态 "Z"。
func TestExecuteStreamingTimeoutReapsChildProcess(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")

	shell := local.NewLocalShell(&types.ShellConfig{Timeout: 300 * time.Millisecond})
	sr, err := shell.ExecuteStreaming(context.Background(), &filesystem.ExecuteRequest{
		Command: fmt.Sprintf("echo $$ > %s; sleep 30", pidFile),
	})
	if err != nil {
		t.Fatalf("ExecuteStreaming: %v", err)
	}
	drain(t, sr)

	assertChildReaped(t, pidFile)
}

// TestExecuteStreamingCancelReapsChildProcess 覆盖调用方取消 context 的分支：
// 与超时同样走 ctx.Err() 提前返回，必须回收子进程。
func TestExecuteStreamingCancelReapsChildProcess(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")

	ctx, cancel := context.WithCancel(context.Background())
	shell := local.NewLocalShell(nil)
	sr, err := shell.ExecuteStreaming(ctx, &filesystem.ExecuteRequest{
		Command: fmt.Sprintf("echo $$ > %s; sleep 30", pidFile),
	})
	if err != nil {
		t.Fatalf("ExecuteStreaming: %v", err)
	}
	// 等命令真正启动，再取消。
	for i := 0; i < 100; i++ {
		if _, statErr := os.Stat(pidFile); statErr == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	drain(t, sr)

	assertChildReaped(t, pidFile)
}
