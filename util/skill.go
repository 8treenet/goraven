package util

import (
	"errors"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// skillNamePattern 技能名称合法规则：字母或@开头，仅允许字母、数字、@、连字符、下划线（需可作为目录名）
// 长度上限 64，与数据库 varchar(64) 列保持一致，确保目录名与 DB skillname 不会因截断而不一致
var skillNamePattern = regexp.MustCompile(`^[a-zA-Z@][a-zA-Z0-9@_-]{1,63}$`)

var (
	ErrInvalidSkillFormat = errors.New("invalid skill format: missing frontmatter")
	ErrMissingSkillName   = errors.New("skill name is required")
	ErrInvalidSkillName   = errors.New("skill name must start with a letter or @ and contain only letters, digits, @, hyphens, and underscores")
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

	content = strings.TrimPrefix(content, "\n")

	return frontmatter, content, nil
}

// skillNameReplacer 将不可作为目录名的字符（Windows/Unix 保留字符）统一替换为 -
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

// RewriteSkillName 将 SKILL.md frontmatter 中的 name 字段重写为指定值（目录名）。
//
// 背景问题：
// Eino 的 skill filesystemBackend 以 SKILL.md frontmatter 的 name 字段作为技能唯一标识
// （List 返回 FrontMatter.Name，Get 用 skill.Name == name 匹配），而非使用技能所在的目录名。
// 但市面上的技能生态（ClawHub、OpenSkill 等）大量 frontmatter name 带有 / @ : 等字符
// （如 "openakita/skills@fliggy-travel"），这些字符不能作为目录名。
//
// 这导致一个冲突：目录名（如 fliggy-travel）和 frontmatter name
// （如 openakita/skills@fliggy-travel）不一致，运行时 IsSkillSelected 用 DB 里的目录名匹配，
// 而 Eino 传过来的是 frontmatter name，两者对不上，技能被过滤掉，LLM 看不到。
//
// 解决方案：导入/同步时以目录名为唯一真相源，重写 SKILL.md 的 name 字段与目录名对齐，
// 使 Eino 读到的 name == 目录名 == 数据库 skillname，三方一致。
//
// 如果 name 字段不存在，在 frontmatter 开头添加。
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
