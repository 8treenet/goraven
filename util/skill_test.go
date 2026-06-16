package util

import (
	"testing"
)

func TestParseSkillFrontmatter(t *testing.T) {
	content := `---
name: raven-git-commit
description: 根据代码变更生成规范的 Git commit message。
---

# Git Commit Message 生成器

根据当前的代码变更，生成符合规范的 commit message。`

	name, description, err := ParseSkillFrontmatter(content)
	if err != nil {
		t.Fatalf("ParseSkillFrontmatter failed: %v", err)
	}
	t.Log(name, ":", description)

	if name != "raven-git-commit" {
		t.Errorf("expected name 'raven-git-commit', got '%s'", name)
	}

	if description != "根据代码变更生成规范的 Git commit message。" {
		t.Errorf("unexpected description: %s", description)
	}
}

func TestParseSkillFrontmatter_InvalidFormat(t *testing.T) {
	content := `# No frontmatter here`
	_, _, err := ParseSkillFrontmatter(content)
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestParseSkillFrontmatter_MissingRavenPrefix(t *testing.T) {
	content := `---
name: git-commit
description: test
---

# Content`
	_, _, err := ParseSkillFrontmatter(content)
	if err != nil {
		t.Errorf("expected ErrInvalidSkillName, got %v", err)
	}
}

func TestParseSkillFrontmatter_MissingName(t *testing.T) {
	content := `---
description: test
---

# Content`
	_, _, err := ParseSkillFrontmatter(content)
	if err != ErrMissingSkillName {
		t.Errorf("expected ErrMissingSkillName, got %v", err)
	}
}
