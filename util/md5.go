package util

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 对字符串做 MD5 哈希，返回 32 位小写十六进制
// 用于密码入库前的统一哈希（前端传明文，后端哈希）
func MD5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
