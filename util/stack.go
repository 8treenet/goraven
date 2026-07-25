package util

import (
	"bytes"
	"runtime/debug"
	"strings"
)

func TrimStack(n int) string {
	return TrimStackFrom(debug.Stack(), n)
}

func TrimStackFrom(stack []byte, n int) string {
	if len(stack) == 0 || n <= 0 {
		return ""
	}

	lines := bytes.Split(stack, []byte("\n"))
	if len(lines) == 0 {
		return string(stack)
	}

	var frames [][]byte
	for i := 1; i < len(lines); i++ {
		line := bytes.TrimSpace(lines[i])
		if len(line) == 0 {
			continue
		}

		lineStr := string(lines[i])
		if strings.Contains(lineStr, "runtime/debug.Stack") ||
			strings.Contains(lineStr, "util.TrimStack") ||
			strings.Contains(lineStr, "util/stack.go") {

			if i+1 < len(lines) {
				i++
			}
			continue
		}
		frames = append(frames, lines[i])
	}

	if len(frames) == 0 {
		return string(stack)
	}

	limit := n * 2
	if limit > len(frames) {
		limit = len(frames)
	}
	result := bytes.Join(frames[:limit], []byte("\n"))
	return strings.TrimSpace(string(result))
}
