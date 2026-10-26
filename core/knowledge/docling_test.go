package knowledge

import (
	"context"
	"reflect"
	"testing"
)

func TestReadArgs(t *testing.T) {
	tests := []struct {
		name   string
		script string
		src    string
		opts   ReadOptions
		want   []string
	}{
		{
			name:   "缺省值使用 markdown 和默认 max-chars",
			script: "/scripts/docling/read.py",
			src:    "/tmp/a.pdf",
			opts:   ReadOptions{},
			want:   []string{"/scripts/docling/read.py", "--input", "/tmp/a.pdf", "--format", "markdown", "--max-chars", "50000"},
		},
		{
			name:   "非法 format 回退 markdown",
			script: "/s/read.py",
			src:    "/tmp/a.docx",
			opts:   ReadOptions{Format: "html", MaxChars: 30000},
			want:   []string{"/s/read.py", "--input", "/tmp/a.docx", "--format", "markdown", "--max-chars", "30000"},
		},
		{
			name:   "text 格式附带 pages",
			script: "/s/read.py",
			src:    "/tmp/a.pdf",
			opts:   ReadOptions{Format: "text", MaxChars: 30000, Pages: "1,3,5-8"},
			want:   []string{"/s/read.py", "--input", "/tmp/a.pdf", "--format", "text", "--max-chars", "30000", "--pages", "1,3,5-8"},
		},
		{
			name:   "负数 max-chars 回退默认值",
			script: "/s/read.py",
			src:    "/tmp/a.pdf",
			opts:   ReadOptions{MaxChars: -1},
			want:   []string{"/s/read.py", "--input", "/tmp/a.pdf", "--format", "markdown", "--max-chars", "50000"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readArgs(tt.script, tt.src, tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("readArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertArgs(t *testing.T) {
	got := convertArgs("/s/convert.py", "/tmp/a.pdf", "/tmp/out/a.md")
	want := []string{"/s/convert.py", "--input", "/tmp/a.pdf", "--output", "/tmp/out/a.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("convertArgs() = %v, want %v", got, want)
	}
}

func TestReadFile(t *testing.T) {
	t.Log(ReadFile(context.Background(), "/Users/ysmini/Downloads/杨树-Golang.pdf", ReadOptions{
		Format:   "markdown",
		MaxChars: 50000,
	}))
}
