package envfile

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

func Parse(data []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

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

func parseLine(line string) (string, string, error) {

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

	if idx := strings.IndexByte(raw, '#'); idx >= 0 {
		raw = strings.TrimSpace(raw[:idx])
	}
	return raw, nil
}

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

func quoteValueIfNeeded(v string) string {
	if v == "" {
		return `""`
	}
	if strings.ContainsAny(v, " \t#'") {
		return `"` + v + `"`
	}
	return v
}

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
