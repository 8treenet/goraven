package util

import (
	"errors"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// skillNamePattern 技能名称合法规则：字母开头，仅允许字母、数字、连字符、下划线、冒号
var skillNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9:_-]{1,127}$`)

var (
	ErrInvalidSkillFormat = errors.New("invalid skill format: missing frontmatter")
	ErrMissingSkillName   = errors.New("skill name is required")
	ErrInvalidSkillName   = errors.New("skill name must start with a lowercase letter and contain only lowercase letters, digits, hyphens, underscores, and colons")
)

// SkillFrontmatter 技能 frontmatter 结构，与 eino 的 FrontMatter 保持一致
type SkillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Slug        string `yaml:"slug"`
}

// ParseSkillFrontmatter 从 skill content 中解析 frontmatter
// 格式：
// ---
// name: xxx
// description: xxx
// ---
// # 内容...
// 当 name 不合法时，会尝试使用 slug 作为 fallback
func ParseSkillFrontmatter(content string) (name, description string, err error) {
	content = strings.TrimSpace(content)

	frontmatter, _, err := parseFrontmatter(content)
	if err != nil {
		return "", "", err
	}

	var fm SkillFrontmatter
	if err = yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return "", "", err
	}

	if fm.Name == "" && fm.Slug == "" {
		return "", "", ErrMissingSkillName
	}

	if !skillNamePattern.MatchString(fm.Name) {
		if fm.Slug == "" || !skillNamePattern.MatchString(fm.Slug) {
			return "", "", ErrInvalidSkillName
		}
		return fm.Slug, fm.Description, nil
	}

	return fm.Name, fm.Description, nil
}

// parseFrontmatter 分离 frontmatter 和 content，与 eino 的实现保持一致
func parseFrontmatter(data string) (frontmatter, content string, err error) {
	const delimiter = "---"

	if !strings.HasPrefix(data, delimiter) {
		return "", "", ErrInvalidSkillFormat
	}

	rest := data[len(delimiter):]
	endIdx := strings.Index(rest, "\n"+delimiter)
	if endIdx == -1 {
		return "", "", ErrInvalidSkillFormat
	}

	frontmatter = strings.TrimSpace(rest[:endIdx])
	content = rest[endIdx+len("\n"+delimiter):]

	if strings.HasPrefix(content, "\n") {
		content = content[1:]
	}

	return frontmatter, content, nil
}
