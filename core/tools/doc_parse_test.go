package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocParseResolveSourcePath(t *testing.T) {
	workspace := t.TempDir()
	extra := t.TempDir()

	wsFile := filepath.Join(workspace, "temp", "a.pdf")
	if err := os.MkdirAll(filepath.Dir(wsFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(wsFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	extraFile := filepath.Join(extra, "proj", "report.pdf")
	if err := os.MkdirAll(filepath.Dir(extraFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(extraFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	d := &DocParse{workspace: workspace, extraWorkspace: extra}

	t.Run("工作空间内绝对路径", func(t *testing.T) {
		got, err := d.resolveSourcePath(wsFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != wsFile {
			t.Fatalf("got %q, want %q", got, wsFile)
		}
	})

	t.Run("团队项目目录内绝对路径", func(t *testing.T) {
		got, err := d.resolveSourcePath(extraFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != extraFile {
			t.Fatalf("got %q, want %q", got, extraFile)
		}
	})
}

// requireInside 断言 path 位于 root 之内（防止输出写出工作空间）。
func requireInside(t *testing.T, root, path string) {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatalf("path %q is not relative to %q: %v", path, root, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		t.Fatalf("path escaped the workspace: %q is outside %q", path, root)
	}
}

func TestDocParseResolveOutputPath(t *testing.T) {
	workspace := t.TempDir()
	extra := t.TempDir()

	d := &DocParse{workspace: workspace, extraWorkspace: extra}

	t.Run("工作空间内绝对路径原样采用", func(t *testing.T) {
		want := filepath.Join(workspace, "temp", "b.md")
		got, err := d.resolveOutputPath(want)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("团队项目目录内绝对路径原样采用", func(t *testing.T) {
		want := filepath.Join(extra, "proj", "b.md")
		got, err := d.resolveOutputPath(want)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("沙箱内绝对路径拼接到工作空间", func(t *testing.T) {
		got, err := d.resolveOutputPath("/temp/b.md")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(workspace, "temp", "b.md")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("宿主机绝对路径被收敛到工作空间内", func(t *testing.T) {
		got, err := d.resolveOutputPath("/etc/cron.d/malicious.md")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		requireInside(t, workspace, got)
	})
}

func TestDocParseSaveOutputStaysInWorkspace(t *testing.T) {
	workspace := t.TempDir()

	d := &DocParse{workspace: workspace}
	src := filepath.Join(workspace, "temp", "src.md")
	if err := os.MkdirAll(filepath.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("converted"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("写入工作空间内目标路径", func(t *testing.T) {
		dst := filepath.Join(workspace, "out", "b.md")
		if err := d.saveOutput(src, dst); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("output not written: %v", err)
		}
		if string(got) != "converted" {
			t.Fatalf("content = %q, want %q", string(got), "converted")
		}
	})

	t.Run("沙箱外目标路径不会落到根目录之外", func(t *testing.T) {
		outside := filepath.Join(t.TempDir(), "cron.d", "malicious.md")
		// 要么返回错误，要么被收敛到工作空间内；两种结果都不允许在根目录之外落盘
		_ = d.saveOutput(src, outside)
		if _, statErr := os.Stat(outside); statErr == nil {
			t.Fatalf("sandbox escape: %s was created outside the workspace", outside)
		}
	})
}

func TestNormalizeDocParseRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     DocParseRequest
		wantErr bool
		check   func(t *testing.T, got *DocParseRequest)
	}{
		{
			name: "read 缺省值规范化",
			req:  DocParseRequest{Mode: "read", FilePath: "/temp/a.pdf"},
			check: func(t *testing.T, got *DocParseRequest) {
				if got.Format != "markdown" {
					t.Fatalf("Format = %q, want markdown", got.Format)
				}
				if got.MaxChars != 50000 {
					t.Fatalf("MaxChars = %d, want 50000", got.MaxChars)
				}
			},
		},
		{
			name: "read 显式 text 格式保留",
			req:  DocParseRequest{Mode: "read", FilePath: "/temp/a.docx", Format: "text", MaxChars: 30000},
			check: func(t *testing.T, got *DocParseRequest) {
				if got.Format != "text" || got.MaxChars != 30000 {
					t.Fatalf("Format = %q, MaxChars = %d, want text/30000", got.Format, got.MaxChars)
				}
			},
		},
		{
			name:    "mode 非法",
			req:     DocParseRequest{Mode: "chunk", FilePath: "/temp/a.pdf"},
			wantErr: true,
		},
		{
			name:    "mode 为空",
			req:     DocParseRequest{FilePath: "/temp/a.pdf"},
			wantErr: true,
		},
		{
			name:    "file_path 为空",
			req:     DocParseRequest{Mode: "read"},
			wantErr: true,
		},
		{
			name:    "convert 缺 output_path",
			req:     DocParseRequest{Mode: "convert", FilePath: "/temp/a.pdf"},
			wantErr: true,
		},
		{
			name:    "convert output_path 非 .md",
			req:     DocParseRequest{Mode: "convert", FilePath: "/temp/a.pdf", OutputPath: "/temp/a.txt"},
			wantErr: true,
		},
		{
			name:    "convert output_path 与源相同",
			req:     DocParseRequest{Mode: "convert", FilePath: "/temp/a.md", OutputPath: "/temp/a.md"},
			wantErr: true,
		},
		{
			name: "convert 合法",
			req:  DocParseRequest{Mode: "convert", FilePath: "/temp/a.pdf", OutputPath: "/temp/b.md"},
			check: func(t *testing.T, got *DocParseRequest) {
				if got.OutputPath != "/temp/b.md" {
					t.Fatalf("OutputPath = %q, want /temp/b.md", got.OutputPath)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeDocParseRequest(&tt.req)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
