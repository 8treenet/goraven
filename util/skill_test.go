package util

import (
	"testing"
)

func TestParseSkillFrontmatter(t *testing.T) {
	content := `---
name: goraven-git-commit
description: 根据代码变更生成规范的 Git commit message。
---

# Git Commit Message 生成器

根据当前的代码变更，生成符合规范的 commit message。`

	name, description, err := ParseSkillFrontmatter(content)
	if err != nil {
		t.Fatalf("ParseSkillFrontmatter failed: %v", err)
	}
	t.Log(name, ":", description)

	if name != "goraven-git-commit" {
		t.Errorf("expected name 'goraven-git-commit', got '%s'", name)
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

func TestParseSkillFrontmatter_MissingGoRavenPrefix(t *testing.T) {
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

func TestParseSkillFrontmatter_NormalizeSlash(t *testing.T) {
	content := `---
name: "@scope/skill-name"
description: test
---

# Content`
	name, _, err := ParseSkillFrontmatter(content)
	if err != nil {
		t.Fatalf("ParseSkillFrontmatter failed: %v", err)
	}
	if name != "@scope-skill-name" {
		t.Errorf("expected '@scope-skill-name', got '%s'", name)
	}
}

func TestRewriteSkillName(t *testing.T) {
	content := `---
name: "@scope/skill-name"
description: test
---

# Body content`
	rewritten, err := RewriteSkillName(content, "fliggy-travel")
	if err != nil {
		t.Fatalf("RewriteSkillName failed: %v", err)
	}

	name, _, err := ParseSkillFrontmatter(rewritten)
	if err != nil {
		t.Fatalf("ParseSkillFrontmatter on rewritten content failed: %v", err)
	}
	if name != "fliggy-travel" {
		t.Errorf("expected rewritten name 'fliggy-travel', got '%s'", name)
	}
}

func TestRewriteSkillName_NoNameField(t *testing.T) {
	content := `---
slug: some-slug
description: test
---

# Body`
	rewritten, err := RewriteSkillName(content, "normalized-name")
	if err != nil {
		t.Fatalf("RewriteSkillName failed: %v", err)
	}

	name, _, err := ParseSkillFrontmatter(rewritten)
	if err != nil {
		t.Fatalf("ParseSkillFrontmatter on rewritten content failed: %v", err)
	}
	if name != "normalized-name" {
		t.Errorf("expected 'normalized-name', got '%s'", name)
	}
}
