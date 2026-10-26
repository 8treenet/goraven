package service

// 项目 Git 集成（个人项目与团队项目共用）：
// 所有远程操作由后端代执行，凭据只写不回读；并发通过 Redis 锁串行化。
// 本文件是 FileManagerService 基类的 Git 方法集（文件处理的基类），
// 由 MyProjectService/TeamProjectService 通过嵌入继承对外暴露；
// 纯 git 命令执行与环境隔离下沉在 util/git 包。

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"goraven/backend/po"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/util/git"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

const (
	gitLockTTL      = 10 * time.Minute // 手动/定时操作锁 TTL
	gitCloneLockTTL = 35 * time.Minute // 克隆锁 TTL（覆盖克隆超时）
)

var gitCommitHashRe = regexp.MustCompile(`^[0-9a-fA-F]{4,40}$`)

// GitContext 已解析的项目 Git 上下文：由个人/团队项目服务构建，
// 携带项目物理目录与权限角色，供基类 Git 方法使用。
type GitContext struct {
	UserId      string
	OwnerType   uint8
	OwnerId     int
	ProjectName string
	Dir         string
	IsOwner     bool // 个人项目 owner
	IsCreator   bool // 团队项目创建者
	IsMember    bool // 团队项目成员
}

// authorize 权限矩阵：个人 owner 全通；团队 creator 全通、member 可读可写、非 member 拒绝。
func (ctx *GitContext) authorize(needConfig bool) error {
	if ctx.OwnerType == po.GitOwnerUserProject {
		if !ctx.IsOwner {
			return errs.ErrGitPermission
		}
		return nil
	}
	if ctx.IsCreator {
		return nil
	}
	if ctx.IsMember && !needConfig {
		return nil
	}
	return errs.ErrGitPermission
}

// --- 级联清理（删除项目时调用） ---

// gitOwnerKey 项目 Git 克隆取消注册表的键。
func gitOwnerKey(ownerType uint8, ownerId int) string {
	return fmt.Sprintf("%d:%d", ownerType, ownerId)
}

// cancelGitClone 取消进行中的克隆（删除项目时先调用，行删除后 goroutine 不再回写）。
func cancelGitClone(ownerType uint8, ownerId int) {
	git.CancelClone(gitOwnerKey(ownerType, ownerId))
}

// cleanupGitSecrets 删除项目级持久化凭据文件（known_hosts），删除项目时级联调用。
func cleanupGitSecrets(ownerType uint8, ownerId int) {
	git.RemoveKnownHosts(config.Get().GetGitSecretDir(), ownerType, ownerId)
}

// gitCleanupCascade 删除项目时级联清理 Git：取消克隆 + 删除配置 + 清理凭据文件。
func (service *FileManagerService) gitCleanupCascade(ownerType uint8, ownerId int) error {
	cancelGitClone(ownerType, ownerId)
	if err := service.GitSettingRepo.DeleteByOwner(ownerType, ownerId); err != nil {
		return err
	}
	cleanupGitSecrets(ownerType, ownerId)
	return nil
}

// --- 转换与校验工具 ---

// toGitAuthSetting po 配置转 util/git 认证参数。
func toGitAuthSetting(setting *po.ProjectGitSetting) *git.AuthSetting {
	return &git.AuthSetting{
		AuthType:      setting.AuthType,
		SshPrivateKey: setting.SshPrivateKey,
		HttpsUsername: setting.HttpsUsername,
		HttpsSecret:   setting.HttpsSecret,
		OwnerType:     setting.OwnerType,
		OwnerId:       setting.OwnerId,
	}
}

// defaultProjectGitSetting 新项目的默认 Git 配置行（存在即代表是 Git 项目）。
func defaultProjectGitSetting(ownerType uint8, ownerId int) *po.ProjectGitSetting {
	return &po.ProjectGitSetting{
		OwnerType: ownerType,
		OwnerId:   ownerId,
	}
}

// applyGitTestReq 测试连接参数覆盖配置（不落库）。
func applyGitTestReq(setting *po.ProjectGitSetting, req *vo.GitTestReq) {
	if req == nil {
		return
	}
	if strings.TrimSpace(req.RemoteUrl) != "" {
		setting.RemoteUrl = strings.TrimSpace(req.RemoteUrl)
	}
	if req.AuthType != nil {
		setting.AuthType = *req.AuthType
	}
	if req.SshPrivateKey != nil {
		setting.SshPrivateKey = *req.SshPrivateKey
	}
	if req.HttpsUsername != nil {
		setting.HttpsUsername = strings.TrimSpace(*req.HttpsUsername)
	}
	if req.HttpsSecret != nil {
		setting.HttpsSecret = *req.HttpsSecret
	}
}

// validateGitRemoteUrl 校验远程地址，映射为平台错误。
func validateGitRemoteUrl(raw string) error {
	if err := git.ValidateRemoteURL(raw); err != nil {
		return errs.ErrGitInvalidRemoteUrl
	}
	return nil
}

func validateGitSetting(setting *po.ProjectGitSetting) error {
	switch setting.AuthType {
	case po.GitAuthNone, po.GitAuthSSH, po.GitAuthHTTPS:
	default:
		return errs.ErrGitNotConfigured
	}
	if strings.TrimSpace(setting.RemoteUrl) != "" {
		return validateGitRemoteUrl(setting.RemoteUrl)
	}
	return nil
}

func hasGitCredential(setting *po.ProjectGitSetting) bool {
	switch setting.AuthType {
	case po.GitAuthSSH:
		return strings.TrimSpace(setting.SshPrivateKey) != ""
	case po.GitAuthHTTPS:
		return strings.TrimSpace(setting.HttpsSecret) != ""
	}
	return false
}

// gitOperationError 将 util/git 错误映射为平台错误。
func gitOperationError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, git.ErrNoChanges):
		return errs.ErrGitNoChanges
	case errors.Is(err, git.ErrDirtyTree):
		return errs.ErrGitDirtyTree
	case errors.Is(err, git.ErrNotConfigured):
		return errs.ErrGitNotConfigured
	case errors.Is(err, git.ErrAuthFailed):
		return errs.ErrGitAuthFailed
	case errors.Is(err, git.ErrPushRejected):
		return errs.ErrGitPushRejected
	case errors.Is(err, git.ErrUnrelatedHistory):
		return errs.ErrGitUnrelatedHistory
	case errors.Is(err, git.ErrConflict):
		return errs.ErrGitNeedsAttention
	case errors.Is(err, git.ErrBinaryMissing):
		return errs.ErrGitBinaryMissing
	case errors.Is(err, git.ErrInvalidRemoteURL):
		return errs.ErrGitInvalidRemoteUrl
	}
	var cmdErr *git.CommandError
	if errors.As(err, &cmdErr) {
		msg := cmdErr.Stderr
		if msg == "" {
			msg = "unknown git error"
		}
		return errs.NewFormatError("git command failed: %s", "Git 命令执行失败：%s", msg)
	}
	return err
}

// gitChangesToVO 转换变更列表。
func gitChangesToVO(changes []git.Change) []vo.GitChange {
	items := make([]vo.GitChange, 0, len(changes))
	for _, change := range changes {
		items = append(items, vo.GitChange{Path: change.Path, Status: change.Status, Size: change.Size})
	}
	return items
}

// gitCommitsToVO 转换提交历史。
func gitCommitsToVO(items []git.CommitInfo) []vo.GitCommitInfo {
	out := make([]vo.GitCommitInfo, 0, len(items))
	for _, item := range items {
		out = append(out, vo.GitCommitInfo{
			Hash:      item.Hash,
			ShortHash: item.ShortHash,
			Author:    item.Author,
			Email:     item.Email,
			Message:   item.Message,
			Time:      item.Time,
		})
	}
	return out
}

// --- 基础设施（配置读取 / 锁 / 身份 / 仓库句柄） ---

// loadSetting 读取项目 Git 配置；不存在返回 nil。
func (service *FileManagerService) loadSetting(ownerType uint8, ownerId int) (*po.ProjectGitSetting, error) {
	setting, err := service.GitSettingRepo.GetByOwner(ownerType, ownerId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return setting, nil
}

// ensureWritable Preview 模式下拒绝所有写操作。
func (service *FileManagerService) ensureWritable() error {
	if config.Get().Behavior.PreviewUser != "" {
		return errs.ErrGitPreviewReadOnly
	}
	return nil
}

// lockOwner 获取项目 Git 操作锁，失败返回 ErrGitBusy。
func (service *FileManagerService) lockOwner(ownerType uint8, ownerId int, ttl time.Duration, token string) (func(), error) {
	ok, err := service.GitSettingRepo.LockOwner(ownerType, ownerId, token, ttl)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errs.ErrGitBusy
	}
	return func() {
		_ = service.GitSettingRepo.UnlockOwner(ownerType, ownerId)
	}, nil
}

// requireGitSetting 校验项目是 Git 项目（存在配置行）且未在克隆中。
func (service *FileManagerService) requireGitSetting(ownerType uint8, ownerId int) (*po.ProjectGitSetting, error) {
	setting, err := service.loadSetting(ownerType, ownerId)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, errs.ErrGitDisabled
	}
	if setting.CloneState == po.GitCloneRunning {
		return nil, errs.ErrGitCloning
	}
	return setting, nil
}

// manualIdentity 手动操作提交身份（author 为操作者，committer 为机器人）。
func (service *FileManagerService) manualIdentity(userId string) git.Identity {
	name := userId
	if service.UserRepo != nil {
		if user, err := service.UserRepo.FindByUserId(userId); err == nil {
			switch {
			case user.Nickname != "":
				name = user.Nickname
			case user.Username != "":
				name = user.Username
			}
		}
	}
	return git.Identity{
		AuthorName:     name,
		AuthorEmail:    userId + "@goraven.dev",
		CommitterName:  git.BotIdentity.CommitterName,
		CommitterEmail: git.BotIdentity.CommitterEmail,
	}
}

// newGitRepo 构造面向项目目录的仓库操作句柄。
func (service *FileManagerService) newGitRepo(dir string, setting *po.ProjectGitSetting, env []string) *git.Repo {
	repo := &git.Repo{Dir: dir, Env: env}
	if setting != nil {
		repo.Secrets = []string{setting.HttpsSecret, setting.SshPrivateKey}
	}
	return repo
}

// resolveOwnerDir 按 owner 解析项目物理目录（定时同步等无用户场景使用）。
func (service *FileManagerService) resolveOwnerDir(ownerType uint8, ownerId int) (string, string, error) {
	switch ownerType {
	case po.GitOwnerUserProject:
		project, err := service.UserProjectRepo.GetByID(ownerId)
		if err != nil {
			return "", "", errs.ErrUserProjectNotFound
		}
		username, err := service.UserProjectRepo.GetUsernameByUserID(project.UserId)
		if err != nil {
			return "", "", err
		}
		return filepath.Join(config.Get().GetUserSpace(username), "projects", project.ProjectName), project.ProjectName, nil
	case po.GitOwnerTeamProject:
		project, err := service.TeamProjectRepo.GetByID(ownerId)
		if err != nil {
			return "", "", errs.ErrTeamProjectNotFound
		}
		return filepath.Join(config.Get().GetTeamProjectDir(), project.ProjectName), project.ProjectName, nil
	}
	return "", "", errs.ErrGitPermission
}

// initializeRepository 开启 Git 时初始化：init + 默认 .gitignore + 首次提交，已有 .git 直接采纳。
func (service *FileManagerService) initializeRepository(ctx context.Context, gctx *GitContext, setting *po.ProjectGitSetting) error {
	repo := service.newGitRepo(gctx.Dir, nil, nil)
	backfill, err := repo.Initialize(ctx, setting.RemoteUrl, git.BotIdentity)
	if err != nil {
		return gitOperationError(err)
	}
	if backfill != "" {
		setting.RemoteUrl = backfill
	}
	return nil
}

// --- 状态与配置 API ---

// gitStatus 状态聚合：配置、远程、分支、ahead/behind、变更列表与同步时间。
func (service *FileManagerService) gitStatus(gctx *GitContext) (*vo.GitStatusRsp, error) {
	if err := gctx.authorize(false); err != nil {
		return nil, err
	}
	setting, err := service.loadSetting(gctx.OwnerType, gctx.OwnerId)
	if err != nil {
		return nil, err
	}
	rsp := &vo.GitStatusRsp{Changes: []vo.GitChange{}, Skipped: []vo.GitChange{}}
	if setting == nil {
		return rsp, nil
	}
	service.fillStatusFromSetting(rsp, setting)
	rsp.Initialized = git.IsRepository(gctx.Dir)
	if setting.CloneState == po.GitCloneRunning || !rsp.Initialized {
		return rsp, nil
	}

	env, err := git.BuildAuthEnv(config.Get().GetGitSecretDir(), toGitAuthSetting(setting))
	if err != nil {
		// 凭据暂缺不影响状态读取
		return rsp, nil
	}
	defer env.Cleanup()

	ctx := context.Background()
	repo := service.newGitRepo(git.DirOrTemp(gctx.Dir), setting, env.Env)
	changes, err := repo.CollectChanges(ctx)
	if err != nil {
		return rsp, nil
	}
	keep, skipped := git.FilterLargeChanges(gctx.Dir, changes, git.LargeFileLimit)
	if len(keep) > git.StatusMaxChanges {
		keep = keep[:git.StatusMaxChanges]
	}
	rsp.Changes = gitChangesToVO(keep)
	rsp.Skipped = gitChangesToVO(skipped)
	rsp.Branch = repo.CurrentBranch(ctx)
	if upstream := repo.UpstreamName(ctx); upstream != "" {
		rsp.Ahead, rsp.Behind = repo.AheadBehind(ctx, upstream)
	}
	return rsp, nil
}

func (service *FileManagerService) fillStatusFromSetting(rsp *vo.GitStatusRsp, setting *po.ProjectGitSetting) {
	rsp.Enabled = true
	rsp.CloneState = setting.CloneState
	rsp.CloneMessage = git.SanitizeMessage(setting.CloneMessage, setting.HttpsSecret, setting.SshPrivateKey)
	rsp.Remote = vo.GitRemoteInfo{
		Url:           setting.RemoteUrl,
		AuthType:      setting.AuthType,
		HttpsUsername: setting.HttpsUsername,
		HasCredential: hasGitCredential(setting),
	}
}

// gitTestRemoteDirect 新建项目前的测试连接（不建项目、不落库），供创建对话框使用。
func (service *FileManagerService) gitTestRemoteDirect(req *vo.GitTestReq) (*vo.GitTestRsp, error) {
	if req == nil || strings.TrimSpace(req.RemoteUrl) == "" {
		return nil, errs.ErrGitNotConfigured
	}
	setting := defaultProjectGitSetting(po.GitOwnerUserProject, 0)
	applyGitTestReq(setting, req)
	return service.lsRemoteHead(setting, os.TempDir())
}

// lsRemoteHead 执行 ls-remote 并解析远端 HEAD。
func (service *FileManagerService) lsRemoteHead(setting *po.ProjectGitSetting, dir string) (*vo.GitTestRsp, error) {
	if err := validateGitSetting(setting); err != nil {
		return nil, err
	}
	if strings.TrimSpace(setting.RemoteUrl) == "" {
		return nil, errs.ErrGitNotConfigured
	}
	env, err := git.BuildAuthEnv(config.Get().GetGitSecretDir(), toGitAuthSetting(setting))
	if err != nil {
		return nil, gitOperationError(err)
	}
	defer env.Cleanup()

	repo := service.newGitRepo(dir, setting, env.Env)
	head, err := repo.LsRemoteHead(context.Background(), setting.RemoteUrl)
	if err != nil {
		return nil, gitOperationError(err)
	}
	return &vo.GitTestRsp{Head: head}, nil
}

// --- 手动提交 / 推送 / 拉取 ---

// gitCommit 手动提交；push=true 时提交后直接走推送流程。
func (service *FileManagerService) gitCommit(gctx *GitContext, req *vo.GitCommitReq) (*vo.GitCommitRsp, error) {
	if err := service.ensureWritable(); err != nil {
		return nil, err
	}
	if err := gctx.authorize(false); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errs.ErrGitMessageRequired
	}
	message := strings.TrimSpace(req.Message)
	if message == "" || utf8.RuneCountInString(message) > 500 {
		return nil, errs.ErrGitMessageRequired
	}
	unlock, err := service.lockOwner(gctx.OwnerType, gctx.OwnerId, gitLockTTL, gctx.UserId)
	if err != nil {
		return nil, err
	}
	defer unlock()

	setting, err := service.requireGitSetting(gctx.OwnerType, gctx.OwnerId)
	if err != nil {
		return nil, err
	}
	if !git.IsRepository(gctx.Dir) {
		return nil, errs.ErrGitDisabled
	}
	env, err := git.BuildAuthEnv(config.Get().GetGitSecretDir(), toGitAuthSetting(setting))
	if err != nil {
		return nil, gitOperationError(err)
	}
	defer env.Cleanup()

	ctx := context.Background()
	repo := service.newGitRepo(gctx.Dir, setting, env.Env)
	hash, skipped, err := repo.CommitWorkingTree(ctx, service.manualIdentity(gctx.UserId), message)
	if err != nil {
		return nil, gitOperationError(err)
	}
	rsp := &vo.GitCommitRsp{Hash: hash, Skipped: gitChangesToVO(skipped)}
	if req.Push {
		if err := repo.Push(ctx, setting.RemoteUrl); err != nil {
			return nil, gitOperationError(err)
		}
		rsp.Pushed = true
	}
	return rsp, nil
}

// gitPush 执行推送流程（含非快进处理），永不强推。
func (service *FileManagerService) gitPush(gctx *GitContext) error {
	if err := service.ensureWritable(); err != nil {
		return err
	}
	if err := gctx.authorize(false); err != nil {
		return err
	}
	unlock, err := service.lockOwner(gctx.OwnerType, gctx.OwnerId, gitLockTTL, gctx.UserId)
	if err != nil {
		return err
	}
	defer unlock()

	setting, err := service.requireGitSetting(gctx.OwnerType, gctx.OwnerId)
	if err != nil {
		return err
	}
	if strings.TrimSpace(setting.RemoteUrl) == "" {
		return errs.ErrGitNotConfigured
	}
	env, err := git.BuildAuthEnv(config.Get().GetGitSecretDir(), toGitAuthSetting(setting))
	if err != nil {
		return gitOperationError(err)
	}
	defer env.Cleanup()

	repo := service.newGitRepo(gctx.Dir, setting, env.Env)
	return gitOperationError(repo.Push(context.Background(), setting.RemoteUrl))
}

// gitPull 要求工作区干净，执行 fetch + pull --rebase。
func (service *FileManagerService) gitPull(gctx *GitContext) error {
	if err := service.ensureWritable(); err != nil {
		return err
	}
	if err := gctx.authorize(false); err != nil {
		return err
	}
	unlock, err := service.lockOwner(gctx.OwnerType, gctx.OwnerId, gitLockTTL, gctx.UserId)
	if err != nil {
		return err
	}
	defer unlock()

	setting, err := service.requireGitSetting(gctx.OwnerType, gctx.OwnerId)
	if err != nil {
		return err
	}
	if strings.TrimSpace(setting.RemoteUrl) == "" {
		return errs.ErrGitNotConfigured
	}
	env, err := git.BuildAuthEnv(config.Get().GetGitSecretDir(), toGitAuthSetting(setting))
	if err != nil {
		return gitOperationError(err)
	}
	defer env.Cleanup()

	repo := service.newGitRepo(gctx.Dir, setting, env.Env)
	return gitOperationError(repo.Pull(context.Background(), setting.RemoteUrl))
}

// gitResolveUnrelated 历史无关时的用户选择：merge 合并后推送，local_only 仅保留远程配置。
func (service *FileManagerService) gitResolveUnrelated(gctx *GitContext, req *vo.GitUnrelatedReq) error {
	if err := service.ensureWritable(); err != nil {
		return err
	}
	action := ""
	if req != nil {
		action = strings.TrimSpace(req.Action)
	}
	if action != "merge" && action != "local_only" {
		return errs.NewFormatError("action must be 'merge' or 'local_only'", "action 必须为 'merge' 或 'local_only'")
	}
	if err := gctx.authorize(false); err != nil {
		return err
	}
	unlock, err := service.lockOwner(gctx.OwnerType, gctx.OwnerId, gitLockTTL, gctx.UserId)
	if err != nil {
		return err
	}
	defer unlock()

	setting, err := service.requireGitSetting(gctx.OwnerType, gctx.OwnerId)
	if err != nil {
		return err
	}
	if action == "local_only" {
		return nil
	}
	if strings.TrimSpace(setting.RemoteUrl) == "" {
		return errs.ErrGitNotConfigured
	}
	env, err := git.BuildAuthEnv(config.Get().GetGitSecretDir(), toGitAuthSetting(setting))
	if err != nil {
		return gitOperationError(err)
	}
	defer env.Cleanup()

	ctx := context.Background()
	repo := service.newGitRepo(gctx.Dir, setting, env.Env)
	changes, err := repo.CollectChanges(ctx)
	if err != nil {
		return gitOperationError(err)
	}
	if len(changes) > 0 {
		return errs.ErrGitDirtyTree
	}
	branch := repo.CurrentBranch(ctx)
	if branch == "" {
		return errs.ErrGitNoChanges
	}
	if err := repo.MergeUnrelated(ctx, branch); err != nil {
		return gitOperationError(err)
	}
	if err := repo.Push(ctx, setting.RemoteUrl); err != nil {
		return gitOperationError(err)
	}
	return nil
}

// --- 历史与 diff ---

// gitLog 提交历史，limit 默认 20、最大 100。
func (service *FileManagerService) gitLog(gctx *GitContext, req *vo.GitLogReq) (*vo.GitLogRsp, error) {
	if err := gctx.authorize(false); err != nil {
		return nil, err
	}
	if _, err := service.requireGitSetting(gctx.OwnerType, gctx.OwnerId); err != nil {
		return nil, err
	}
	if !git.IsRepository(gctx.Dir) {
		return &vo.GitLogRsp{Items: []vo.GitCommitInfo{}}, nil
	}
	limit, offset := 20, 0
	if req != nil {
		if req.Limit > 0 {
			limit = req.Limit
		}
		if req.Offset > 0 {
			offset = req.Offset
		}
	}
	if limit > 100 {
		limit = 100
	}

	env, err := git.BuildBaseEnv()
	if err != nil {
		return nil, err
	}
	defer env.Cleanup()

	repo := service.newGitRepo(gctx.Dir, nil, env.Env)
	items, err := repo.Log(context.Background(), limit, offset)
	if err != nil {
		return nil, gitOperationError(err)
	}
	return &vo.GitLogRsp{Items: gitCommitsToVO(items)}, nil
}

// gitDiff 无 commit 时为工作区 vs HEAD，否则展示指定提交的 patch。
func (service *FileManagerService) gitDiff(gctx *GitContext, req *vo.GitDiffReq) (*vo.GitDiffRsp, error) {
	if err := gctx.authorize(false); err != nil {
		return nil, err
	}
	if _, err := service.requireGitSetting(gctx.OwnerType, gctx.OwnerId); err != nil {
		return nil, err
	}
	p := ""
	commit := ""
	if req != nil {
		p = strings.TrimSpace(strings.TrimPrefix(req.Path, "/"))
		commit = strings.TrimSpace(req.Commit)
	}
	if commit != "" && !gitCommitHashRe.MatchString(commit) {
		return nil, errs.NewFormatError("invalid commit: %s", "提交记录无效：%s", commit)
	}

	env, err := git.BuildBaseEnv()
	if err != nil {
		return nil, err
	}
	defer env.Cleanup()

	repo := service.newGitRepo(gctx.Dir, nil, env.Env)
	diff, err := repo.Diff(context.Background(), p, commit, git.DiffMaxChars)
	if err != nil {
		return nil, gitOperationError(err)
	}
	return &vo.GitDiffRsp{Path: p, Commit: commit, Diff: diff}, nil
}

// --- 新建即克隆 ---

// gitValidateCloneRequest 创建项目前的 Git 来源参数校验：失败直接拒绝创建，不落任何数据。
func (service *FileManagerService) gitValidateCloneRequest(req *vo.GitCloneReq) error {
	if req == nil {
		return errs.ErrGitNotConfigured
	}
	if req.IsLocal() {
		if strings.TrimSpace(req.RemoteUrl) != "" {
			return errs.NewFormatError("local mode must not set remoteUrl", "local 模式不允许携带 remoteUrl")
		}
		return nil
	}
	if err := validateGitRemoteUrl(req.RemoteUrl); err != nil {
		return err
	}
	switch req.AuthType {
	case po.GitAuthNone:
	case po.GitAuthSSH:
		if strings.TrimSpace(req.SshPrivateKey) == "" {
			return errs.ErrGitNotConfigured
		}
	case po.GitAuthHTTPS:
		if strings.TrimSpace(req.HttpsSecret) == "" {
			return errs.ErrGitNotConfigured
		}
	default:
		return errs.ErrGitNotConfigured
	}
	return nil
}

// gitInitOrClone 项目创建后的 Git 初始化入口：local 模式同步初始化本地仓库，clone 模式异步克隆。
func (service *FileManagerService) gitInitOrClone(gctx *GitContext, req *vo.GitCloneReq) error {
	if req == nil {
		return nil
	}
	if req.IsLocal() {
		return service.gitStartLocalInit(gctx)
	}
	return service.gitStartClone(gctx, req)
}

// gitStartLocalInit 创建项目时开启本地 Git：同步 init + 默认 .gitignore + 首次提交，无远程仓库。
// 初始化为毫秒级本地操作，走同步路径；失败由调用方（服务层）回滚删除项目。
func (service *FileManagerService) gitStartLocalInit(gctx *GitContext) error {
	if err := service.ensureWritable(); err != nil {
		return err
	}
	if err := gctx.authorize(true); err != nil {
		return err
	}
	if !git.BinaryAvailable() {
		return errs.ErrGitBinaryMissing
	}

	unlock, err := service.lockOwner(gctx.OwnerType, gctx.OwnerId, gitLockTTL, gctx.UserId)
	if err != nil {
		return err
	}
	defer unlock()

	setting, err := service.loadSetting(gctx.OwnerType, gctx.OwnerId)
	if err != nil {
		return err
	}
	if setting == nil {
		setting = defaultProjectGitSetting(gctx.OwnerType, gctx.OwnerId)
	}
	if setting.CloneState == po.GitCloneRunning {
		return errs.ErrGitCloning
	}
	if err := validateGitSetting(setting); err != nil {
		return err
	}
	if err := service.initializeRepository(context.Background(), gctx, setting); err != nil {
		return err
	}
	return service.GitSettingRepo.Upsert(setting)
}

// gitStartClone 创建项目后启动异步克隆：写入配置（克隆中）并返回，goroutine 完成后回写状态。
func (service *FileManagerService) gitStartClone(gctx *GitContext, req *vo.GitCloneReq) error {
	if err := service.ensureWritable(); err != nil {
		return err
	}
	if err := service.gitValidateCloneRequest(req); err != nil {
		return err
	}
	if err := gctx.authorize(true); err != nil {
		return err
	}
	if !git.BinaryAvailable() {
		return errs.ErrGitBinaryMissing
	}

	setting := defaultProjectGitSetting(gctx.OwnerType, gctx.OwnerId)
	if existing, err := service.loadSetting(gctx.OwnerType, gctx.OwnerId); err != nil {
		return err
	} else if existing != nil {
		if existing.CloneState == po.GitCloneRunning {
			return errs.ErrGitCloning
		}
		setting.Id = existing.Id
	}
	setting.RemoteUrl = strings.TrimSpace(req.RemoteUrl)
	setting.AuthType = req.AuthType
	setting.SshPrivateKey = req.SshPrivateKey
	setting.HttpsUsername = req.HttpsUsername
	setting.HttpsSecret = req.HttpsSecret
	setting.CloneState = po.GitCloneRunning
	setting.CloneMessage = ""
	if err := service.GitSettingRepo.Upsert(setting); err != nil {
		return err
	}
	service.launchClone(setting, gctx.Dir, req.ShallowEnabled())
	return nil
}

// gitRetryClone 克隆失败后重试：清空项目目录内容后重新克隆。
func (service *FileManagerService) gitRetryClone(gctx *GitContext) error {
	if err := service.ensureWritable(); err != nil {
		return err
	}
	if err := gctx.authorize(true); err != nil {
		return err
	}
	if !git.BinaryAvailable() {
		return errs.ErrGitBinaryMissing
	}
	setting, err := service.loadSetting(gctx.OwnerType, gctx.OwnerId)
	if err != nil {
		return err
	}
	if setting == nil || strings.TrimSpace(setting.RemoteUrl) == "" {
		return errs.ErrGitNotConfigured
	}
	if setting.CloneState == po.GitCloneRunning {
		return errs.ErrGitCloning
	}

	entries, err := os.ReadDir(gctx.Dir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(gctx.Dir, entry.Name())); err != nil {
			return err
		}
	}
	setting.CloneState = po.GitCloneRunning
	setting.CloneMessage = ""
	if err := service.GitSettingRepo.Upsert(setting); err != nil {
		return err
	}
	service.launchClone(setting, gctx.Dir, true)
	return nil
}

// launchClone 注册取消句柄并异步执行克隆。
func (service *FileManagerService) launchClone(setting *po.ProjectGitSetting, dir string, shallow bool) {
	ctx, cancel := context.WithCancel(context.Background())
	key := gitOwnerKey(setting.OwnerType, setting.OwnerId)
	git.RegisterCloneCancel(key, cancel)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				freedom.Logger().Errorf("git clone panic owner=%d:%d: %v", setting.OwnerType, setting.OwnerId, r)
			}
			cancel()
			git.UnregisterCloneCancel(key)
		}()
		service.runClone(ctx, setting, dir, shallow)
	}()
}

func (service *FileManagerService) runClone(ctx context.Context, setting *po.ProjectGitSetting, dir string, shallow bool) {
	ok, err := service.GitSettingRepo.LockOwner(setting.OwnerType, setting.OwnerId, "clone", gitCloneLockTTL)
	if err != nil || !ok {
		service.finishClone(setting, errs.ErrGitBusy, "")
		return
	}
	defer func() { _ = service.GitSettingRepo.UnlockOwner(setting.OwnerType, setting.OwnerId) }()

	if err := os.MkdirAll(dir, 0755); err != nil {
		service.finishClone(setting, err, "")
		return
	}
	opCtx, cancel := context.WithTimeout(ctx, git.CloneTimeout)
	defer cancel()

	env, authErr := git.BuildAuthEnv(config.Get().GetGitSecretDir(), toGitAuthSetting(setting))
	if authErr != nil {
		service.finishClone(setting, gitOperationError(authErr), "")
		return
	}
	defer env.Cleanup()

	repo := service.newGitRepo(dir, setting, env.Env)
	if cloneErr := repo.Clone(opCtx, setting.RemoteUrl, shallow); cloneErr != nil {
		stderr := ""
		var cmdErr *git.CommandError
		if errors.As(cloneErr, &cmdErr) {
			stderr = cmdErr.Stderr
		}
		service.finishClone(setting, cloneErr, stderr)
		return
	}
	service.finishClone(setting, nil, "")
}

// finishClone 回写克隆结果；项目已删除时更新无副作用。
func (service *FileManagerService) finishClone(setting *po.ProjectGitSetting, opErr error, stderr string) {
	secrets := []string{setting.HttpsSecret, setting.SshPrivateKey}
	if opErr == nil {
		_ = service.GitSettingRepo.UpdateCloneResult(setting.OwnerType, setting.OwnerId, po.GitCloneSuccess, "")
		return
	}
	msg := git.SanitizeMessage(strings.TrimSpace(stderr), secrets...)
	if msg == "" {
		msg = git.SanitizeMessage(opErr.Error(), secrets...)
	}
	if git.IsAuthError(stderr) || errors.Is(opErr, git.ErrAuthFailed) {
		msg = errs.ErrGitAuthFailed.Error()
	}
	_ = service.GitSettingRepo.UpdateCloneResult(setting.OwnerType, setting.OwnerId, po.GitCloneFailed, msg)
}

// --- 定时同步 ---

// gitVerifierTimer 启动后台调度：每天 03:00 触发一次全量自动本地提交；未初始化或 Preview 模式不启动。
func (service *FileManagerService) gitVerifierTimer() {
	if !config.Get().System.Initialized {
		return
	}
	if config.Get().Behavior.PreviewUser != "" {
		return
	}
	go func() {
		for {
			time.Sleep(time.Until(git.NextDailySyncAt(time.Now())))
			service.scanAutoSyncs()
		}
	}()
}

// scanAutoSyncs 扫描所有 Git 项目（存在配置行，含无远程的本地仓库），逐个异步执行定时本地提交。
func (service *FileManagerService) scanAutoSyncs() {
	defer func() {
		if r := recover(); r != nil {
			freedom.Logger().Errorf("scanAutoSyncs panic: %v", r)
		}
	}()
	settings, err := service.GitSettingRepo.ListAutoSync()
	if err != nil {
		freedom.Logger().Errorf("scanAutoSyncs list err: %v", err)
		return
	}
	for i := range settings {
		go service.autoSync(settings[i])
	}
}

// autoSync 定时本地 add + commit（机器人身份），不推送远程。
func (service *FileManagerService) autoSync(setting po.ProjectGitSetting) {
	defer func() {
		if r := recover(); r != nil {
			freedom.Logger().Errorf("autoSync panic owner=%d:%d: %v", setting.OwnerType, setting.OwnerId, r)
		}
	}()
	unlock, err := service.lockOwner(setting.OwnerType, setting.OwnerId, gitLockTTL, "timer")
	if err != nil {
		return
	}
	defer unlock()

	fresh, err := service.loadSetting(setting.OwnerType, setting.OwnerId)
	if err != nil || fresh == nil {
		return
	}
	setting = *fresh
	if setting.CloneState == po.GitCloneRunning {
		return
	}
	dir, _, err := service.resolveOwnerDir(setting.OwnerType, setting.OwnerId)
	if err != nil {
		return
	}
	if !git.IsRepository(dir) {
		return
	}
	env, err := git.BuildBaseEnv()
	if err != nil {
		return
	}
	defer env.Cleanup()

	ctx := context.Background()
	repo := service.newGitRepo(dir, &setting, env.Env)
	message := fmt.Sprintf("chore: auto commit %s", time.Now().Format("2006-01-02 15:04"))
	if _, _, err := repo.CommitWorkingTree(ctx, git.BotIdentity, message); err != nil && !errors.Is(err, git.ErrNoChanges) {
		freedom.Logger().Errorf("autoSync commit owner=%d:%d err: %v", setting.OwnerType, setting.OwnerId, err)
	}
}
