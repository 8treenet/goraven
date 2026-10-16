package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goraven/backend/po"
	"goraven/backend/vo"
	"goraven/util/git"
)

func TestGitRetryClonePreservesDepth(t *testing.T) {
	requireGitBinary(t)
	shallow, full := true, false
	for _, tc := range []struct {
		name    string
		shallow *bool
		commits string
	}{
		{"full history", &full, "2"},
		{"shallow", &shallow, "1"},
		{"default shallow", nil, "1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := newGitServiceTestEnv(t)
			bare := makeBareRemote(t)
			cloneAndPush(t, bare, "a.txt", "a\n")
			cloneAndPush(t, bare, "b.txt", "b\n")
			failed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			}))
			defer failed.Close()
			gctx := &GitContext{
				OwnerType: po.GitOwnerUserProject,
				OwnerId:   110,
				Dir:       t.TempDir(),
				IsOwner:   true,
			}
			wait := func(want uint8) *po.ProjectGitSetting {
				t.Helper()
				deadline := time.Now().Add(5 * time.Second)
				for time.Now().Before(deadline) {
					setting, err := svc.GitSettingRepo.GetByOwner(gctx.OwnerType, gctx.OwnerId)
					if err == nil && setting.CloneState == want && !git.CloneCancelRegistered(gitOwnerKey(gctx.OwnerType, gctx.OwnerId)) {
						return setting
					}
					time.Sleep(10 * time.Millisecond)
				}
				t.Fatalf("clone did not reach state %d", want)
				return nil
			}
			if err := svc.gitStartClone(gctx, &vo.GitCloneReq{
				RemoteUrl: failed.URL + "/repo.git", Shallow: tc.shallow,
			}); err != nil {
				t.Fatalf("start clone: %v", err)
			}
			setting := wait(po.GitCloneFailed)
			// Replace only the failed transport, retaining the persisted clone options.
			setting.RemoteUrl = "file://" + bare
			if err := svc.GitSettingRepo.Upsert(setting); err != nil {
				t.Fatal(err)
			}
			// The retry must recover its options from the database, not the request.
			retry := &FileManagerService{GitSettingRepo: svc.GitSettingRepo}
			if err := retry.gitRetryClone(gctx); err != nil {
				t.Fatalf("retry clone: %v", err)
			}
			wait(po.GitCloneSuccess)
			if got := hostGitOutput(t, gctx.Dir, "rev-list", "--count", "HEAD"); got != tc.commits {
				t.Errorf("retried commit count = %s, want %s", got, tc.commits)
			}
		})
	}
}
