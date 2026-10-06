package git

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func newPathspecTestRepo(t *testing.T) *Repo {
	t.Helper()
	if !BinaryAvailable() {
		t.Skip("git binary is required")
	}
	auth, err := BuildBaseEnv()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(auth.Cleanup)
	repo := &Repo{Dir: t.TempDir(), Env: auth.Env}
	if _, stderr, err := repo.run(context.Background(), "init", "-b", "main"); err != nil {
		t.Fatalf("git init: %v: %s", err, stderr)
	}
	return repo
}

func TestCommitWorkingTreeKeepsFilteredFilesOutOfGlobPath(t *testing.T) {
	for _, name := range []string{"[a].txt", ":(glob)*.txt"} {
		t.Run(name, func(t *testing.T) {
			if runtime.GOOS == "windows" && strings.Contains(name, ":") {
				t.Skip("Windows does not permit colons in filenames")
			}
			checkFilteredPath(t, name)
		})
	}
}

func checkFilteredPath(t *testing.T, name string) {
	t.Helper()
	repo := newPathspecTestRepo(t)
	if err := os.WriteFile(filepath.Join(repo.Dir, name), []byte("small\n"), 0644); err != nil {
		t.Fatal(err)
	}
	large, err := os.Create(filepath.Join(repo.Dir, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if err := large.Truncate(LargeFileLimit + 1); err != nil {
		large.Close()
		t.Fatal(err)
	}
	if err := large.Close(); err != nil {
		t.Fatal(err)
	}
	_, skipped, err := repo.CommitWorkingTree(context.Background(), BotIdentity, "small file only")
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 1 || skipped[0].Path != "a.txt" {
		t.Fatalf("skipped = %#v, want only a.txt", skipped)
	}
	stdout, stderr, err := repo.run(context.Background(), "ls-tree", "--name-only", "HEAD")
	if err != nil {
		t.Fatalf("git ls-tree: %v: %s", err, stderr)
	}
	if strings.TrimSpace(stdout) != name {
		t.Fatalf("committed paths = %q; skipped a.txt must not be committed", stdout)
	}
}

func TestDiffTreatsFilePathLiterally(t *testing.T) {
	repo := newPathspecTestRepo(t)
	ctx := context.Background()
	for _, name := range []string{"[a].txt", "a.txt"} {
		if err := os.WriteFile(filepath.Join(repo.Dir, name), []byte("base\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := repo.CommitWorkingTree(ctx, BotIdentity, "base"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"[a].txt", "a.txt"} {
		if err := os.WriteFile(filepath.Join(repo.Dir, name), []byte("changed "+name+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	check := func(t *testing.T, commit string) {
		t.Helper()
		patch, err := repo.Diff(ctx, "[a].txt", commit, DiffMaxChars)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(patch, "+changed [a].txt") {
			t.Fatalf("literal file is absent from diff: %s", patch)
		}
		if strings.Contains(patch, "+changed a.txt") {
			t.Fatalf("diff for [a].txt also includes a.txt: %s", patch)
		}
		all, err := repo.Diff(ctx, "", commit, DiffMaxChars)
		if err != nil || !strings.Contains(all, "+changed a.txt") {
			t.Fatalf("unfiltered diff no longer includes a.txt: %v: %s", err, all)
		}
	}
	t.Run("working tree", func(t *testing.T) { check(t, "") })
	head, _, err := repo.CommitWorkingTree(ctx, BotIdentity, "both changes")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("commit", func(t *testing.T) { check(t, head) })
}
