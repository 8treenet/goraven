package util

import "regexp"

// usernamePattern 限定 username 字符集
// 首尾必须是字母或数字，中间允许字母/数字/_/-
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$`)

// IsValidUsername 校验 username 是否可作为文件系统路径段安全使用
// 规则：长度 8-16；首尾字母或数字；中间仅允许字母/数字/_/-；
//       拒绝路径分隔符、控制字符、空白符、点号、非 ASCII 字符
func IsValidUsername(s string) bool {
	if len(s) < 8 || len(s) > 16 {
		return false
	}
	return usernamePattern.MatchString(s)
}
