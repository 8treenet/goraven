package util

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 对字符串做 MD5 哈希，返回 32 位小写十六进制。
//
// 安全说明：新口令一律使用 util.HashPassword（bcrypt + 随机盐）入库，
// 禁止再用本函数存储口令。此函数仅保留给 VerifyPassword 校验历史遗留的
// 无盐 MD5 哈希，以及与非口令场景（缓存键、内容摘要）使用。
func MD5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
