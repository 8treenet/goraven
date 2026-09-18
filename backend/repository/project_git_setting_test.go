package repository_test

import (
	"testing"
	"time"

	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"

	"github.com/8treenet/freedom"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newProjectGitSettingRepositoryTest(t *testing.T) (*gorm.DB, *repository.ProjectGitSettingRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&po.ProjectGitSetting{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	unitTest := freedom.NewUnitTest()
	unitTest.InstallDB(func() interface{} { return db })
	unitTest.InstallRedis(func() redis.Cmdable { return infra.NewCacheWrapper(5*time.Minute, 10*time.Minute) })
	unitTest.Run()

	var repo *repository.ProjectGitSettingRepository
	unitTest.FetchRepository(&repo)
	return db, repo
}

func TestProjectGitSettingRepositoryCRUD(t *testing.T) {
	_, repo := newProjectGitSettingRepositoryTest(t)

	setting := &po.ProjectGitSetting{
		OwnerType:     po.GitOwnerUserProject,
		OwnerId:       11,
		RemoteUrl:     "https://example.com/a.git",
		AuthType:      po.GitAuthHTTPS,
		HttpsUsername: "bot",
		HttpsSecret:   "token",
	}
	if err := repo.Upsert(setting); err != nil {
		t.Fatalf("upsert insert: %v", err)
	}
	if setting.Id == 0 {
		t.Fatal("upsert did not fill id")
	}

	got, err := repo.GetByOwner(po.GitOwnerUserProject, 11)
	if err != nil {
		t.Fatalf("get by owner: %v", err)
	}
	if got.RemoteUrl != setting.RemoteUrl || got.AuthType != po.GitAuthHTTPS {
		t.Fatalf("got = %#v", got)
	}

	setting.RemoteUrl = "https://example.com/b.git"
	if err := repo.Upsert(setting); err != nil {
		t.Fatalf("upsert update: %v", err)
	}
	got, err = repo.GetByOwner(po.GitOwnerUserProject, 11)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if got.RemoteUrl != "https://example.com/b.git" {
		t.Fatalf("RemoteUrl = %q, want %q", got.RemoteUrl, "https://example.com/b.git")
	}

	byOwners, err := repo.ListByOwners(po.GitOwnerUserProject, []int{11, 12})
	if err != nil {
		t.Fatalf("list by owners: %v", err)
	}
	if len(byOwners) != 1 || byOwners[11] == nil {
		t.Fatalf("list by owners = %#v", byOwners)
	}

	if err := repo.DeleteByOwner(po.GitOwnerUserProject, 11); err != nil {
		t.Fatalf("delete by owner: %v", err)
	}
	if _, err := repo.GetByOwner(po.GitOwnerUserProject, 11); err == nil {
		t.Fatal("get after delete returned no error")
	}
}

func TestProjectGitSettingRepositoryListAutoSync(t *testing.T) {
	_, repo := newProjectGitSettingRepositoryTest(t)

	withRemote := &po.ProjectGitSetting{
		OwnerType: po.GitOwnerUserProject,
		OwnerId:   1,
		RemoteUrl: "https://example.com/a.git",
	}
	noRemote := &po.ProjectGitSetting{
		OwnerType: po.GitOwnerUserProject,
		OwnerId:   3,
	}
	for _, setting := range []*po.ProjectGitSetting{withRemote, noRemote} {
		if err := repo.Upsert(setting); err != nil {
			t.Fatalf("upsert: %v", err)
		}
	}

	// 存在配置行即 Git 项目，含无远程的本地仓库
	list, err := repo.ListAutoSync()
	if err != nil {
		t.Fatalf("list auto sync: %v", err)
	}
	if len(list) != 2 || list[0].OwnerId != 1 || list[1].OwnerId != 3 {
		t.Fatalf("list auto sync = %#v, want owner 1 and 3", list)
	}

	if err := repo.UpdateCloneResult(po.GitOwnerUserProject, 1, po.GitCloneFailed, "bad url"); err != nil {
		t.Fatalf("update clone result: %v", err)
	}
	got, _ := repo.GetByOwner(po.GitOwnerUserProject, 1)
	if got.CloneState != po.GitCloneFailed || got.CloneMessage != "bad url" {
		t.Fatalf("clone result = %#v", got)
	}
}

func TestProjectGitSettingRepositoryLock(t *testing.T) {
	_, repo := newProjectGitSettingRepositoryTest(t)

	ok, err := repo.LockOwner(po.GitOwnerTeamProject, 5, "user-1", 10*time.Minute)
	if err != nil {
		t.Fatalf("lock owner: %v", err)
	}
	if !ok {
		t.Fatal("first lock = false, want true")
	}
	ok, err = repo.LockOwner(po.GitOwnerTeamProject, 5, "user-2", 10*time.Minute)
	if err != nil {
		t.Fatalf("second lock: %v", err)
	}
	if ok {
		t.Fatal("second lock = true, want false")
	}
	if err := repo.UnlockOwner(po.GitOwnerTeamProject, 5); err != nil {
		t.Fatalf("unlock owner: %v", err)
	}
	ok, err = repo.LockOwner(po.GitOwnerTeamProject, 5, "user-2", 10*time.Minute)
	if err != nil {
		t.Fatalf("lock after unlock: %v", err)
	}
	if !ok {
		t.Fatal("lock after unlock = false, want true")
	}
}
