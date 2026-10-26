package service

import (
	"errors"
	"testing"
	"time"

	"goraven/backend/po"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/util/git"
)

func TestValidateCloneRequest(t *testing.T) {
	service := &FileManagerService{}

	t.Run("nil payload rejected", func(t *testing.T) {
		if err := service.gitValidateCloneRequest(nil); !errors.Is(err, errs.ErrGitNotConfigured) {
			t.Errorf("nil req = %v, want ErrGitNotConfigured", err)
		}
	})

	t.Run("local mode without remote accepted", func(t *testing.T) {
		if err := service.gitValidateCloneRequest(&vo.GitCloneReq{Mode: vo.GitModeLocal}); err != nil {
			t.Errorf("local mode = %v, want nil", err)
		}
	})

	t.Run("local mode with remoteUrl rejected", func(t *testing.T) {
		err := service.gitValidateCloneRequest(&vo.GitCloneReq{Mode: vo.GitModeLocal, RemoteUrl: "https://github.com/org/repo.git"})
		if err == nil || errors.Is(err, errs.ErrGitInvalidRemoteUrl) {
			t.Errorf("local mode with remoteUrl = %v, want format error", err)
		}
	})

	t.Run("clone mode requires remote url and credentials", func(t *testing.T) {
		if err := service.gitValidateCloneRequest(&vo.GitCloneReq{}); !errors.Is(err, errs.ErrGitInvalidRemoteUrl) {
			t.Errorf("clone without url = %v, want ErrGitInvalidRemoteUrl", err)
		}
		ssh := &vo.GitCloneReq{RemoteUrl: "https://github.com/org/repo.git", AuthType: po.GitAuthSSH}
		if err := service.gitValidateCloneRequest(ssh); !errors.Is(err, errs.ErrGitNotConfigured) {
			t.Errorf("ssh clone without key = %v, want ErrGitNotConfigured", err)
		}
	})

	t.Run("missing mode defaults to clone", func(t *testing.T) {
		if (&vo.GitCloneReq{}).IsLocal() {
			t.Error("empty mode should not be local")
		}
		if !(&vo.GitCloneReq{Mode: vo.GitModeLocal}).IsLocal() {
			t.Error("mode local should be local")
		}
		var nilReq *vo.GitCloneReq
		if nilReq.IsLocal() {
			t.Error("nil req should not be local")
		}
	})
}

func TestGitPermissionMatrix(t *testing.T) {
	personalOwner := &GitContext{OwnerType: po.GitOwnerUserProject, IsOwner: true}
	if err := personalOwner.authorize(false); err != nil {
		t.Errorf("personal owner read authorize() = %v, want nil", err)
	}
	if err := personalOwner.authorize(true); err != nil {
		t.Errorf("personal owner config authorize() = %v, want nil", err)
	}

	personalOther := &GitContext{OwnerType: po.GitOwnerUserProject, IsOwner: false}
	if err := personalOther.authorize(false); !errors.Is(err, errs.ErrGitPermission) {
		t.Errorf("non-owner authorize() = %v, want ErrGitPermission", err)
	}

	teamCreator := &GitContext{OwnerType: po.GitOwnerTeamProject, IsCreator: true}
	if err := teamCreator.authorize(true); err != nil {
		t.Errorf("team creator config authorize() = %v, want nil", err)
	}

	teamMember := &GitContext{OwnerType: po.GitOwnerTeamProject, IsMember: true}
	if err := teamMember.authorize(false); err != nil {
		t.Errorf("team member write authorize() = %v, want nil", err)
	}
	if err := teamMember.authorize(true); !errors.Is(err, errs.ErrGitPermission) {
		t.Errorf("team member config authorize() = %v, want ErrGitPermission", err)
	}

	teamOutsider := &GitContext{OwnerType: po.GitOwnerTeamProject}
	if err := teamOutsider.authorize(false); !errors.Is(err, errs.ErrGitPermission) {
		t.Errorf("team outsider authorize() = %v, want ErrGitPermission", err)
	}
}

func TestGitOperationErrorMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want error
	}{
		{"no changes", errs.ErrGitNoChanges, errs.ErrGitNoChanges},
		{"dirty tree", errs.ErrGitDirtyTree, errs.ErrGitDirtyTree},
		{"not configured", errs.ErrGitNotConfigured, errs.ErrGitNotConfigured},
		{"auth failed", errs.ErrGitAuthFailed, errs.ErrGitAuthFailed},
		{"push rejected", errs.ErrGitPushRejected, errs.ErrGitPushRejected},
		{"unrelated history", errs.ErrGitUnrelatedHistory, errs.ErrGitUnrelatedHistory},
		{"needs attention", errs.ErrGitNeedsAttention, errs.ErrGitNeedsAttention},
		{"binary missing", errs.ErrGitBinaryMissing, errs.ErrGitBinaryMissing},
		{"invalid remote url", errs.ErrGitInvalidRemoteUrl, errs.ErrGitInvalidRemoteUrl},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := gitOperationError(c.err); got != c.want {
				t.Errorf("gitOperationError(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}

	t.Run("nil passthrough", func(t *testing.T) {
		if gitOperationError(nil) != nil {
			t.Error("gitOperationError(nil) should be nil")
		}
	})

	t.Run("unknown error passthrough", func(t *testing.T) {
		unknown := errors.New("boom")
		if got := gitOperationError(unknown); got != unknown {
			t.Errorf("gitOperationError(unknown) = %v, want unchanged", got)
		}
	})
}

func TestNextDailySyncAt(t *testing.T) {
	loc := time.Local
	cases := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "before 3am same day",
			now:  time.Date(2026, 9, 12, 1, 30, 0, 0, loc),
			want: time.Date(2026, 9, 12, 3, 0, 0, 0, loc),
		},
		{
			name: "after 3am next day",
			now:  time.Date(2026, 9, 12, 4, 0, 0, 0, loc),
			want: time.Date(2026, 9, 13, 3, 0, 0, 0, loc),
		},
		{
			name: "exactly 3am next day",
			now:  time.Date(2026, 9, 12, 3, 0, 0, 0, loc),
			want: time.Date(2026, 9, 13, 3, 0, 0, 0, loc),
		},
		{
			name: "crosses month boundary",
			now:  time.Date(2026, 9, 30, 23, 0, 0, 0, loc),
			want: time.Date(2026, 10, 1, 3, 0, 0, 0, loc),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := git.NextDailySyncAt(c.now); !got.Equal(c.want) {
				t.Errorf("NextDailySyncAt(%v) = %v, want %v", c.now, got, c.want)
			}
		})
	}
}
