package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestZipLeadingSlashPaths(t *testing.T) {
	t.Run("accepts leading-slash source paths", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<html></html>"), 0644); err != nil {
			t.Fatalf("write source: %v", err)
		}

		r, err := Open(root)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		zipPath, err := r.Zip([]string{"/index.html"}, "page")
		if err != nil {
			t.Fatalf("Zip() error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "page.zip")); err != nil {
			t.Fatalf("archive missing: %v", err)
		}
		if filepath.Clean(zipPath) != "page.zip" && filepath.Clean(zipPath) != "/page.zip" {
			t.Errorf("ZipPath = %q, want root-relative page.zip", zipPath)
		}
	})

	t.Run("places output next to leading-slash nested source", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "docs"), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "docs", "a.txt"), []byte("a"), 0644); err != nil {
			t.Fatalf("write source: %v", err)
		}

		r, err := Open(root)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		if _, err := r.Zip([]string{"/docs/a.txt"}, "archive"); err != nil {
			t.Fatalf("Zip() error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "docs", "archive.zip")); err != nil {
			t.Fatalf("archive not next to source: %v", err)
		}
	})
}
