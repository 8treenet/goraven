package util

import (
	"bytes"
	"runtime/debug"
	"strings"
)

// TrimStack returns the top n frames of the stack trace (closest to panic),
// skipping internal frames like runtime/debug.Stack and TrimStack itself.
func TrimStack(n int) string {
	return TrimStackFrom(debug.Stack(), n)
}

// TrimStackFrom returns the top n frames of the stack trace (closest to panic),
// skipping internal frames.
func TrimStackFrom(stack []byte, n int) string {
	if len(stack) == 0 || n <= 0 {
		return ""
	}

	lines := bytes.Split(stack, []byte("\n"))
	if len(lines) == 0 {
		return string(stack)
	}

	// Each frame is typically 2 lines:
	//   path/to/package.Function()
	//   \tfile.go:123 +0xabc
	var frames [][]byte
	for i := 1; i < len(lines); i++ {
		line := bytes.TrimSpace(lines[i])
		if len(line) == 0 {
			continue
		}
		// Skip internal frames
		lineStr := string(lines[i])
		if strings.Contains(lineStr, "runtime/debug.Stack") ||
			strings.Contains(lineStr, "util.TrimStack") ||
			strings.Contains(lineStr, "util/stack.go") {
			// Skip the next line too (file:line)
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

	// Limit to n frames (each frame is 2 lines)
	limit := n * 2
	if limit > len(frames) {
		limit = len(frames)
	}
	result := bytes.Join(frames[:limit], []byte("\n"))
	return strings.TrimSpace(string(result))
}
