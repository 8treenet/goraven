package util

import (
	"errors"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var skillNamePattern = regexp.MustCompile(`^[a-zA-Z@][a-zA-Z0-9@_-]{1,63}$`)

var (
	ErrInvalidSkillFormat = errors.New("invalid skill format: missing frontmatter")
	ErrMissingSkillName   = errors.New("skill name is required")
	ErrInvalidSkillName   = errors.New("skill name must start with a letter or @ and contain only letters, digits, @, hyphens, and underscores")
)

type SkillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Slug        string `yaml:"slug"`
}

// name: xxx
// description: xxx

// # 内容...

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

	fm.Name = normalizeSkillName(fm.Name)
	fm.Slug = normalizeSkillName(fm.Slug)

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

	content = strings.TrimPrefix(content, "\n")

	return frontmatter, content, nil
}

var skillNameReplacer = strings.NewReplacer(
	"/", "-",
	"\\", "-",
	":", "-",
	"*", "-",
	"?", "-",
	"\"", "-",
	"<", "-",
	">", "-",
	"|", "-",
)

func normalizeSkillName(s string) string {
	return skillNameReplacer.Replace(s)
}

var frontmatterNameRe = regexp.MustCompile(`(?m)^(name:)\s*.+$`)

func RewriteSkillName(content, newName string) (string, error) {
	frontmatter, body, err := parseFrontmatter(content)
	if err != nil {
		return "", err
	}
	quoted := "\"" + newName + "\""
	if frontmatterNameRe.MatchString(frontmatter) {
		frontmatter = frontmatterNameRe.ReplaceAllString(frontmatter, "${1} "+quoted)
	} else {
		frontmatter = "name: " + quoted + "\n" + frontmatter
	}
	return "---\n" + frontmatter + "\n---\n" + body, nil
}
