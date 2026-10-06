package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseStatusZUnmergedKeepsNextEntry(t *testing.T) {
	for _, xy := range []string{"DD", "AU", "UD", "UA", "DU", "AA", "UU"} {
		t.Run(xy, func(t *testing.T) {
			raw := xy + " a-conflict.txt\x00 M z-neighbor.txt\x00"
			changes := ParseStatusZ(raw)
			if len(changes) != 2 {
				t.Fatalf("changes = %#v, want conflict and neighbor", changes)
			}
			if changes[0].Path != "a-conflict.txt" || changes[0].Status != "C" || changes[0].Source != "" {
				t.Errorf("conflict = %#v, want C with no source", changes[0])
			}
			if changes[1].Path != "z-neighbor.txt" || changes[1].Status != "M" || changes[1].Source != "" {
				t.Errorf("neighbor = %#v, want independent modified entry", changes[1])
			}
		})
	}
}

func TestParseStatusZRenameAndCopySource(t *testing.T) {
	for _, xy := range []string{"R ", " R", "C ", " C"} {
		t.Run(xy, func(t *testing.T) {
			raw := xy + " new name.txt\x00old name.txt\x00 M z-neighbor.txt\x00"
			changes := ParseStatusZ(raw)
			if len(changes) != 2 {
				t.Fatalf("changes = %#v, want renamed/copied path and neighbor", changes)
			}
			if changes[0].Path != "new name.txt" || changes[0].Source != "old name.txt" {
				t.Errorf("rename/copy = %#v, want both paths", changes[0])
			}
			wantStatus := "M" // Preserve the existing copy display status.
			if strings.Contains(xy, "R") {
				wantStatus = "R"
			}
			if changes[0].Status != wantStatus {
				t.Errorf("status = %q, want %q", changes[0].Status, wantStatus)
			}
			if changes[1].Path != "z-neighbor.txt" || changes[1].Source != "" {
				t.Errorf("neighbor = %#v, want independent entry", changes[1])
			}
		})
	}
}

func TestCollectChangesWithMergeConflict(t *testing.T) {
	if !BinaryAvailable() {
		t.Skip("git binary is required")
	}
	auth, err := BuildBaseEnv()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(auth.Cleanup)
	repo := &Repo{Dir: t.TempDir(), Env: WithEnv(auth.Env,
		"GIT_AUTHOR_NAME", BotIdentity.AuthorName,
		"GIT_AUTHOR_EMAIL", BotIdentity.AuthorEmail,
		"GIT_COMMITTER_NAME", BotIdentity.CommitterName,
		"GIT_COMMITTER_EMAIL", BotIdentity.CommitterEmail,
	)}
	ctx := context.Background()
	run := func(args ...string) {
		t.Helper()
		if _, stderr, err := repo.run(ctx, args...); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, stderr)
		}
	}
	write := func(name, contents string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(repo.Dir, name), []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}
	run("init", "-b", "main")
	write("a-conflict.txt", "base\n")
	write("z-neighbor.txt", "base\n")
	run("add", "--", "a-conflict.txt", "z-neighbor.txt")
	run("commit", "-m", "base")
	run("checkout", "-b", "side")
	write("a-conflict.txt", "side\n")
	run("commit", "-am", "side")
	run("checkout", "main")
	write("a-conflict.txt", "main\n")
	run("commit", "-am", "main")
	if _, _, err := repo.run(ctx, "merge", "--no-edit", "side"); err == nil {
		t.Fatal("merge unexpectedly succeeded")
	}
	write("z-neighbor.txt", "independent change\n")
	raw, stderr, err := repo.run(ctx, "status", "--porcelain", "-z", "-uall")
	if err != nil || !strings.Contains(raw, "UU a-conflict.txt\x00") || !strings.Contains(raw, " M z-neighbor.txt\x00") {
		t.Fatalf("unexpected native conflict status: %q: %v: %s", raw, err, stderr)
	}
	changes, err := repo.CollectChanges(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 {
		t.Fatalf("CollectChanges = %#v, want conflict and modified neighbor", changes)
	}
	if changes[0].Path != "a-conflict.txt" || changes[0].Status != "C" || changes[0].Source != "" {
		t.Errorf("conflict = %#v, want C without a source path", changes[0])
	}
	if changes[1].Path != "z-neighbor.txt" || changes[1].Status != "M" {
		t.Errorf("modified neighbor = %#v", changes[1])
	}
}
