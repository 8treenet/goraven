package util

import (
	"strings"

	"github.com/google/uuid"
)

// UUID 生成一个 UUID 字符串
func UUID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
