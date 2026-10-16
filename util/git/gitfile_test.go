package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitfileCommand(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = baseEnv(t.TempDir())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestIsRepositoryGitfiles(t *testing.T) {
	if !BinaryAvailable() {
		t.Skip("git is unavailable")
	}
	t.Run("linked worktree", func(t *testing.T) {
		root := t.TempDir()
		gitfileCommand(t, root, "init", "-b", "main")
		gitfileCommand(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "initial")
		linked := filepath.Join(t.TempDir(), "linked")
		gitfileCommand(t, root, "worktree", "add", "-b", "linked", linked)
		if !IsRepository(linked) {
			t.Fatal("linked worktree is not recognized as a repository")
		}
	})
	t.Run("relative gitdir", func(t *testing.T) {
		root := t.TempDir()
		metadata := filepath.Join(root, "metadata")
		work := filepath.Join(root, "work")
		gitfileCommand(t, root, "init", "--separate-git-dir", metadata, work)
		if err := os.WriteFile(filepath.Join(work, ".git"), []byte("gitdir: ../metadata\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if !IsRepository(work) {
			t.Fatal("relative gitdir is not recognized as a repository")
		}
	})
	for _, contents := range []string{"not a gitfile\n", "gitdir: missing\n"} {
		t.Run(contents, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, ".git"), []byte(contents), 0644); err != nil {
				t.Fatal(err)
			}
			if IsRepository(dir) {
				t.Fatal("invalid gitfile is recognized as a repository")
			}
		})
	}
	t.Run("ordinary repository and child directory", func(t *testing.T) {
		root := t.TempDir()
		if IsRepository(root) {
			t.Fatal("empty directory is recognized as a repository")
		}
		gitfileCommand(t, root, "init", "-b", "main")
		if !IsRepository(root) {
			t.Fatal("ordinary repository is not recognized")
		}
		child := filepath.Join(root, "child")
		if err := os.Mkdir(child, 0755); err != nil {
			t.Fatal(err)
		}
		if IsRepository(child) {
			t.Fatal("child directory is mistaken for a repository root")
		}
	})
}

func TestInitializeAdoptsLinkedWorktree(t *testing.T) {
	if !BinaryAvailable() {
		t.Skip("git is unavailable")
	}
	root := t.TempDir()
	gitfileCommand(t, root, "init", "-b", "main")
	gitfileCommand(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "initial")
	linked := filepath.Join(t.TempDir(), "linked")
	gitfileCommand(t, root, "worktree", "add", "-b", "linked", linked)
	head := gitfileCommand(t, linked, "rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(linked, "uncommitted.txt"), []byte("keep uncommitted\n"), 0644); err != nil {
		t.Fatal(err)
	}
	repo := &Repo{Dir: linked}
	if _, err := repo.Initialize(context.Background(), "", BotIdentity); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if got := gitfileCommand(t, linked, "rev-parse", "HEAD"); got != head {
		t.Fatal("adopting a linked worktree created a commit")
	}
	if _, err := os.Stat(filepath.Join(linked, ".gitignore")); !os.IsNotExist(err) {
		t.Fatalf("adopting a linked worktree created a default gitignore: %v", err)
	}
}
