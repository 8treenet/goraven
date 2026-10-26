package git

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

// Repo 封装对单一仓库工作目录的 git 命令操作。
// Env 为受控执行环境（见 BuildBaseEnv/BuildAuthEnv），Executor 可注入测试替身，
// Secrets 用于命令失败消息脱敏。
type Repo struct {
	Dir      string
	Env      []string
	Secrets  []string
	Executor Executor
}

func (r *Repo) exec() Executor {
	if r.Executor == nil {
		return CLIExecutor{}
	}
	return r.Executor
}

// run 执行 git 命令；未显式设置超时时套用默认操作超时。
func (r *Repo) run(ctx context.Context, args ...string) (string, string, error) {
	return r.runEnv(ctx, r.Env, args...)
}

func (r *Repo) runEnv(ctx context.Context, env []string, args ...string) (string, string, error) {
	if !BinaryAvailable() {
		return "", "", ErrBinaryMissing
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, OpTimeout)
		defer cancel()
	}
	return r.exec().Run(ctx, r.Dir, env, args...)
}

// cmdError 将 git 失败归类为平台错误并脱敏。
func (r *Repo) cmdError(stderr string, err error) error {
	if IsUnrelatedHistoryError(stderr) {
		return ErrUnrelatedHistory
	}
	if IsAuthError(stderr) {
		return ErrAuthFailed
	}
	msg := SanitizeMessage(strings.TrimSpace(stderr), r.Secrets...)
	if msg == "" && err != nil {
		msg = SanitizeMessage(err.Error(), r.Secrets...)
	}
	if msg == "" {
		msg = "unknown git error"
	}
	return &CommandError{Stderr: msg, Err: err}
}

// CollectChanges 读取工作区变更列表。
func (r *Repo) CollectChanges(ctx context.Context) ([]Change, error) {
	stdout, stderr, err := r.run(ctx, "status", "--porcelain", "-z", "-uall")
	if err != nil {
		return nil, r.cmdError(stderr, err)
	}
	return ParseStatusZ(stdout), nil
}

// CurrentBranch 当前分支名，未出生分支返回空串。
func (r *Repo) CurrentBranch(ctx context.Context) string {
	stdout, _, err := r.run(ctx, "branch", "--show-current")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(stdout)
}

// UpstreamName 上游引用名，未建立 upstream 返回空串。
func (r *Repo) UpstreamName(ctx context.Context) string {
	stdout, _, err := r.run(ctx, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(stdout)
}

// AheadBehind 本地相对 upstream 的领先/落后提交数，命令失败返回 0,0。
func (r *Repo) AheadBehind(ctx context.Context, upstream string) (int, int) {
	stdout, _, err := r.run(ctx, "rev-list", "--left-right", "--count", "HEAD..."+upstream)
	if err != nil {
		return 0, 0
	}
	fields := strings.Fields(stdout)
	if len(fields) != 2 {
		return 0, 0
	}
	ahead, _ := strconv.Atoi(fields[0])
	behind, _ := strconv.Atoi(fields[1])
	return ahead, behind
}

// CheckRemoteHistory fetch 后检查远程分支是否存在、是否与本地历史无关。
func (r *Repo) CheckRemoteHistory(ctx context.Context, branch string) (remoteExists, unrelated bool) {
	if _, _, err := r.run(ctx, "fetch", "origin"); err != nil {
		return false, false
	}
	if _, _, err := r.run(ctx, "rev-parse", "--verify", "refs/remotes/origin/"+branch); err != nil {
		return false, false
	}
	if _, _, err := r.run(ctx, "merge-base", "HEAD", "origin/"+branch); err != nil {
		return true, true
	}
	return true, false
}

// HasHead 判断仓库是否已有提交。
func (r *Repo) HasHead(ctx context.Context) bool {
	_, _, err := r.run(ctx, "rev-parse", "--verify", "HEAD")
	return err == nil
}

// setOriginURL 设置/更新 origin 地址。
func (r *Repo) setOriginURL(ctx context.Context, url string) error {
	if _, _, err := r.run(ctx, "remote", "get-url", "origin"); err == nil {
		if _, stderr, err := r.run(ctx, "remote", "set-url", "origin", url); err != nil {
			return r.cmdError(stderr, err)
		}
		return nil
	}
	if _, stderr, err := r.run(ctx, "remote", "add", "origin", url); err != nil {
		return r.cmdError(stderr, err)
	}
	return nil
}

// CommitWorkingTree staging 过滤大文件后提交，返回提交 hash 与被跳过的文件。
func (r *Repo) CommitWorkingTree(ctx context.Context, identity Identity, message string) (string, []Change, error) {
	changes, err := r.CollectChanges(ctx)
	if err != nil {
		return "", nil, err
	}
	if len(changes) == 0 {
		return "", nil, ErrNoChanges
	}
	keep, skipped := FilterLargeChanges(r.Dir, changes, LargeFileLimit)
	if len(keep) == 0 {
		return "", skipped, ErrNoChanges
	}

	paths := make([]string, 0, len(keep)*2)
	for _, change := range keep {
		paths = append(paths, change.Path)
		if change.Source != "" {
			paths = append(paths, change.Source)
		}
	}
	if _, stderr, err := r.run(ctx, append([]string{"add", "--"}, paths...)...); err != nil {
		return "", nil, r.cmdError(stderr, err)
	}

	commitEnv := WithEnv(r.Env,
		"GIT_AUTHOR_NAME", identity.AuthorName,
		"GIT_AUTHOR_EMAIL", identity.AuthorEmail,
		"GIT_COMMITTER_NAME", identity.CommitterName,
		"GIT_COMMITTER_EMAIL", identity.CommitterEmail,
	)
	if _, stderr, err := r.runEnv(ctx, commitEnv, "commit", "-m", message); err != nil {
		return "", nil, r.cmdError(stderr, err)
	}
	stdout, _, err := r.run(ctx, "rev-parse", "HEAD")
	if err != nil {
		return "", skipped, nil
	}
	return strings.TrimSpace(stdout), skipped, nil
}

// Push 推送（含首次建立 upstream 与非快进恢复），永不强推。
func (r *Repo) Push(ctx context.Context, remoteURL string) error {
	if strings.TrimSpace(remoteURL) == "" {
		return ErrNotConfigured
	}
	branch := r.CurrentBranch(ctx)
	if branch == "" {
		return ErrNoChanges
	}
	upstream := r.UpstreamName(ctx)
	pushArgs := []string{"push"}
	if upstream == "" {
		remoteExists, unrelated := r.CheckRemoteHistory(ctx, branch)
		if remoteExists && unrelated {
			return ErrUnrelatedHistory
		}
		pushArgs = []string{"push", "-u", "origin", branch}
	}
	_, stderr, err := r.run(ctx, pushArgs...)
	if err == nil {
		return nil
	}
	return r.recoverPushRejection(ctx, pushArgs, branch, upstream, stderr, err)
}

// recoverPushRejection push 被拒后的自动恢复：pull --rebase 重试一次，绝不强推。
func (r *Repo) recoverPushRejection(ctx context.Context, pushArgs []string, branch, upstream, stderr string, runErr error) error {
	if IsUnrelatedHistoryError(stderr) {
		return ErrUnrelatedHistory
	}
	if !IsPushRejected(stderr) {
		return r.cmdError(stderr, runErr)
	}
	if changes, err := r.CollectChanges(ctx); err == nil && len(changes) > 0 {
		return ErrDirtyTree
	}

	pullArgs := []string{"pull", "--rebase"}
	if upstream == "" {
		pullArgs = append(pullArgs, "origin", branch)
	}
	_, rebaseStderr, rebaseErr := r.run(ctx, pullArgs...)
	if rebaseErr != nil {
		// 冲突时恢复原状，绝不留半途 rebase
		_, _, _ = r.run(ctx, "rebase", "--abort")
		if IsUnrelatedHistoryError(rebaseStderr) {
			return ErrUnrelatedHistory
		}
		if IsConflict(rebaseStderr) {
			return ErrConflict
		}
		return r.cmdError(rebaseStderr, rebaseErr)
	}

	if _, retryStderr, retryErr := r.run(ctx, pushArgs...); retryErr != nil {
		if IsPushRejected(retryStderr) || IsUnrelatedHistoryError(retryStderr) {
			return ErrPushRejected
		}
		return r.cmdError(retryStderr, retryErr)
	}
	return nil
}

// Pull 要求工作区干净，执行 fetch + pull --rebase。
func (r *Repo) Pull(ctx context.Context, remoteURL string) error {
	if strings.TrimSpace(remoteURL) == "" {
		return ErrNotConfigured
	}
	changes, err := r.CollectChanges(ctx)
	if err != nil {
		return err
	}
	if len(changes) > 0 {
		return ErrDirtyTree
	}
	branch := r.CurrentBranch(ctx)
	if branch == "" {
		return ErrNoChanges
	}
	upstream := r.UpstreamName(ctx)
	pullArgs := []string{"pull", "--rebase"}
	if upstream == "" {
		remoteExists, unrelated := r.CheckRemoteHistory(ctx, branch)
		if !remoteExists {
			return nil
		}
		if unrelated {
			return ErrUnrelatedHistory
		}
		pullArgs = append(pullArgs, "origin", branch)
	}

	_, stderr, err := r.run(ctx, pullArgs...)
	if err == nil {
		return nil
	}
	if IsUnrelatedHistoryError(stderr) {
		return ErrUnrelatedHistory
	}
	if IsConflict(stderr) {
		_, _, _ = r.run(ctx, "rebase", "--abort")
		return ErrConflict
	}
	return r.cmdError(stderr, err)
}

// MergeUnrelated 历史无关时的用户选择：pull --allow-unrelated-histories --no-rebase 合并，冲突时 abort。
func (r *Repo) MergeUnrelated(ctx context.Context, branch string) error {
	pullArgs := []string{"pull", "--allow-unrelated-histories", "--no-rebase"}
	if r.UpstreamName(ctx) == "" {
		pullArgs = append(pullArgs, "origin", branch)
	}
	_, stderr, err := r.run(ctx, pullArgs...)
	if err != nil {
		_, _, _ = r.run(ctx, "merge", "--abort")
		if IsConflict(stderr) {
			return ErrConflict
		}
		return r.cmdError(stderr, err)
	}
	return nil
}

// Initialize 初始化仓库：未初始化时 init + 默认 .gitignore + 首次提交，已有 .git 直接采纳。
// remoteURL 非空时设置/更新 origin；为空时回读已采纳仓库的 origin 地址返回（由调用方回填配置）。
func (r *Repo) Initialize(ctx context.Context, remoteURL string, identity Identity) (string, error) {
	auth, err := BuildBaseEnv()
	if err != nil {
		return "", err
	}
	defer auth.Cleanup()
	inner := &Repo{Dir: r.Dir, Env: auth.Env, Executor: r.Executor, Secrets: r.Secrets}

	if !IsRepository(r.Dir) {
		if _, stderr, err := inner.run(ctx, "init", "-b", "main"); err != nil {
			return "", inner.cmdError(stderr, err)
		}
		if _, err := EnsureDefaultGitignore(r.Dir); err != nil {
			return "", err
		}
		if _, _, err := inner.CommitWorkingTree(ctx, identity, InitCommitMessage); err != nil && !errors.Is(err, ErrNoChanges) {
			return "", err
		}
	}

	if strings.TrimSpace(remoteURL) == "" {
		// 采纳已有仓库时回填 remote 配置
		if stdout, _, err := inner.run(ctx, "remote", "get-url", "origin"); err == nil {
			url := strings.TrimSpace(stdout)
			if ValidateRemoteURL(url) == nil {
				return url, nil
			}
		}
		return "", nil
	}
	return "", inner.setOriginURL(ctx, remoteURL)
}

// Clone 克隆远程仓库到 Dir（目录需已存在）；克隆后重置 origin 地址确保无凭据残留。
func (r *Repo) Clone(ctx context.Context, url string, shallow bool) error {
	args := []string{"clone"}
	if shallow {
		args = append(args, "--depth", "1")
	}
	args = append(args, "--", url, ".")
	_, stderr, err := r.run(ctx, args...)
	if err != nil {
		return r.cmdError(stderr, err)
	}
	// 确保 remote 地址无凭据残留
	_, _, _ = r.run(ctx, "remote", "set-url", "origin", url)
	return nil
}

// LsRemoteHead 执行 ls-remote 并解析远端 HEAD 提交。
func (r *Repo) LsRemoteHead(ctx context.Context, url string) (string, error) {
	stdout, stderr, err := r.run(ctx, "ls-remote", "--", url, "HEAD")
	if err != nil {
		return "", r.cmdError(stderr, err)
	}
	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "HEAD" {
			return fields[0], nil
		}
	}
	return "", nil
}

// Log 提交历史，limit 上限 100；仓库尚无提交（unborn HEAD）时返回空历史。
func (r *Repo) Log(ctx context.Context, limit, offset int) ([]CommitInfo, error) {
	if !r.HasHead(ctx) {
		return []CommitInfo{}, nil
	}
	stdout, stderr, err := r.run(ctx,
		"log", "--pretty=format:"+LogFormat,
		"-n", strconv.Itoa(limit), "--skip", strconv.Itoa(offset))
	if err != nil {
		return nil, r.cmdError(stderr, err)
	}
	return ParseLog(stdout), nil
}

// Diff 无 commit 时为工作区 vs HEAD，否则展示指定提交的 patch；超出 maxChars 截断。
func (r *Repo) Diff(ctx context.Context, path, commit string, maxChars int) (string, error) {
	var args []string
	if commit == "" {
		args = []string{"diff"}
		if r.HasHead(ctx) {
			args = append(args, "HEAD")
		}
	} else {
		args = []string{"show", "--format=", "--patch", commit}
	}
	args = append(args, "--")
	if path != "" {
		args = append(args, path)
	}
	stdout, stderr, err := r.run(ctx, args...)
	if err != nil {
		return "", r.cmdError(stderr, err)
	}
	if len(stdout) > maxChars {
		stdout = stdout[:maxChars]
	}
	return stdout, nil
}
