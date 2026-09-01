package util

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 生成 bcrypt 哈希，用于密码入库。
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPassword 校验明文密码与存储哈希是否匹配。
//
// 兼容历史版本的无盐 MD5 存储：存储值不是 bcrypt 时按 MD5 比对。
// 返回 ok 表示密码是否正确，legacy 表示是否命中旧版 MD5（调用方可借此透明升级）。
func VerifyPassword(storedHash, password string) (ok bool, legacy bool) {
	if IsBcryptHash(storedHash) {
		err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
		return err == nil, false
	}
	return storedHash == MD5(password), true
}

// IsBcryptHash 判断存储值是否为 bcrypt 哈希。
func IsBcryptHash(stored string) bool {
	// bcrypt 哈希以 $2 开头：$2a$、$2b$、$2y$ 等
	return strings.HasPrefix(stored, "$2")
}
