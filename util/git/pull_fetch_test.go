package git

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRepoPullWithoutUpstream(t *testing.T) {
	if !BinaryAvailable() {
		t.Skip("git binary not available")
	}
	for _, kind := range []string{"missing remote", "empty remote", "matching history", "unrelated history"} {
		t.Run(kind, func(t *testing.T) {
			auth, err := BuildBaseEnv()
			if err != nil {
				t.Fatal(err)
			}
			defer auth.Cleanup()
			env := WithEnv(auth.Env, "GIT_AUTHOR_NAME", "test", "GIT_AUTHOR_EMAIL", "test@example.com",
				"GIT_COMMITTER_NAME", "test", "GIT_COMMITTER_EMAIL", "test@example.com")
			run := func(dir string, args ...string) {
				t.Helper()
				cmd := exec.Command("git", args...)
				cmd.Dir, cmd.Env = dir, env
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("git %v: %v: %s", args, err, out)
				}
			}
			dir := t.TempDir()
			run(dir, "init", "-b", "main")
			run(dir, "commit", "--allow-empty", "-m", "local history")
			remote := filepath.Join(t.TempDir(), "remote.git")
			if kind != "missing remote" {
				run(dir, "init", "--bare", "--initial-branch=main", remote)
			}
			run(dir, "remote", "add", "origin", remote)
			switch kind {
			case "matching history":
				run(dir, "push", "origin", "main") // Deliberately do not establish upstream.
			case "unrelated history":
				seed := t.TempDir()
				run(seed, "init", "-b", "main")
				run(seed, "commit", "--allow-empty", "-m", "unrelated history")
				run(seed, "push", remote, "main")
			}
			repo := &Repo{Dir: dir, Env: env}
			if upstream := repo.UpstreamName(context.Background()); upstream != "" {
				t.Fatalf("unexpected upstream %q", upstream)
			}
			err = repo.Pull(context.Background(), remote)
			switch kind {
			case "missing remote":
				var commandErr *CommandError
				if !errors.As(err, &commandErr) {
					t.Fatalf("Pull() = %v, want fetch CommandError", err)
				}
			case "unrelated history":
				if !errors.Is(err, ErrUnrelatedHistory) {
					t.Fatalf("Pull() = %v, want ErrUnrelatedHistory", err)
				}
			default:
				if err != nil {
					t.Fatalf("Pull() = %v, want nil", err)
				}
			}
		})
	}
}

type fetchFailureExecutor struct {
	calls [][]string
}

func (e *fetchFailureExecutor) Run(_ context.Context, _ string, _ []string, args ...string) (string, string, error) {
	e.calls = append(e.calls, append([]string(nil), args...))
	switch args[0] {
	case "status":
		return "", "", nil
	case "branch":
		return "main", "", nil
	case "rev-parse":
		return "", "no upstream configured", errors.New("no upstream")
	case "fetch":
		return "", "fatal: Authentication failed", errors.New("fetch failed")
	default:
		return "", "", errors.New("unexpected command")
	}
}

func TestRepoRemoteHistoryPreservesFetchError(t *testing.T) {
	if !BinaryAvailable() {
		t.Skip("git binary not available")
	}
	for _, operation := range []string{"pull", "push"} {
		t.Run(operation, func(t *testing.T) {
			executor := &fetchFailureExecutor{}
			repo := &Repo{Executor: executor}
			var err error
			if operation == "pull" {
				err = repo.Pull(context.Background(), "https://example.com/repo.git")
			} else {
				err = repo.Push(context.Background(), "https://example.com/repo.git")
			}
			if !errors.Is(err, ErrAuthFailed) {
				t.Fatalf("%s = %v, want ErrAuthFailed", operation, err)
			}
			if last := executor.calls[len(executor.calls)-1][0]; last != "fetch" {
				t.Fatalf("continued after failed fetch: %v", executor.calls)
			}
		})
	}
}
