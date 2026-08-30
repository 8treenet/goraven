package service

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"goraven/backend/vo/errs"
)

func TestResolveSharedAkPath(t *testing.T) {
	teamRoot := t.TempDir()
	projectDir := filepath.Join(teamRoot, "天气预报")
	if err := os.MkdirAll(filepath.Join(projectDir, "docs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "北京未来2天天气预报.docx"), []byte("doc"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "docs", "readme.md"), []byte("md"), 0644); err != nil {
		t.Fatal(err)
	}

	const (
		name      = "天气预报"
		boundFile = "/projects/天气预报/北京未来2天天气预报.docx"
	)

	t.Run("file 类型精确匹配", func(t *testing.T) {
		got, err := resolveSharedAkPath(teamRoot, boundFile, "file", boundFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join(projectDir, "北京未来2天天气预报.docx"); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("file 类型路径不匹配", func(t *testing.T) {
		if _, err := resolveSharedAkPath(teamRoot, boundFile, "file", "/projects/"+name+"/docs/readme.md"); !errors.Is(err, errs.ErrTempAccessPathNotAllowed) {
			t.Errorf("want ErrTempAccessPathNotAllowed, got %v", err)
		}
	})

	t.Run("URL 项目名与凭证绑定不一致", func(t *testing.T) {
		if _, err := resolveSharedAkPath(teamRoot, boundFile, "file", "/projects/其他项目/北京未来2天天气预报.docx"); !errors.Is(err, errs.ErrTempAccessPathNotAllowed) {
			t.Errorf("want ErrTempAccessPathNotAllowed, got %v", err)
		}
	})

	t.Run("URL 携带路径穿越", func(t *testing.T) {
		if _, err := resolveSharedAkPath(teamRoot, boundFile, "file", "/projects/"+name+"/../其他项目/secret.docx"); !errors.Is(err, errs.ErrTempAccessPathNotAllowed) {
			t.Errorf("want ErrTempAccessPathNotAllowed, got %v", err)
		}
	})

	t.Run("dir 类型绑定子目录内文件", func(t *testing.T) {
		got, err := resolveSharedAkPath(teamRoot, "/projects/"+name+"/docs", "dir", "/projects/"+name+"/docs/readme.md")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join(projectDir, "docs", "readme.md"); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("dir 类型目录本身不可下载", func(t *testing.T) {
		if _, err := resolveSharedAkPath(teamRoot, "/projects/"+name+"/docs", "dir", "/projects/"+name+"/docs"); !errors.Is(err, errs.ErrTempAccessPathNotAllowed) {
			t.Errorf("want ErrTempAccessPathNotAllowed, got %v", err)
		}
	})

	t.Run("dir 类型绑定目录外文件", func(t *testing.T) {
		if _, err := resolveSharedAkPath(teamRoot, "/projects/"+name+"/docs", "dir", boundFile); !errors.Is(err, errs.ErrTempAccessPathNotAllowed) {
			t.Errorf("want ErrTempAccessPathNotAllowed, got %v", err)
		}
	})

	t.Run("dir 类型绑定项目根目录", func(t *testing.T) {
		got, err := resolveSharedAkPath(teamRoot, "/projects/"+name, "dir", boundFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join(projectDir, "北京未来2天天气预报.docx"); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if _, err := resolveSharedAkPath(teamRoot, "/projects/"+name, "dir", "/projects/"+name); !errors.Is(err, errs.ErrTempAccessPathNotAllowed) {
			t.Errorf("project root itself should be rejected, got %v", err)
		}
		if _, err := resolveSharedAkPath(teamRoot, "/projects/"+name, "dir", "/projects/其他项目/file.docx"); !errors.Is(err, errs.ErrTempAccessPathNotAllowed) {
			t.Errorf("other project should be rejected, got %v", err)
		}
	})
}

func TestMoveFileCrossDeviceFallsBackToCopy(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := srcDir + "/hello.md"
	dstPath := dstDir + "/hello.md"
	contents := []byte("hello from upload")
	if err := os.WriteFile(srcPath, contents, 0644); err != nil {
		t.Fatal(err)
	}

	rename := func(_, _ string) error {
		return &os.LinkError{Op: "rename", Err: syscall.EXDEV}
	}
	if err := moveFileCrossDevice(srcPath, dstPath, rename); err != nil {
		t.Fatalf("moveFileCrossDevice returned error: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(contents) {
		t.Errorf("destination contents = %q, want %q", got, contents)
	}
	if _, err := os.Stat(srcPath); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("source still exists or stat failed: %v", err)
	}
}
