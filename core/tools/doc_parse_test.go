package tools

import (
	"os"
	"path/filepath"
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
