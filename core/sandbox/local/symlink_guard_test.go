package local

import (
	"os"
	"path/filepath"
	"testing"

	"goraven/core/sandbox/types"
)

func newTestWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "documents"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "documents", "a.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCheckContainmentPlainPaths(t *testing.T) {
	ws := newTestWorkspace(t)

	if err := checkContainment(filepath.Join(ws, "documents"), []string{ws}); err != nil {
		t.Errorf("inside path should pass: err=%v", err)
	}
	if err := checkContainment(filepath.Join(ws, "missing", "file.txt"), []string{ws}); err != nil {
		t.Errorf("missing file inside workspace should pass, err=%v", err)
	}
	if err := checkContainment("/etc/passwd", []string{ws}); err == nil {
		t.Errorf("outside path should fail")
	}
}

func TestCheckContainmentSymlinkEscape(t *testing.T) {
	ws := newTestWorkspace(t)

	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("s"), 0644); err != nil {
		t.Fatal(err)
	}
	// ws/link -> outside
	if err := os.Symlink(outside, filepath.Join(ws, "link")); err != nil {
		t.Fatal(err)
	}

	// 直接触及链接本身
	if err := checkContainment(filepath.Join(ws, "link"), []string{ws}); err == nil {
		t.Errorf("symlink escaping workspace should fail")
	}
	// 透过链接触及内部文件
	if err := checkContainment(filepath.Join(ws, "link", "secret.txt"), []string{ws}); err == nil {
		t.Errorf("file through escaping symlink should fail")
	}
}

func TestCheckContainmentSymlinkInsideWorkspace(t *testing.T) {
	ws := newTestWorkspace(t)
	// documents -> docs（工作区内链接应当放行）
	if err := os.Symlink(filepath.Join(ws, "documents"), filepath.Join(ws, "docs")); err != nil {
		t.Fatal(err)
	}
	if err := checkContainment(filepath.Join(ws, "docs", "a.txt"), []string{ws}); err != nil {
		t.Errorf("internal symlink should pass, err=%v", err)
	}
}

func TestSandboxValidateFilePathRejectsSymlinkEscape(t *testing.T) {
	ws := newTestWorkspace(t)
	sb := &LocalSandbox{userName: "u", Workspace: ws}

	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(ws, "escape")); err != nil {
		t.Fatal(err)
	}

	if _, err := sb.validateFilePath(filepath.Join(ws, "escape")); err == nil {
		t.Errorf("validateFilePath should reject escaping symlink")
	}
	if _, err := sb.validateFilePath(filepath.Join(ws, "documents")); err != nil {
		t.Errorf("validateFilePath should accept normal path, err=%v", err)
	}
	// 尚不存在的文件（写入场景）也应放行
	if _, err := sb.validateFilePath(filepath.Join(ws, "documents", "new", "b.txt")); err != nil {
		t.Errorf("validateFilePath should accept missing file inside workspace, err=%v", err)
	}
}

func TestFileManagerRejectsSymlinkEscape(t *testing.T) {
	ws := newTestWorkspace(t)
	fm := NewLocalFileManager(&types.FileManagerConfig{UserName: "u", Workspace: ws})

	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("s"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(ws, "escape")); err != nil {
		t.Fatal(err)
	}

	if _, err := fm.resolvePath("escape/secret.txt"); err == nil {
		t.Errorf("resolvePath should reject file through escaping symlink")
	}
	if _, err := fm.resolvePath("documents/a.txt"); err != nil {
		t.Errorf("resolvePath should accept normal relative path, err=%v", err)
	}
	if _, err := fm.resolvePath("../../../etc/passwd"); err == nil {
		t.Errorf("resolvePath should reject traversal")
	}
}

func TestGuardedBackendValidate(t *testing.T) {
	ws := newTestWorkspace(t)
	g := newGuardedBackend(nil, ws)

	if got, err := g.validate(""); err != nil || got != ws {
		t.Errorf("empty path should resolve to workspace, got %q err=%v", got, err)
	}
	if got, err := g.validate("documents/a.txt"); err != nil || got != filepath.Join(ws, "documents", "a.txt") {
		t.Errorf("relative path should resolve inside workspace, got %q err=%v", got, err)
	}
	if _, err := g.validate("/etc/passwd"); err == nil {
		t.Errorf("absolute outside path should fail")
	}

	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(ws, "escape")); err != nil {
		t.Fatal(err)
	}
	if _, err := g.validate("escape/x.txt"); err == nil {
		t.Errorf("symlink escape should fail")
	}
}
