package fs

import (
	"strings"
	"testing"
)

func TestValidateDirName(t *testing.T) {
	valid := []string{
		"abc",
		"a-b_c",
		"foo.bar",
		".hidden",
		"my project",
		"中文项目",
		"项目 2026",
		"a",
		strings.Repeat("a", 255),
	}
	for _, name := range valid {
		if err := ValidateDirName(name); err != nil {
			t.Errorf("ValidateDirName(%q) = %v, want nil", name, err)
		}
	}

	invalid := []string{
		"",
		".",
		"..",
		"a/b",
		"/lead",
		"trail/",
		"a\x00b",
		strings.Repeat("a", 256),
	}
	for _, name := range invalid {
		if err := ValidateDirName(name); err == nil {
			t.Errorf("ValidateDirName(%q) = nil, want error", name)
		}
	}
}