package git

import (
	"bytes"
	"context"
	"os/exec"
)

// Executor 抽象 git 命令执行，v1 为本机 CLI 实现，将来可替换为任务队列而不改上层调用。
type Executor interface {
	Run(ctx context.Context, dir string, env []string, args ...string) (stdout, stderr string, err error)
}

// CLIExecutor 使用 exec.CommandContext 参数数组执行 git，不经 shell。
type CLIExecutor struct{}

func (CLIExecutor) Run(ctx context.Context, dir string, env []string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// BinaryAvailable 检测宿主机是否可执行 git
func BinaryAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}
