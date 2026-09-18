package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/util/git"

	"github.com/8treenet/freedom"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func requireGitBinary(t *testing.T) {
	t.Helper()
	if !git.BinaryAvailable() {
		t.Skip("git binary not available")
	}
}

// hostGit 在宿主机上直接运行 git（测试夹具搭建用），身份通过环境变量注入。
func hostGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_AUTHOR_NAME=fixture",
		"GIT_AUTHOR_EMAIL=fixture@test.dev",
		"GIT_COMMITTER_NAME=fixture",
		"GIT_COMMITTER_EMAIL=fixture@test.dev",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func hostGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func makeBareRemote(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "remote.git")
	hostGit(t, "", "init", "--bare", "--initial-branch=main", dir)
	return dir
}

// cloneAndPush 从裸仓库克隆后新增一个文件并推送（模拟他人变更）。
func cloneAndPush(t *testing.T, bare, filename, content string) {
	t.Helper()
	other := filepath.Join(t.TempDir(), "other")
	hostGit(t, "", "clone", bare, other)
	if err := os.WriteFile(filepath.Join(other, filename), []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", filename, err)
	}
	hostGit(t, other, "add", "--", filename)
	hostGit(t, other, "commit", "-m", "other: "+filename)
	hostGit(t, other, "push")
}

// testRepo 构造带脱敏密钥的仓库操作句柄。
func testRepo(dir string, setting *po.ProjectGitSetting, env []string) *git.Repo {
	repo := &git.Repo{Dir: dir, Env: env}
	if setting != nil {
		repo.Secrets = []string{setting.HttpsSecret, setting.SshPrivateKey}
	}
	return repo
}

// bootstrapSharedRepo 建立与裸仓库共享历史的本地仓库并完成首次推送。
func bootstrapSharedRepo(t *testing.T, svc *FileManagerService, bare string) (work string, setting *po.ProjectGitSetting, repo *git.Repo) {
	t.Helper()
	secretDir := t.TempDir()
	work = t.TempDir()
	setting = &po.ProjectGitSetting{
		OwnerType: po.GitOwnerUserProject,
		OwnerId:   1,
		RemoteUrl: bare,
		AuthType:  po.GitAuthNone,
	}
	ctx := context.Background()
	repo = testRepo(work, setting, nil)
	if _, err := repo.Initialize(ctx, setting.RemoteUrl, git.BotIdentity); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	env, err := git.BuildAuthEnv(secretDir, toGitAuthSetting(setting))
	if err != nil {
		t.Fatalf("BuildAuthEnv: %v", err)
	}
	t.Cleanup(env.Cleanup)
	repo.Env = env.Env
	if err := os.WriteFile(filepath.Join(work, "readme.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	if _, _, err := repo.CommitWorkingTree(ctx, svc.manualIdentity("u1"), "feat: readme"); err != nil {
		t.Fatalf("first commit: %v", err)
	}
	if err := repo.Push(ctx, setting.RemoteUrl); err != nil {
		t.Fatalf("first push: %v", err)
	}
	return work, setting, repo
}

func TestGitIntegrationInitCommitPushPull(t *testing.T) {
	requireGitBinary(t)
	svc := newGitServiceTestEnv(t)
	bare := makeBareRemote(t)
	work, setting, repo := bootstrapSharedRepo(t, svc, bare)
	ctx := context.Background()

	if !git.IsRepository(work) {
		t.Fatal("repository not initialized")
	}
	if _, err := os.Stat(filepath.Join(work, ".gitignore")); err != nil {
		t.Fatalf(".gitignore not created: %v", err)
	}
	if got := hostGitOutput(t, bare, "rev-parse", "--verify", "refs/heads/main"); got == "" {
		t.Fatal("remote main branch missing after push")
	}

	cloneAndPush(t, bare, "remote.txt", "from other\n")
	if err := repo.Pull(ctx, setting.RemoteUrl); err != nil {
		t.Fatalf("Pull: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(work, "remote.txt"))
	if err != nil || string(data) != "from other\n" {
		t.Fatalf("pulled file = %q, err = %v", string(data), err)
	}
}

func TestGitIntegrationPushRejectedRebaseRetry(t *testing.T) {
	requireGitBinary(t)
	svc := newGitServiceTestEnv(t)
	bare := makeBareRemote(t)
	work, setting, repo := bootstrapSharedRepo(t, svc, bare)
	ctx := context.Background()

	cloneAndPush(t, bare, "remote.txt", "remote\n")
	if err := os.WriteFile(filepath.Join(work, "local.txt"), []byte("local\n"), 0644); err != nil {
		t.Fatalf("write local file: %v", err)
	}
	if _, _, err := repo.CommitWorkingTree(ctx, svc.manualIdentity("u1"), "feat: local"); err != nil {
		t.Fatalf("local commit: %v", err)
	}
	if err := repo.Push(ctx, setting.RemoteUrl); err != nil {
		t.Fatalf("Push should recover via rebase: %v", err)
	}
	if _, err := os.Stat(filepath.Join(work, "remote.txt")); err != nil {
		t.Fatalf("remote change missing after rebase: %v", err)
	}
	if got := hostGitOutput(t, work, "status", "--porcelain"); got != "" {
		t.Fatalf("worktree not clean after retry: %q", got)
	}
	log := hostGitOutput(t, bare, "log", "--oneline", "main")
	if !strings.Contains(log, "feat: local") || !strings.Contains(log, "remote") {
		t.Fatalf("remote log missing commits:\n%s", log)
	}
}

func TestGitIntegrationRebaseConflictAborts(t *testing.T) {
	requireGitBinary(t)
	svc := newGitServiceTestEnv(t)
	bare := makeBareRemote(t)
	work, setting, repo := bootstrapSharedRepo(t, svc, bare)
	ctx := context.Background()

	// 远端与本地修改同一文件制造冲突
	cloneAndPush(t, bare, "readme.md", "theirs\n")
	if err := os.WriteFile(filepath.Join(work, "readme.md"), []byte("ours\n"), 0644); err != nil {
		t.Fatalf("write conflict file: %v", err)
	}
	if _, _, err := repo.CommitWorkingTree(ctx, svc.manualIdentity("u1"), "local change"); err != nil {
		t.Fatalf("local commit: %v", err)
	}
	err := repo.Push(ctx, setting.RemoteUrl)
	if !errors.Is(err, git.ErrConflict) {
		t.Fatalf("Push error = %v, want ErrConflict", err)
	}
	if _, statErr := os.Stat(filepath.Join(work, ".git", "rebase-merge")); !os.IsNotExist(statErr) {
		t.Fatal("rebase state left behind after abort")
	}
	if _, statErr := os.Stat(filepath.Join(work, ".git", "rebase-apply")); !os.IsNotExist(statErr) {
		t.Fatal("rebase state left behind after abort")
	}
	if got := hostGitOutput(t, work, "status", "--porcelain"); got != "" {
		t.Fatalf("worktree not restored after abort: %q", got)
	}
}

func TestGitIntegrationUnrelatedHistory(t *testing.T) {
	requireGitBinary(t)
	bare := makeBareRemote(t)
	cloneAndPush(t, bare, "remote.txt", "remote\n")

	work := t.TempDir()
	setting := &po.ProjectGitSetting{
		OwnerType: po.GitOwnerUserProject,
		OwnerId:   2,
		RemoteUrl: bare,
		AuthType:  po.GitAuthNone,
	}
	ctx := context.Background()
	repo := testRepo(work, setting, nil)
	if _, err := repo.Initialize(ctx, setting.RemoteUrl, git.BotIdentity); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	env, err := git.BuildAuthEnv(t.TempDir(), toGitAuthSetting(setting))
	if err != nil {
		t.Fatalf("BuildAuthEnv: %v", err)
	}
	defer env.Cleanup()
	repo.Env = env.Env

	if err := repo.Push(ctx, setting.RemoteUrl); !errors.Is(err, git.ErrUnrelatedHistory) {
		t.Fatalf("Push error = %v, want ErrUnrelatedHistory", err)
	}

	// 合并路径：pull --allow-unrelated-histories 后推送
	if err := repo.MergeUnrelated(ctx, "main"); err != nil {
		t.Fatalf("MergeUnrelated: %v", err)
	}
	if err := repo.Push(ctx, setting.RemoteUrl); err != nil {
		t.Fatalf("push after merge: %v", err)
	}
	if _, err := os.Stat(filepath.Join(work, "remote.txt")); err != nil {
		t.Fatalf("remote content missing after merge: %v", err)
	}
}

func TestGitIntegrationDirtyPullRejected(t *testing.T) {
	requireGitBinary(t)
	svc := newGitServiceTestEnv(t)
	bare := makeBareRemote(t)
	work, setting, repo := bootstrapSharedRepo(t, svc, bare)
	ctx := context.Background()

	cloneAndPush(t, bare, "remote.txt", "remote\n")
	if err := os.WriteFile(filepath.Join(work, "dirty.txt"), []byte("uncommitted\n"), 0644); err != nil {
		t.Fatalf("write dirty file: %v", err)
	}
	if err := repo.Pull(ctx, setting.RemoteUrl); !errors.Is(err, git.ErrDirtyTree) {
		t.Fatalf("Pull error = %v, want ErrDirtyTree", err)
	}
}

func TestGitIntegrationAdoptExistingRepository(t *testing.T) {
	requireGitBinary(t)
	bare := makeBareRemote(t)
	work := t.TempDir()
	hostGit(t, work, "init", "-b", "main")
	if err := os.WriteFile(filepath.Join(work, "existing.txt"), []byte("existing\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	hostGit(t, work, "add", "--", "existing.txt")
	hostGit(t, work, "commit", "-m", "preexisting")
	hostGit(t, work, "remote", "add", "origin", bare)

	setting := &po.ProjectGitSetting{OwnerType: po.GitOwnerUserProject, OwnerId: 3}
	ctx := context.Background()
	repo := testRepo(work, setting, nil)
	backfill, err := repo.Initialize(ctx, "", git.BotIdentity)
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if backfill != bare {
		t.Fatalf("backfilled RemoteUrl = %q, want %q", backfill, bare)
	}
	if count := hostGitOutput(t, work, "rev-list", "--count", "HEAD"); count != "1" {
		t.Fatalf("commit count = %s, want existing history untouched", count)
	}
	if _, err := os.Stat(filepath.Join(work, ".gitignore")); !os.IsNotExist(err) {
		t.Fatal("adopted repository should not be re-initialized with default .gitignore")
	}
}

// newGitServiceTestEnv 构造带 DB/Redis 的完整 FileManagerService，用于接口级集成测试。
func newGitServiceTestEnv(t *testing.T) *FileManagerService {
	t.Helper()
	origSecretDir := config.Get().Paths.GitSecretDir
	config.Get().Paths.GitSecretDir = t.TempDir()
	t.Cleanup(func() { config.Get().Paths.GitSecretDir = origSecretDir })

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(
		&po.ProjectGitSetting{},
		&po.UserProject{},
		&po.TeamProject{},
		&po.TeamProjectMember{},
		&po.User{},
		&po.Session{},
		&po.AutomationTask{},
	); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	unitTest := freedom.NewUnitTest()
	unitTest.InstallDB(func() interface{} { return db })
	unitTest.InstallRedis(func() redis.Cmdable { return infra.NewCacheWrapper(5*time.Minute, 10*time.Minute) })
	unitTest.Run()

	var settingRepo *repository.ProjectGitSettingRepository
	unitTest.FetchRepository(&settingRepo)
	var userProjectRepo *repository.UserProjectRepository
	unitTest.FetchRepository(&userProjectRepo)
	var teamProjectRepo *repository.TeamProjectRepository
	unitTest.FetchRepository(&teamProjectRepo)
	var userRepo *repository.UserRepository
	unitTest.FetchRepository(&userRepo)

	return &FileManagerService{
		GitSettingRepo:  settingRepo,
		UserProjectRepo: userProjectRepo,
		TeamProjectRepo: teamProjectRepo,
		UserRepo:        userRepo,
	}
}

// newTeamProjectServiceForTest 基于测试环境构造团队项目服务（覆盖基类依赖）。
func newTeamProjectServiceForTest(svc *FileManagerService) *TeamProjectService {
	return &TeamProjectService{
		FileManagerService: FileManagerService{
			GitSettingRepo: svc.GitSettingRepo,
			UserRepo:       svc.UserRepo,
			TeamProjectRepo: svc.TeamProjectRepo,
		},
		TPRepo: svc.TeamProjectRepo,
	}
}

// newMyProjectServiceForTest 基于测试环境构造个人项目服务（覆盖基类依赖）。
func newMyProjectServiceForTest(svc *FileManagerService) *MyProjectService {
	return &MyProjectService{
		FileManagerService: FileManagerService{
			GitSettingRepo:  svc.GitSettingRepo,
			UserRepo:        svc.UserRepo,
			UserProjectRepo: svc.UserProjectRepo,
		},
		Repo: svc.UserProjectRepo,
	}
}

func TestGitServiceCloneSuccessAndFailureRetry(t *testing.T) {
	requireGitBinary(t)
	svc := newGitServiceTestEnv(t)
	ctx := context.Background()

	bare := makeBareRemote(t)
	cloneAndPush(t, bare, "a.txt", "a\n")
	cloneAndPush(t, bare, "b.txt", "b\n")
	fileURL := "file://" + bare

	t.Run("clone success is shallow", func(t *testing.T) {
		work := t.TempDir()
		setting := &po.ProjectGitSetting{
			OwnerType:  po.GitOwnerUserProject,
			OwnerId:    100,
			RemoteUrl:  fileURL,
			AuthType:   po.GitAuthNone,
			CloneState: po.GitCloneRunning,
		}
		if err := svc.GitSettingRepo.Upsert(setting); err != nil {
			t.Fatalf("upsert setting: %v", err)
		}
		svc.runClone(ctx, setting, work, true)

		got, err := svc.GitSettingRepo.GetByOwner(po.GitOwnerUserProject, 100)
		if err != nil {
			t.Fatalf("get setting: %v", err)
		}
		if got.CloneState != po.GitCloneSuccess || got.CloneMessage != "" {
			t.Fatalf("clone state = %d message = %q, want success", got.CloneState, got.CloneMessage)
		}
		if _, err := os.Stat(filepath.Join(work, "b.txt")); err != nil {
			t.Fatalf("cloned file missing: %v", err)
		}
		if count := hostGitOutput(t, work, "rev-list", "--count", "HEAD"); count != "1" {
			t.Fatalf("shallow clone commit count = %s, want 1", count)
		}
	})

	t.Run("clone failure then retry succeeds", func(t *testing.T) {
		work := t.TempDir()
		setting := &po.ProjectGitSetting{
			OwnerType:  po.GitOwnerUserProject,
			OwnerId:    101,
			RemoteUrl:  "http://127.0.0.1:1/nope.git",
			AuthType:   po.GitAuthNone,
			CloneState: po.GitCloneRunning,
		}
		if err := svc.GitSettingRepo.Upsert(setting); err != nil {
			t.Fatalf("upsert setting: %v", err)
		}
		svc.runClone(ctx, setting, work, true)

		got, _ := svc.GitSettingRepo.GetByOwner(po.GitOwnerUserProject, 101)
		if got.CloneState != po.GitCloneFailed || got.CloneMessage == "" {
			t.Fatalf("clone state = %d message = %q, want failure with message", got.CloneState, got.CloneMessage)
		}

		// 清空后以正确地址重试
		got.RemoteUrl = fileURL
		got.CloneState = po.GitCloneRunning
		got.CloneMessage = ""
		if err := svc.GitSettingRepo.Upsert(got); err != nil {
			t.Fatalf("upsert retry setting: %v", err)
		}
		entries, _ := os.ReadDir(work)
		for _, entry := range entries {
			_ = os.RemoveAll(filepath.Join(work, entry.Name()))
		}
		svc.runClone(ctx, got, work, true)

		got, _ = svc.GitSettingRepo.GetByOwner(po.GitOwnerUserProject, 101)
		if got.CloneState != po.GitCloneSuccess {
			t.Fatalf("retry clone state = %d message = %q, want success", got.CloneState, got.CloneMessage)
		}
	})

	t.Run("auth failure is classified", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("WWW-Authenticate", `Basic realm="git"`)
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		work := t.TempDir()
		setting := &po.ProjectGitSetting{
			OwnerType:     po.GitOwnerUserProject,
			OwnerId:       102,
			RemoteUrl:     server.URL + "/repo.git",
			AuthType:      po.GitAuthHTTPS,
			HttpsUsername: "bot",
			HttpsSecret:   "wrong-token",
			CloneState:    po.GitCloneRunning,
		}
		if err := svc.GitSettingRepo.Upsert(setting); err != nil {
			t.Fatalf("upsert setting: %v", err)
		}
		svc.runClone(ctx, setting, work, true)

		got, _ := svc.GitSettingRepo.GetByOwner(po.GitOwnerUserProject, 102)
		if got.CloneState != po.GitCloneFailed {
			t.Fatalf("clone state = %d, want failed", got.CloneState)
		}
		if strings.Contains(got.CloneMessage, "wrong-token") {
			t.Fatalf("clone message leaked credential: %q", got.CloneMessage)
		}
	})
}

func TestGitCloneCancelRegistry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	key := gitOwnerKey(po.GitOwnerTeamProject, 88)
	git.RegisterCloneCancel(key, cancel)
	cancelGitClone(po.GitOwnerTeamProject, 88)
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("registered clone context was not canceled")
	}
	if git.CloneCancelRegistered(key) {
		t.Fatal("cancel registry entry not removed")
	}
}

// setupTestGitProject 模拟“创建项目时选择本地 Git”：init + 默认 .gitignore + 首次提交，
// 创建后仓库不可开启或关闭（是否为 Git 项目由配置行是否存在决定）。
func setupTestGitProject(t *testing.T, svc *FileManagerService, ownerType uint8, ownerId int, dir, userId string) *po.ProjectGitSetting {
	t.Helper()
	gctx := &GitContext{UserId: userId, OwnerType: ownerType, OwnerId: ownerId, Dir: dir}
	if ownerType == po.GitOwnerUserProject {
		gctx.IsOwner = true
	} else {
		gctx.IsCreator = true
		gctx.IsMember = true
	}
	if err := svc.gitStartLocalInit(gctx); err != nil {
		t.Fatalf("gitStartLocalInit: %v", err)
	}
	setting, err := svc.GitSettingRepo.GetByOwner(ownerType, ownerId)
	if err != nil {
		t.Fatalf("get git setting: %v", err)
	}
	return setting
}

func TestGitServicePermission(t *testing.T) {
	requireGitBinary(t)
	svc := newGitServiceTestEnv(t)
	tpSev := newTeamProjectServiceForTest(svc)

	// 团队项目目录指向临时目录，避免污染真实工作区
	origTeamDir := config.Get().Paths.TeamProjectDir
	config.Get().Paths.TeamProjectDir = t.TempDir()

	teamRepo := svc.TeamProjectRepo
	project := &po.TeamProject{CreatorId: "creator", ProjectName: "team-proj"}
	if err := teamRepo.Create(project); err != nil {
		t.Fatalf("create team project: %v", err)
	}
	if err := teamRepo.AddMember(project.Id, "member"); err != nil {
		t.Fatalf("add member: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(config.Get().GetTeamProjectDir(), project.ProjectName), 0755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	t.Cleanup(func() { config.Get().Paths.TeamProjectDir = origTeamDir })

	// 非成员的 git 只读也被拒绝
	if _, err := tpSev.GitStatus("outsider", project.Id); !errors.Is(err, errs.ErrGitPermission) {
		t.Fatalf("outsider status error = %v, want ErrGitPermission", err)
	}

	// 成员可读
	if _, err := tpSev.GitStatus("member", project.Id); err != nil {
		t.Fatalf("member status error = %v", err)
	}

	// 非 Git 项目（无配置行）：Git 只读 API 也不可用
	if _, err := tpSev.GitLog("creator", project.Id, nil); !errors.Is(err, errs.ErrGitDisabled) {
		t.Fatalf("log on non-git project error = %v, want ErrGitDisabled", err)
	}

	// 创建项目时选择本地 Git：init + .gitignore + 首次提交
	projectDir := filepath.Join(config.Get().GetTeamProjectDir(), project.ProjectName)
	setupTestGitProject(t, svc, po.GitOwnerTeamProject, project.Id, projectDir, "creator")
	rsp, err := tpSev.GitStatus("creator", project.Id)
	if err != nil {
		t.Fatalf("creator status: %v", err)
	}
	if !rsp.Enabled || !rsp.Initialized {
		t.Fatalf("status after local init = %#v", rsp)
	}
	if !git.IsRepository(projectDir) {
		t.Fatal("repository not initialized on create")
	}

	// 成员可提交
	writeErr := os.WriteFile(filepath.Join(projectDir, "member.txt"), []byte("member\n"), 0644)
	if writeErr != nil {
		t.Fatalf("write member file: %v", writeErr)
	}
	if _, err := tpSev.GitCommit("member", project.Id, &vo.GitCommitReq{Message: "member commit"}); err != nil {
		t.Fatalf("member commit: %v", err)
	}

	// 成员推送（未配置远程）返回未配置
	if err := tpSev.GitPush("member", project.Id); !errors.Is(err, errs.ErrGitNotConfigured) {
		t.Fatalf("member push error = %v, want ErrGitNotConfigured", err)
	}

	// preview 模式写操作拒绝
	origPreview := config.Get().Behavior.PreviewUser
	config.Get().Behavior.PreviewUser = "creator"
	t.Cleanup(func() { config.Get().Behavior.PreviewUser = origPreview })
	if _, err := tpSev.GitCommit("creator", project.Id, &vo.GitCommitReq{Message: "x"}); !errors.Is(err, errs.ErrGitPreviewReadOnly) {
		t.Fatalf("preview commit error = %v, want ErrGitPreviewReadOnly", err)
	}
	config.Get().Behavior.PreviewUser = origPreview
}

func TestGitServiceAutoSyncScheduling(t *testing.T) {
	requireGitBinary(t)
	svc := newGitServiceTestEnv(t)
	tpSev := newTeamProjectServiceForTest(svc)

	origTeamDir := config.Get().Paths.TeamProjectDir
	config.Get().Paths.TeamProjectDir = t.TempDir()
	t.Cleanup(func() { config.Get().Paths.TeamProjectDir = origTeamDir })

	project := &po.TeamProject{CreatorId: "creator", ProjectName: "sched-proj"}
	if err := svc.TeamProjectRepo.Create(project); err != nil {
		t.Fatalf("create team project: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(config.Get().GetTeamProjectDir(), project.ProjectName), 0755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	projectDir := filepath.Join(config.Get().GetTeamProjectDir(), project.ProjectName)
	setupTestGitProject(t, svc, po.GitOwnerTeamProject, project.Id, projectDir, "creator")

	// 定时扫描：所有 Git 项目（存在配置行）
	list, err := svc.GitSettingRepo.ListAutoSync()
	if err != nil {
		t.Fatalf("list auto sync: %v", err)
	}
	if len(list) != 1 || list[0].OwnerId != project.Id {
		t.Fatalf("list auto sync = %#v, want only project %d", list, project.Id)
	}

	// 定时提交：本地 add + commit（机器人身份），不推送远程
	if err := os.WriteFile(filepath.Join(projectDir, "auto.txt"), []byte("auto\n"), 0644); err != nil {
		t.Fatalf("write auto file: %v", err)
	}
	tpSev.autoSync(*mustGetSetting(t, svc, po.GitOwnerTeamProject, project.Id))
	if log := hostGitOutput(t, projectDir, "log", "--format=%an", "main"); !strings.Contains(log, git.DefaultBotName) {
		t.Fatalf("local commit author = %q, want bot identity", log)
	}
	if got := hostGitOutput(t, projectDir, "status", "--porcelain"); got != "" {
		t.Fatalf("working tree not clean after auto sync: %q", got)
	}
}

func mustGetSetting(t *testing.T, svc *FileManagerService, ownerType uint8, ownerId int) *po.ProjectGitSetting {
	t.Helper()
	setting, err := svc.GitSettingRepo.GetByOwner(ownerType, ownerId)
	if err != nil {
		t.Fatalf("get setting: %v", err)
	}
	return setting
}

func TestGitServiceLogAndDiff(t *testing.T) {
	requireGitBinary(t)
	svc := newGitServiceTestEnv(t)
	tpSev := newTeamProjectServiceForTest(svc)

	origTeamDir := config.Get().Paths.TeamProjectDir
	config.Get().Paths.TeamProjectDir = t.TempDir()
	t.Cleanup(func() { config.Get().Paths.TeamProjectDir = origTeamDir })

	project := &po.TeamProject{CreatorId: "creator", ProjectName: "log-proj"}
	if err := svc.TeamProjectRepo.Create(project); err != nil {
		t.Fatalf("create team project: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(config.Get().GetTeamProjectDir(), project.ProjectName), 0755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	setupTestGitProject(t, svc, po.GitOwnerTeamProject, project.Id, filepath.Join(config.Get().GetTeamProjectDir(), project.ProjectName), "creator")

	logRsp, err := tpSev.GitLog("creator", project.Id, &vo.GitLogReq{Limit: 10})
	if err != nil {
		t.Fatalf("log: %v", err)
	}
	if len(logRsp.Items) != 1 || logRsp.Items[0].Message != git.InitCommitMessage {
		t.Fatalf("log items = %#v, want initialize commit", logRsp.Items)
	}
	if logRsp.Items[0].Author != git.DefaultBotName {
		t.Fatalf("init commit author = %q, want bot", logRsp.Items[0].Author)
	}

	projectDir := filepath.Join(config.Get().GetTeamProjectDir(), project.ProjectName)
	gitignorePath := filepath.Join(projectDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(git.DefaultGitignoreContent+"tmp/\n"), 0644); err != nil {
		t.Fatalf("modify tracked file: %v", err)
	}
	diffRsp, err := tpSev.GitDiff("creator", project.Id, &vo.GitDiffReq{Path: ".gitignore"})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if !strings.Contains(diffRsp.Diff, "tmp/") {
		t.Fatalf("diff missing tracked change:\n%s", diffRsp.Diff)
	}

	if err := os.WriteFile(filepath.Join(projectDir, "committed.txt"), []byte("content\n"), 0644); err != nil {
		t.Fatalf("write committed file: %v", err)
	}
	commitRsp, err := tpSev.GitCommit("creator", project.Id, &vo.GitCommitReq{Message: "feat: committed"})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	showRsp, err := tpSev.GitDiff("creator", project.Id, &vo.GitDiffReq{Commit: commitRsp.Hash})
	if err != nil {
		t.Fatalf("diff commit: %v", err)
	}
	if !strings.Contains(showRsp.Diff, "committed.txt") {
		t.Fatalf("commit diff missing file:\n%s", showRsp.Diff)
	}
}

func TestMyProjectGitBadgesPermissionAndCascade(t *testing.T) {
	svc := newGitServiceTestEnv(t)
	mpSev := newMyProjectServiceForTest(svc)

	origUserSpace := config.Get().Paths.UserSpace
	config.Get().Paths.UserSpace = t.TempDir()
	t.Cleanup(func() { config.Get().Paths.UserSpace = origUserSpace })

	if err := svc.UserRepo.CreateUser(&po.User{UserId: "p-owner", Username: "powner", Password: "x"}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	project := &po.UserProject{UserId: "p-owner", ProjectName: "my-proj"}
	if err := svc.UserProjectRepo.Create(project); err != nil {
		t.Fatalf("create user project: %v", err)
	}
	projectDir := filepath.Join(config.Get().GetUserSpace("powner"), "projects", project.ProjectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	setting := &po.ProjectGitSetting{
		OwnerType:  po.GitOwnerUserProject,
		OwnerId:    project.Id,
		CloneState: po.GitCloneSuccess,
	}
	if err := svc.GitSettingRepo.Upsert(setting); err != nil {
		t.Fatalf("upsert setting: %v", err)
	}

	// 个人项目权限：仅 owner
	if _, err := mpSev.GitStatus("p-owner", project.Id); err != nil {
		t.Fatalf("owner status: %v", err)
	}
	if _, err := mpSev.GitStatus("other", project.Id); !errors.Is(err, errs.ErrGitPermission) {
		t.Fatalf("other status error = %v, want ErrGitPermission", err)
	}

	// 列表徽标
	list, err := mpSev.List("p-owner")
	if err != nil {
		t.Fatalf("my project list: %v", err)
	}
	if len(list.Items) != 1 || !list.Items[0].GitEnabled || list.Items[0].CloneState != po.GitCloneSuccess {
		t.Fatalf("list items = %#v", list.Items)
	}

	// 删除级联：Git 配置与 known_hosts 清理
	knownHosts := git.ProjectKnownHostsPath(config.Get().GetGitSecretDir(), po.GitOwnerUserProject, project.Id)
	if err := os.MkdirAll(filepath.Dir(knownHosts), 0700); err != nil {
		t.Fatalf("mkdir known_hosts dir: %v", err)
	}
	if err := os.WriteFile(knownHosts, []byte("host key"), 0600); err != nil {
		t.Fatalf("write known_hosts: %v", err)
	}
	if err := mpSev.DeleteProject("p-owner", project.Id); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if _, err := svc.GitSettingRepo.GetByOwner(po.GitOwnerUserProject, project.Id); err == nil {
		t.Fatal("git setting row still exists after project delete")
	}
	if _, err := os.Stat(knownHosts); !os.IsNotExist(err) {
		t.Fatalf("known_hosts not cleaned: %v", err)
	}
}

func TestTeamProjectGitBadgesAndCascade(t *testing.T) {
	svc := newGitServiceTestEnv(t)
	tpSev := newTeamProjectServiceForTest(svc)

	origTeamDir := config.Get().Paths.TeamProjectDir
	config.Get().Paths.TeamProjectDir = t.TempDir()
	t.Cleanup(func() { config.Get().Paths.TeamProjectDir = origTeamDir })

	project := &po.TeamProject{CreatorId: "creator", ProjectName: "cascade-proj"}
	if err := svc.TeamProjectRepo.Create(project); err != nil {
		t.Fatalf("create team project: %v", err)
	}
	projectDir := filepath.Join(config.Get().GetTeamProjectDir(), project.ProjectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	setting := &po.ProjectGitSetting{
		OwnerType:  po.GitOwnerTeamProject,
		OwnerId:    project.Id,
		CloneState: po.GitCloneSuccess,
	}
	if err := svc.GitSettingRepo.Upsert(setting); err != nil {
		t.Fatalf("upsert setting: %v", err)
	}

	list, err := tpSev.List("creator")
	if err != nil {
		t.Fatalf("team project list: %v", err)
	}
	if len(list.Items) != 1 || !list.Items[0].GitEnabled {
		t.Fatalf("list items = %#v", list.Items)
	}

	if err := tpSev.DeleteProject("outsider", project.Id); !errors.Is(err, errs.ErrTeamProjectPermission) {
		t.Fatalf("delete by outsider error = %v, want ErrTeamProjectPermission", err)
	}
	if err := tpSev.DeleteProject("creator", project.Id); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if _, err := svc.GitSettingRepo.GetByOwner(po.GitOwnerTeamProject, project.Id); err == nil {
		t.Fatal("git setting row still exists after project delete")
	}
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		t.Fatalf("project dir not removed: %v", err)
	}
}

func TestGitServiceTestRemoteDirect(t *testing.T) {
	requireGitBinary(t)
	svc := newGitServiceTestEnv(t)
	bare := makeBareRemote(t)
	cloneAndPush(t, bare, "a.txt", "a\n")
	wantHead := hostGitOutput(t, bare, "rev-parse", "HEAD")

	authNone := po.GitAuthNone
	rsp, err := svc.gitTestRemoteDirect(&vo.GitTestReq{RemoteUrl: bare, AuthType: &authNone})
	if err != nil {
		t.Fatalf("gitTestRemoteDirect: %v", err)
	}
	if rsp.Head != wantHead {
		t.Fatalf("head = %q, want %q", rsp.Head, wantHead)
	}

	if _, err := svc.gitTestRemoteDirect(&vo.GitTestReq{}); !errors.Is(err, errs.ErrGitNotConfigured) {
		t.Fatalf("empty request error = %v, want ErrGitNotConfigured", err)
	}
	if _, err := svc.gitTestRemoteDirect(&vo.GitTestReq{RemoteUrl: "https://user:pass@example.com/x.git"}); !errors.Is(err, errs.ErrGitInvalidRemoteUrl) {
		t.Fatalf("embedded credential error = %v, want ErrGitInvalidRemoteUrl", err)
	}
}

func TestProjectCreateWithGitFlow(t *testing.T) {
	requireGitBinary(t)
	svc := newGitServiceTestEnv(t)
	mpSev := newMyProjectServiceForTest(svc)
	tpSev := newTeamProjectServiceForTest(svc)

	origUserSpace := config.Get().Paths.UserSpace
	config.Get().Paths.UserSpace = t.TempDir()
	origTeamDir := config.Get().Paths.TeamProjectDir
	config.Get().Paths.TeamProjectDir = t.TempDir()
	t.Cleanup(func() {
		config.Get().Paths.UserSpace = origUserSpace
		config.Get().Paths.TeamProjectDir = origTeamDir
	})

	if err := svc.UserRepo.CreateUser(&po.User{UserId: "git-owner", Username: "gitowner", Password: "x"}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	t.Run("local init rolls back on user project", func(t *testing.T) {
		// local 模式：创建即初始化仓库
		rsp, err := mpSev.Create("git-owner", "local-proj", "", &vo.GitCloneReq{Mode: vo.GitModeLocal})
		if err != nil {
			t.Fatalf("create with local git: %v", err)
		}
		projectDir := filepath.Join(config.Get().GetUserSpace("gitowner"), "projects", "local-proj")
		if !git.IsRepository(projectDir) {
			t.Fatal("repository not initialized on create")
		}
		if _, err := svc.GitSettingRepo.GetByOwner(po.GitOwnerUserProject, rsp.Id); err != nil {
			t.Fatalf("git setting row missing: %v", err)
		}
	})

	t.Run("invalid clone request rejected without side effects", func(t *testing.T) {
		before, _ := mpSev.List("git-owner")
		countBefore := len(before.Items)
		if _, err := mpSev.Create("git-owner", "bad-proj", "", &vo.GitCloneReq{Mode: vo.GitModeClone, AuthType: po.GitAuthSSH}); !errors.Is(err, errs.ErrGitInvalidRemoteUrl) {
			t.Fatalf("invalid clone request error = %v, want ErrGitInvalidRemoteUrl", err)
		}
		after, _ := mpSev.List("git-owner")
		if len(after.Items) != countBefore {
			t.Fatalf("invalid git request should not create project")
		}
	})

	t.Run("team project local init", func(t *testing.T) {
		rsp, err := tpSev.Create("git-owner", "team-local-proj", "", &vo.GitCloneReq{Mode: vo.GitModeLocal})
		if err != nil {
			t.Fatalf("create team project with local git: %v", err)
		}
		projectDir := filepath.Join(config.Get().GetTeamProjectDir(), "team-local-proj")
		if !git.IsRepository(projectDir) {
			t.Fatal("team repository not initialized on create")
		}
		if _, err := svc.GitSettingRepo.GetByOwner(po.GitOwnerTeamProject, rsp.Id); err != nil {
			t.Fatalf("team git setting row missing: %v", err)
		}
	})
}
