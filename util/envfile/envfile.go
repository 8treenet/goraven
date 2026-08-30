// Package envfile 解析 dotenv 风格的环境变量文件（.profile / .env）。
//
// 支持的格式：
//   - KEY=VALUE
//   - KEY="quoted value"  / KEY='quoted value'
//   - 以 # 开头的整行注释
//   - 空行
//   - 行尾若无引号，# 之后视为注释
//
// 不做 shell 展开（如 $VAR、$(...)、`...`），保持纯字面量。
package envfile

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// Parse 将 dotenv 风格内容解析为 "KEY=VALUE" 形式的切片，
// 顺序与文件中出现顺序一致；重复键由调用方决定保留策略
// （exec.Cmd 默认取最后一个）。
//
// 解析失败的行返回错误并附带行号，便于排查。
func Parse(data []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	// 允许较长的单行
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	result := make([]string, 0)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("envfile: line %d: %w", lineNum, err)
		}
		if key == "" {
			continue
		}
		result = append(result, key+"="+value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("envfile: scan failed: %w", err)
	}
	return result, nil
}

// parseLine 解析单行 KEY=VALUE，返回 key、value。允许 `export KEY=VALUE`。
func parseLine(line string) (string, string, error) {
	// 兼容 shell 习惯写法 `export FOO=bar`
	if rest, ok := strings.CutPrefix(line, "export "); ok {
		line = strings.TrimSpace(rest)
	}

	idx := strings.IndexByte(line, '=')
	if idx <= 0 {
		return "", "", fmt.Errorf("invalid format, expected KEY=VALUE: %q", line)
	}

	key := strings.TrimSpace(line[:idx])
	if !isValidKey(key) {
		return "", "", fmt.Errorf("invalid key: %q", key)
	}

	raw := strings.TrimSpace(line[idx+1:])
	value, err := unquoteValue(raw)
	if err != nil {
		return "", "", err
	}
	return key, value, nil
}

// unquoteValue 处理可选的引号包裹，并去除非引号情形下的行尾注释。
func unquoteValue(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	switch raw[0] {
	case '"':
		end := strings.LastIndexByte(raw, '"')
		if end == 0 {
			return "", fmt.Errorf("unterminated double quote")
		}
		return raw[1:end], nil
	case '\'':
		end := strings.LastIndexByte(raw, '\'')
		if end == 0 {
			return "", fmt.Errorf("unterminated single quote")
		}
		return raw[1:end], nil
	}

	// 非引号情形：行尾 # 视为注释
	if idx := strings.IndexByte(raw, '#'); idx >= 0 {
		raw = strings.TrimSpace(raw[:idx])
	}
	return raw, nil
}

// Serialize 将 KEY=VALUE 列表序列化为 dotenv 风格文本，每项一行，以 \n 结尾。
// 值中含空格 / # 时会用双引号包裹；不附加任何注释（CRUD 写回会丢失原文件中的注释）。
// 值含换行或双引号时返回错误——Parse 当前不解析转义序列，写出去无法被反向解析。
// key 不合法或不存在 '=' 时同样返回错误。
func Serialize(entries []string) ([]byte, error) {
	var buf bytes.Buffer
	for _, e := range entries {
		idx := strings.IndexByte(e, '=')
		if idx <= 0 {
			return nil, fmt.Errorf("envfile: invalid entry, expected KEY=VALUE: %q", e)
		}
		key := e[:idx]
		if !isValidKey(key) {
			return nil, fmt.Errorf("envfile: invalid key: %q", key)
		}
		value := e[idx+1:]
		if strings.ContainsAny(value, "\n\r\"") {
			return nil, fmt.Errorf("envfile: value of %s contains newline or double-quote which is not supported", key)
		}
		buf.WriteString(key)
		buf.WriteByte('=')
		buf.WriteString(quoteValueIfNeeded(value))
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

// quoteValueIfNeeded 在值含空格 / # / 单引号时用双引号包裹，保证再次 Parse 时能完整还原。
// 由 Serialize 在调用前确保不含双引号和换行。
func quoteValueIfNeeded(v string) string {
	if v == "" {
		return `""`
	}
	if strings.ContainsAny(v, " \t#'") {
		return `"` + v + `"`
	}
	return v
}

// isValidKey 判定环境变量名是否合法：[A-Za-z_][A-Za-z0-9_]*。
func isValidKey(key string) bool {
	if key == "" {
		return false
	}
	for i, r := range key {
		switch {
		case r == '_':
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}
