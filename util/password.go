package util

import (
	"crypto/subtle"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// 口令哈希参数。bcrypt 自带每用户随机盐与可调代价因子，避免无盐 MD5 的
// 彩虹表 / 相同口令可关联问题。
const passwordHashCost = 12

// HashPassword 生成新账号 / 改密时应写入数据库的口令哈希。
// 返回值带有 bcrypt 前缀（$2a$ / $2b$），VerifyPassword 据此区分新旧格式。
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("password must not be empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), passwordHashCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword 校验明文口令。
//
// 返回 (matched, needUpgrade)：
//   - matched=false 一律视为口令错误；
//   - needUpgrade=true 表示库中仍是历史遗留的无盐 MD5 哈希，调用方在登录成功后
//     应立刻用 HashPassword 重写该字段，实现平滑迁移（无需停服、无需用户改密）。
//
// 未识别的哈希格式不会放行，避免以后出现「未知格式直接当明文比对」的降级风险。
func VerifyPassword(stored, plain string) (bool, bool) {
	if stored == "" || plain == "" {
		return false, false
	}

	if len(stored) >= 2 && stored[:2] == "$2" {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil, false
	}

	// 旧格式：32 位小写十六进制 MD5。使用常量时间比较，避免时序侧信道。
	legacy := MD5(plain)
	if len(stored) != len(legacy) {
		return false, false
	}
	return subtle.ConstantTimeCompare([]byte(stored), []byte(legacy)) == 1, true
}
