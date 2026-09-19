package git

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEnsureDefaultGitignore(t *testing.T) {
	t.Run("creates default gitignore when missing", func(t *testing.T) {
		dir := t.TempDir()
		created, err := EnsureDefaultGitignore(dir)
		if err != nil {
			t.Fatalf("EnsureDefaultGitignore() error = %v", err)
		}
		if !created {
			t.Fatal("EnsureDefaultGitignore() created = false, want true")
		}
		data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
		if err != nil {
			t.Fatalf("read .gitignore: %v", err)
		}
		if string(data) != DefaultGitignoreContent {
			t.Fatalf(".gitignore content = %q, want default content", string(data))
		}
	})

	t.Run("never overwrites existing gitignore", func(t *testing.T) {
		dir := t.TempDir()
		custom := "# custom\nbuild/\n"
		if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(custom), 0644); err != nil {
			t.Fatalf("write custom .gitignore: %v", err)
		}
		created, err := EnsureDefaultGitignore(dir)
		if err != nil {
			t.Fatalf("EnsureDefaultGitignore() error = %v", err)
		}
		if created {
			t.Fatal("EnsureDefaultGitignore() created = true, want false")
		}
		data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
		if err != nil {
			t.Fatalf("read .gitignore: %v", err)
		}
		if string(data) != custom {
			t.Fatalf(".gitignore content = %q, want existing content preserved", string(data))
		}
	})
}

func TestParseStatusZ(t *testing.T) {
	raw := strings.Join([]string{
		"?? new.txt",
		" M modified.txt",
		"A  added.txt",
		" D deleted.txt",
		"R  new-name.txt",
		"old-name.txt",
		"",
	}, "\x00")

	changes := ParseStatusZ(raw)
	if len(changes) != 5 {
		t.Fatalf("len(changes) = %d, want 5: %#v", len(changes), changes)
	}
	want := []struct {
		path   string
		status string
	}{
		{"new.txt", "U"},
		{"modified.txt", "M"},
		{"added.txt", "A"},
		{"deleted.txt", "D"},
		{"new-name.txt", "R"},
	}
	for i, w := range want {
		if changes[i].Path != w.path || changes[i].Status != w.status {
			t.Errorf("changes[%d] = {%q,%q}, want {%q,%q}", i, changes[i].Path, changes[i].Status, w.path, w.status)
		}
	}
	if changes[4].Source != "old-name.txt" {
		t.Errorf("rename source = %q, want old-name.txt", changes[4].Source)
	}
}

func TestFilterLargeChanges(t *testing.T) {
	dir := t.TempDir()
	writeSize := func(name string, size int) {
		if err := os.WriteFile(filepath.Join(dir, name), make([]byte, size), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	writeSize("small.txt", 4)
	writeSize("big.bin", 16)

	changes := []Change{
		{Path: "small.txt", Status: "M"},
		{Path: "big.bin", Status: "A"},
		{Path: "gone.txt", Status: "D"},
	}

	const limit = 10
	keep, skipped := FilterLargeChanges(dir, changes, limit)
	if len(keep) != 2 {
		t.Fatalf("len(keep) = %d, want 2: %#v", len(keep), keep)
	}
	if len(skipped) != 1 || skipped[0].Path != "big.bin" {
		t.Fatalf("skipped = %#v, want big.bin", skipped)
	}
	if skipped[0].Size != 16 {
		t.Errorf("skipped size = %d, want 16", skipped[0].Size)
	}
	if keep[1].Path != "gone.txt" {
		t.Errorf("deleted file should not be skipped: %#v", keep)
	}
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
			if got := NextDailySyncAt(c.now); !got.Equal(c.want) {
				t.Errorf("NextDailySyncAt(%v) = %v, want %v", c.now, got, c.want)
			}
		})
	}
}

func TestSanitizeMessage(t *testing.T) {
	t.Run("strips url userinfo", func(t *testing.T) {
		msg := "fatal: could not read Username for 'https://user:pass@example.com': terminal prompts disabled"
		got := SanitizeMessage(msg)
		if strings.Contains(got, "user:pass") {
			t.Fatalf("sanitized message still contains credentials: %q", got)
		}
		if !strings.Contains(got, "https://***@example.com") {
			t.Fatalf("sanitized message = %q, want masked url", got)
		}
	})

	t.Run("masks known secrets", func(t *testing.T) {
		got := SanitizeMessage("auth failed for token ghp_supersecret", "ghp_supersecret")
		if strings.Contains(got, "ghp_supersecret") {
			t.Fatalf("sanitized message still contains secret: %q", got)
		}
	})

	t.Run("keeps scp-like urls untouched", func(t *testing.T) {
		msg := "fatal: Could not read from remote repository git@github.com:org/repo.git"
		if got := SanitizeMessage(msg); got != msg {
			t.Fatalf("sanitized message = %q, want unchanged %q", got, msg)
		}
	})
}

func TestValidateRemoteURL(t *testing.T) {
	valid := []string{
		"https://github.com/org/repo.git",
		"http://127.0.0.1:8080/repo.git",
		"git@github.com:org/repo.git",
		"ssh://git@github.com/org/repo.git",
		"file:///tmp/repo.git",
		"/tmp/repo.git",
	}
	for _, url := range valid {
		if err := ValidateRemoteURL(url); err != nil {
			t.Errorf("ValidateRemoteURL(%q) = %v, want nil", url, err)
		}
	}

	invalid := []string{
		"",
		"-u evil",
		"https://user:pass@github.com/org/repo.git",
		"ext::sh -c 'touch /tmp/pwned'",
		"foo::bar",
		"http://a\nb",
	}
	for _, url := range invalid {
		if err := ValidateRemoteURL(url); !errors.Is(err, ErrInvalidRemoteURL) {
			t.Errorf("ValidateRemoteURL(%q) = %v, want ErrInvalidRemoteURL", url, err)
		}
	}
}

func TestBuildAuthEnv(t *testing.T) {
	findEnv := func(env []string, key string) string {
		prefix := key + "="
		for _, entry := range env {
			if strings.HasPrefix(entry, prefix) {
				return strings.TrimPrefix(entry, prefix)
			}
		}
		return ""
	}
	keyPathFromCommand := func(sshCmd string) string {
		fields := strings.Fields(sshCmd)
		for i, field := range fields {
			if field == "-i" && i+1 < len(fields) {
				return strings.Trim(fields[i+1], "'")
			}
		}
		return ""
	}
	assertSSHIsolation := func(t *testing.T, sshCmd string) {
		t.Helper()
		for _, want := range []string{
			"-F /dev/null",
			"IdentitiesOnly=yes",
			"IdentityFile=none",
			"IdentityAgent=none",
			"BatchMode=yes",
			"StrictHostKeyChecking=accept-new",
		} {
			if !strings.Contains(sshCmd, want) {
				t.Errorf("GIT_SSH_COMMAND = %q, missing %q", sshCmd, want)
			}
		}
	}

	t.Run("ssh injects key file and known hosts", func(t *testing.T) {
		secretDir := t.TempDir()
		setting := &AuthSetting{
			OwnerType:     1,
			OwnerId:       7,
			AuthType:      AuthSSH,
			SshPrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nfake\n-----END OPENSSH PRIVATE KEY-----",
		}
		auth, err := BuildAuthEnv(secretDir, setting)
		if err != nil {
			t.Fatalf("BuildAuthEnv() error = %v", err)
		}
		defer auth.Cleanup()

		sshCmd := findEnv(auth.Env, "GIT_SSH_COMMAND")
		if !strings.Contains(sshCmd, "-i ") {
			t.Fatalf("GIT_SSH_COMMAND = %q", sshCmd)
		}
		assertSSHIsolation(t, sshCmd)
		keyPath := keyPathFromCommand(sshCmd)
		if keyPath == "" {
			t.Fatalf("no -i key in GIT_SSH_COMMAND = %q", sshCmd)
		}
		info, err := os.Stat(keyPath)
		if err != nil {
			t.Fatalf("stat key file: %v", err)
		}
		if info.Mode().Perm() != 0600 {
			t.Errorf("key file mode = %o, want 0600", info.Mode().Perm())
		}
		knownHosts := filepath.Join(secretDir, "known_hosts", "1_7")
		if _, err := os.Stat(knownHosts); err != nil {
			t.Errorf("known_hosts not created: %v", err)
		}
		if home := findEnv(auth.Env, "HOME"); home == "" {
			t.Error("HOME not set to isolated directory")
		}
		if findEnv(auth.Env, "GIT_TERMINAL_PROMPT") != "0" {
			t.Error("GIT_TERMINAL_PROMPT != 0")
		}
		if findEnv(auth.Env, "GIT_CONFIG_NOSYSTEM") != "1" {
			t.Error("GIT_CONFIG_NOSYSTEM != 1")
		}
		if findEnv(auth.Env, "GIT_CONFIG_GLOBAL") != "/dev/null" {
			t.Error("GIT_CONFIG_GLOBAL != /dev/null")
		}

		homeDir := findEnv(auth.Env, "HOME")
		auth.Cleanup()
		if _, err := os.Stat(homeDir); !os.IsNotExist(err) {
			t.Errorf("temp auth dir still exists after cleanup: %v", err)
		}
	})

	t.Run("ssh normalizes pasted key endings", func(t *testing.T) {
		secretDir := t.TempDir()
		pasted := "-----BEGIN OPENSSH PRIVATE KEY-----\r\n  ZmFrZS1rZXktYm9keQ==  \r\n-----END OPENSSH PRIVATE KEY-----"
		setting := &AuthSetting{
			OwnerType:     1,
			OwnerId:       8,
			AuthType:      AuthSSH,
			SshPrivateKey: pasted,
		}
		auth, err := BuildAuthEnv(secretDir, setting)
		if err != nil {
			t.Fatalf("BuildAuthEnv() error = %v", err)
		}
		defer auth.Cleanup()

		sshCmd := findEnv(auth.Env, "GIT_SSH_COMMAND")
		keyPath := keyPathFromCommand(sshCmd)
		if keyPath == "" {
			t.Fatalf("no -i key in GIT_SSH_COMMAND = %q", sshCmd)
		}
		data, err := os.ReadFile(keyPath)
		if err != nil {
			t.Fatalf("read key file: %v", err)
		}
		want := "-----BEGIN OPENSSH PRIVATE KEY-----\nZmFrZS1rZXktYm9keQ==\n-----END OPENSSH PRIVATE KEY-----\n"
		if string(data) != want {
			t.Fatalf("key file = %q, want %q", string(data), want)
		}
	})

	t.Run("https injects askpass without argv secrets", func(t *testing.T) {
		secretDir := t.TempDir()
		setting := &AuthSetting{
			OwnerType:     2,
			OwnerId:       3,
			AuthType:      AuthHTTPS,
			HttpsUsername: "bot",
			HttpsSecret:   "s3cr3t-token",
		}
		auth, err := BuildAuthEnv(secretDir, setting)
		if err != nil {
			t.Fatalf("BuildAuthEnv() error = %v", err)
		}
		defer auth.Cleanup()

		askpass := findEnv(auth.Env, "GIT_ASKPASS")
		if askpass == "" {
			t.Fatal("GIT_ASKPASS not set")
		}
		info, err := os.Stat(askpass)
		if err != nil {
			t.Fatalf("stat askpass: %v", err)
		}
		if info.Mode().Perm() != 0700 {
			t.Errorf("askpass mode = %o, want 0700", info.Mode().Perm())
		}
		if findEnv(auth.Env, "GORAVEN_GIT_PASSWORD") != "s3cr3t-token" {
			t.Error("password env not injected")
		}
		if findEnv(auth.Env, "GORAVEN_GIT_USERNAME") != "bot" {
			t.Error("username env not injected")
		}
	})

	t.Run("public repo isolates ssh from host keys", func(t *testing.T) {
		secretDir := t.TempDir()
		setting := &AuthSetting{OwnerType: 1, OwnerId: 1}
		auth, err := BuildAuthEnv(secretDir, setting)
		if err != nil {
			t.Fatalf("BuildAuthEnv() error = %v", err)
		}
		defer auth.Cleanup()

		for _, name := range []string{"GIT_ASKPASS", "GORAVEN_GIT_USERNAME", "GORAVEN_GIT_PASSWORD"} {
			if findEnv(auth.Env, name) != "" {
				t.Errorf("public repo should not inject %s", name)
			}
		}
		sshCmd := findEnv(auth.Env, "GIT_SSH_COMMAND")
		assertSSHIsolation(t, sshCmd)
		if strings.Contains(sshCmd, "-i ") {
			t.Errorf("public repo must not pass a key: %q", sshCmd)
		}
		if _, err := os.Stat(filepath.Join(secretDir, "known_hosts", "1_1")); err != nil {
			t.Errorf("known_hosts not created for public repo: %v", err)
		}
	})

	t.Run("missing credentials rejected", func(t *testing.T) {
		secretDir := t.TempDir()
		if _, err := BuildAuthEnv(secretDir, &AuthSetting{AuthType: AuthSSH}); !errors.Is(err, ErrNotConfigured) {
			t.Errorf("ssh without key = %v, want ErrNotConfigured", err)
		}
		if _, err := BuildAuthEnv(secretDir, &AuthSetting{AuthType: AuthHTTPS}); !errors.Is(err, ErrNotConfigured) {
			t.Errorf("https without secret = %v, want ErrNotConfigured", err)
		}
	})
}

func TestWithEnv(t *testing.T) {
	env := []string{"PATH=/usr/bin", "GIT_AUTHOR_NAME=old"}
	got := WithEnv(env, "GIT_AUTHOR_NAME", "new", "GIT_AUTHOR_EMAIL", "a@b.c")
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3: %#v", len(got), got)
	}
	if got[1] != "GIT_AUTHOR_NAME=new" {
		t.Errorf("existing key not replaced in place: %q", got[1])
	}
	if env[1] != "GIT_AUTHOR_NAME=old" {
		t.Error("WithEnv mutated the input slice")
	}
}

func TestStderrClassification(t *testing.T) {
	cases := []struct {
		name   string
		stderr string
		check  func(string) bool
		want   bool
	}{
		{"non-fast-forward", "! [rejected]        main -> main (non-fast-forward)\nerror: failed to push some refs", IsPushRejected, true},
		{"fetch first", "! [rejected]        main -> main (fetch first)", IsPushRejected, true},
		{"rebase conflict", "CONFLICT (content): Merge conflict in a.txt\nerror: could not apply 1234567", IsConflict, true},
		{"auth failed https", "fatal: Authentication failed for 'https://example.com/repo.git/'", IsAuthError, true},
		{"auth prompt disabled", "fatal: could not read Username for 'https://example.com': terminal prompts disabled", IsAuthError, true},
		{"auth publickey", "git@github.com: Permission denied (publickey).\nfatal: Could not read from remote repository.", IsAuthError, true},
		{"dns failure is not auth", "fatal: unable to access 'https://example.com/repo.git/': Could not resolve host: example.com", IsAuthError, false},
		{"rejected is not auth", "! [rejected] main -> main (non-fast-forward)", IsAuthError, false},
		{"unrelated history", "fatal: refusing to merge unrelated histories", IsUnrelatedHistoryError, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.check(c.stderr); got != c.want {
				t.Errorf("classification = %v, want %v for %q", got, c.want, c.stderr)
			}
		})
	}
}

func TestRepoLogEmptyHistory(t *testing.T) {
	if !BinaryAvailable() {
		t.Skip("git binary not available")
	}
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", "-b", "main", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	env, err := BuildBaseEnv()
	if err != nil {
		t.Fatalf("BuildBaseEnv() error = %v", err)
	}
	defer env.Cleanup()

	repo := &Repo{Dir: dir, Env: env.Env}
	items, err := repo.Log(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("Log() on unborn branch error = %v, want nil", err)
	}
	if len(items) != 0 {
		t.Fatalf("Log() = %#v, want empty history", items)
	}
}

func TestCloneRegistry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	key := "t:1"
	RegisterCloneCancel(key, cancel)
	if !CloneCancelRegistered(key) {
		t.Fatal("handler not registered")
	}
	CancelClone(key)
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("registered clone context was not canceled")
	}
	if CloneCancelRegistered(key) {
		t.Fatal("cancel registry entry not removed")
	}

	// 重复注册应取消旧句柄
	first, firstCancel := context.WithCancel(context.Background())
	RegisterCloneCancel(key, firstCancel)
	second, secondCancel := context.WithCancel(context.Background())
	RegisterCloneCancel(key, secondCancel)
	select {
	case <-first.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("old cancel handler was not canceled on re-register")
	}
	select {
	case <-second.Done():
		t.Fatal("new cancel handler must stay alive after re-register")
	default:
	}
	UnregisterCloneCancel(key)
	if CloneCancelRegistered(key) {
		t.Fatal("unregister did not remove entry")
	}
	secondCancel()
}
