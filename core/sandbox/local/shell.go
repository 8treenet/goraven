// Package local 中的 shell.go 实现本地 StreamingShell。
// 整体逻辑参照 github.com/cloudwego/eino-ext/adk/backend/local 的 ExecuteStreaming
// 链路移植而来，在此基础上额外支持通过 ShellConfig.Env 注入环境变量。
package local

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/8treenet/freedom"
	"github.com/cloudwego/eino/adk/filesystem"
	"github.com/cloudwego/eino/schema"

	"goraven/core/sandbox/types"
)

var _ types.StreamingShell = (*LocalShell)(nil)

// LocalShell 本地 shell 实现，相比 eino-ext 版本额外支持自定义环境变量和单命令超时。
type LocalShell struct {
	validateCommand func(string) error
	extraEnv        []string
	timeout         time.Duration
}

var defaultValidateCommand = func(string) error { return nil }

// NewLocalShell 根据 ShellConfig 构造本地 shell。
// ValidateCommand 为空时不做命令校验；Env 为 KEY=VALUE 形式的附加变量，
// 会在 os.Environ() 之上追加，后者中同名变量被覆盖（exec.Cmd 的语义）。
// Timeout 为单命令超时，0 表示不设超时。
func NewLocalShell(cfg *types.ShellConfig) *LocalShell {
	s := &LocalShell{validateCommand: defaultValidateCommand}
	if cfg == nil {
		return s
	}
	if cfg.ValidateCommand != nil {
		s.validateCommand = cfg.ValidateCommand
	}
	if len(cfg.Env) > 0 {
		s.extraEnv = append([]string(nil), cfg.Env...)
	}
	if cfg.Timeout > 0 {
		s.timeout = cfg.Timeout
	}
	return s
}

// ExecuteStreaming 执行命令并以流的形式返回 stdout，行为与 eino-ext 一致：
//   - 通过 /bin/sh -c 执行
//   - RunInBackendGround=true 时仅启动并立即返回，由 ctx 控制生命周期
//   - 否则按行流式返回 stdout，结束时附带 exitCode/stderr
func (s *LocalShell) ExecuteStreaming(ctx context.Context, input *filesystem.ExecuteRequest) (*schema.StreamReader[*filesystem.ExecuteResponse], error) {
	if input.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	freedom.Logger().Debugf("ExecuteStreaming :%v", *input)
	if err := s.validateCommand(input.Command); err != nil {
		return nil, err
	}

	if amended := ensureConnectTimeout(input.Command); amended != input.Command {
		freedom.Logger().Infof("[connect-timeout] inject: %q -> %q", input.Command, amended)
		input.Command = amended
	}

	execCtx := ctx
	var cmdCancel context.CancelFunc
	if !input.RunInBackendGround && s.timeout > 0 {
		execCtx, cmdCancel = context.WithTimeout(ctx, s.timeout)
		freedom.Logger().Infof("[shell timeout] command=%q, timeout=%v, runInBackground=%v", input.Command, s.timeout, input.RunInBackendGround)
	} else {
		freedom.Logger().Debugf("[shell no timeout] command=%q, timeout=%v, runInBackground=%v", input.Command, s.timeout, input.RunInBackendGround)
	}

	cmd, stdout, stderr, err := s.initStreamingCmd(execCtx, input.Command)
	if err != nil {
		if cmdCancel != nil {
			cmdCancel()
		}
		return nil, err
	}

	sr, w := schema.Pipe[*filesystem.ExecuteResponse](100)

	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		_ = stderr.Close()
		go sendErrorAndClose(w, fmt.Errorf("failed to start command: %w", err))
		if cmdCancel != nil {
			cmdCancel()
		}
		return sr, nil
	}

	if input.RunInBackendGround {
		if cmdCancel != nil {
			cmdCancel()
		}
		s.runCmdInBackground(execCtx, cmd, stdout, stderr, w)
		return sr, nil
	}

	go func() {
		if cmdCancel != nil {
			defer cmdCancel()
		}
		s.streamCmdOutput(execCtx, cmd, stdout, stderr, w)
	}()
	return sr, nil
}

// initStreamingCmd 创建命令并打开 stdout/stderr 管道，同时按需注入环境变量。
// 使用 Setpgid 创建进程组，并通过 cmd.Cancel 在 context 取消时杀掉整个进程组，
// 确保 curl|head 等管道命令的孙进程也能被正确终止。
func (s *LocalShell) initStreamingCmd(ctx context.Context, command string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return killProcessGroup(cmd)
	}
	cmd.Env = s.buildEnv()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdout.Close()
		return nil, nil, nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	return cmd, stdout, stderr, nil
}

// killProcessGroup 向 cmd 所属的进程组发送 SIGKILL，确保所有子进程（包括孙进程）都被终止。
func killProcessGroup(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

// connectTimeoutRule 描述一条"连接超时注入"规则：命令首 token 命中 cmd、
// 且未出现任何 detector（精确匹配或 "<detector>=" 前缀）时，在命令词后插入 inject。
// 仅限制 TCP 连接建立阶段，不影响数据传输耗时，因此下载大文件不会被误杀。
type connectTimeoutRule struct {
	cmd       string   // 命令首 token
	inject    string   // 待插入命令词后的参数
	detectors []string // 已设置超时的标志（精确或 "=" 前缀匹配）
}

// connectTimeoutRules 连接超时注入规则表。
// 国内访问被墙站点（如 google）时，curl/wget 会在 TCP 连接阶段长时间卡死。
// 在未显式设置连接超时时统一注入 5 秒上限，连接失败快速返回，下载/传输不受影响。
var connectTimeoutRules = []connectTimeoutRule{
	{cmd: "curl", inject: "--connect-timeout 5", detectors: []string{"--connect-timeout"}},
	{cmd: "wget", inject: "--connect-timeout 5", detectors: []string{"--connect-timeout"}},
}

// ensureConnectTimeout 按规则表为命令注入连接超时；命中规则并改写时返回新命令，否则原样返回。
func ensureConnectTimeout(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return command
	}
	for _, r := range connectTimeoutRules {
		if fields[0] != r.cmd {
			continue
		}
		if hasConnectTimeoutFlag(fields[1:], r.detectors) {
			return command
		}
		return insertAfterFirstToken(command, r.inject)
	}
	return command
}

// hasConnectTimeoutFlag 检查是否已显式设置连接超时：字段等于 detector，
// 或以 "<detector>=" 开头（如 --connect-timeout=10）。
func hasConnectTimeoutFlag(fields, detectors []string) bool {
	for _, f := range fields {
		for _, d := range detectors {
			if f == d || strings.HasPrefix(f, d+"=") {
				return true
			}
		}
	}
	return false
}

// insertAfterFirstToken 在命令首个 token 之后插入参数，保留前导空白与首个 token 之间的原始间隔。
func insertAfterFirstToken(command, inject string) string {
	lead := 0
	for lead < len(command) && (command[lead] == ' ' || command[lead] == '\t' || command[lead] == '\n') {
		lead++
	}
	rest := command[lead:]
	end := strings.IndexAny(rest, " \t\n")
	if end < 0 {
		return command + " " + inject
	}
	return command[:lead] + rest[:end] + " " + inject + rest[end:]
}

// buildEnv 返回子进程的环境变量。始终继承父进程环境；
// 若有 extraEnv 则追加其后（重复 key 以 extraEnv 为准）。
// 无 extraEnv 时返回 nil，让 exec.Cmd 自动继承。
func (s *LocalShell) buildEnv() []string {
	if len(s.extraEnv) == 0 {
		return nil
	}
	return append(os.Environ(), s.extraEnv...)
}

// runCmdInBackground 后台运行，仅发送一条 "command started in background" 后关闭流。
// 子进程的生命周期由 ctx 控制，context 取消时会通过 cmd.Cancel 杀掉进程组。
func (s *LocalShell) runCmdInBackground(ctx context.Context, cmd *exec.Cmd, stdout, stderr io.ReadCloser, w *schema.StreamWriter[*filesystem.ExecuteResponse]) {
	go func() {
		defer func() {
			if pe := recover(); pe != nil {
				_ = killProcessGroup(cmd)
			}
			_ = stdout.Close()
			_ = stderr.Close()
		}()

		done := make(chan struct{})
		go func() {
			drainPipesConcurrently(stdout, stderr)
			_ = cmd.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
			_ = killProcessGroup(cmd)
		}
	}()

	go func() {
		defer w.Close()
		w.Send(&filesystem.ExecuteResponse{Output: "command started in background\n", ExitCode: new(int)}, nil)
	}()
}

// drainPipesConcurrently 并发消费 stdout/stderr，防止管道阻塞导致子进程卡死。
func drainPipesConcurrently(stdout, stderr io.Reader) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(io.Discard, stdout)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(io.Discard, stderr)
	}()
	wg.Wait()
}

// streamCmdOutput 将命令输出按行推送给流的消费方。
func (s *LocalShell) streamCmdOutput(ctx context.Context, cmd *exec.Cmd, stdout, stderr io.ReadCloser, w *schema.StreamWriter[*filesystem.ExecuteResponse]) {
	defer func() {
		if pe := recover(); pe != nil {
			w.Send(nil, newPanicErr(pe, debug.Stack()))
			return
		}
		w.Close()
	}()

	stderrData, stderrErr := s.readStderrAsync(stderr)

	hasOutput, err := s.streamStdout(ctx, cmd, stdout, w)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			exitCode := 124
			timeoutMsg := "command timed out or was cancelled"
			if s.timeout > 0 {
				timeoutMsg = fmt.Sprintf("command timed out after %v", s.timeout)
			}
			w.Send(&filesystem.ExecuteResponse{Output: timeoutMsg + "\n", ExitCode: &exitCode}, nil)
			return
		}
		w.Send(nil, err)
		return
	}

	if stdError := <-stderrErr; stdError != nil {
		w.Send(nil, stdError)
		return
	}

	s.handleCmdCompletion(ctx, cmd, stderrData, hasOutput, w)
}

// readStderrAsync 异步读取 stderr，返回数据指针和完成信号通道。
func (s *LocalShell) readStderrAsync(stderr io.Reader) (*[]byte, <-chan error) {
	stderrData := new([]byte)
	stderrErr := make(chan error, 1)

	go func() {
		defer func() {
			if pe := recover(); pe != nil {
				stderrErr <- newPanicErr(pe, debug.Stack())
				return
			}
			close(stderrErr)
		}()
		var err error
		*stderrData, err = io.ReadAll(stderr)
		if err != nil {
			stderrErr <- fmt.Errorf("failed to read stderr: %w", err)
		}
	}()

	return stderrData, stderrErr
}

// streamStdout 按行读取 stdout 并通过 writer 发送。
// ReadString 在独立 goroutine 中执行，主循环通过 select 同时监听 ctx.Done()
// 和读取结果，确保 context 取消时能立即响应并杀掉进程组。
func (s *LocalShell) streamStdout(ctx context.Context, cmd *exec.Cmd, stdout io.Reader, w *schema.StreamWriter[*filesystem.ExecuteResponse]) (bool, error) {
	reader := bufio.NewReader(stdout)
	hasOutput := false

	type readResult struct {
		line string
		err  error
	}

	for {
		resultCh := make(chan readResult, 1)
		go func() {
			line, err := reader.ReadString('\n')
			resultCh <- readResult{line, err}
		}()

		select {
		case <-ctx.Done():
			_ = killProcessGroup(cmd)
			return hasOutput, ctx.Err()
		case r := <-resultCh:
			if r.line != "" {
				hasOutput = true
				select {
				case <-ctx.Done():
					_ = killProcessGroup(cmd)
					return hasOutput, ctx.Err()
				default:
					w.Send(&filesystem.ExecuteResponse{Output: r.line}, nil)
				}
			}
			if r.err != nil {
				if r.err != io.EOF {
					return hasOutput, fmt.Errorf("error reading stdout: %w", r.err)
				}
				return hasOutput, nil
			}
		}
	}
}

// handleCmdCompletion 等待命令结束并发送最终响应（含 exitCode 与 stderr 摘要）。
func (s *LocalShell) handleCmdCompletion(ctx context.Context, cmd *exec.Cmd, stderrData *[]byte, hasOutput bool, w *schema.StreamWriter[*filesystem.ExecuteResponse]) {
	if err := cmd.Wait(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode := exitError.ExitCode()
			parts := []string{fmt.Sprintf("command exited with non-zero code %d", exitCode)}
			if stderrStr := string(*stderrData); stderrStr != "" {
				parts = append(parts, "[stderr]:\n"+stderrStr)
			}
			w.Send(&filesystem.ExecuteResponse{
				Output:   strings.Join(parts, "\n"),
				ExitCode: &exitCode,
			}, nil)
			return
		}

		w.Send(nil, fmt.Errorf("command failed: %w", err))
		return
	}

	if !hasOutput {
		select {
		case <-ctx.Done():
			return
		default:
			w.Send(&filesystem.ExecuteResponse{ExitCode: new(int)}, nil)
		}
	}
}

// sendErrorAndClose 发送一个错误后关闭流。
func sendErrorAndClose(w *schema.StreamWriter[*filesystem.ExecuteResponse], err error) {
	defer w.Close()
	w.Send(nil, err)
}

type panicErr struct {
	info  any
	stack []byte
}

func (p *panicErr) Error() string {
	return fmt.Sprintf("panic error: %v, \nstack: %s", p.info, string(p.stack))
}

func newPanicErr(info any, stack []byte) error {
	return &panicErr{info: info, stack: stack}
}
