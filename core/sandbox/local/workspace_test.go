package local

import (
	"path/filepath"
	"testing"
)

// TestResolveWorkspace verifies that a relative user workspace is normalized
// to an absolute path.
//
// The shipped default config uses `user_space: "./data/users"`, so the
// workspace handed to the sandbox was relative, while validateFilePath
// requires an absolute path. Every write into the user space failed with
// "path must be absolute", which also broke the attachment upload.
func TestResolveWorkspace(t *testing.T) {
	cases := []struct {
		name string
		in   string
		keep string
	}{
		{name: "relative", in: "./data/users"},
		{name: "relative nested", in: "data/users/999"},
		{name: "absolute", in: "/goraven/data/users", keep: "/goraven/data/users"},
	}

	for _, c := range cases {
		got := resolveWorkspace(c.in)
		if !filepath.IsAbs(got) {
			t.Errorf("%s: resolveWorkspace(%q) = %q, expected an absolute path", c.name, c.in, got)
		}
		if c.keep != "" && got != c.keep {
			t.Errorf("%s: resolveWorkspace(%q) = %q, expected %q", c.name, c.in, got, c.keep)
		}
	}
}

// TestNewLocalSandbox_WorkspaceIsAbsolute verifies that the sandbox workspace
// is always an absolute path.
func TestNewLocalSandbox_WorkspaceIsAbsolute(t *testing.T) {
	box := NewLocalSandbox("999")
	ws := box.GetWorkspace()
	if !filepath.IsAbs(ws) {
		t.Fatalf("GetWorkspace() = %q, expected an absolute path", ws)
	}
}
