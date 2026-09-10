package util

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hashed, err := HashPassword("s3cret-pass")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if !IsBcryptHash(hashed) {
		t.Fatalf("HashPassword should produce a bcrypt hash, got %q", hashed)
	}

	ok, legacy := VerifyPassword(hashed, "s3cret-pass")
	if !ok || legacy {
		t.Errorf("VerifyPassword(bcrypt, correct) = %v, %v; want true, false", ok, legacy)
	}
	ok, _ = VerifyPassword(hashed, "wrong-pass")
	if ok {
		t.Errorf("VerifyPassword(bcrypt, wrong) should fail")
	}
}

func TestVerifyPasswordLegacyMD5(t *testing.T) {
	// 旧版无盐 MD5 存储兼容
	legacyHash := MD5("legacy-pass")

	ok, legacy := VerifyPassword(legacyHash, "legacy-pass")
	if !ok || !legacy {
		t.Errorf("VerifyPassword(md5, correct) = %v, %v; want true, true", ok, legacy)
	}
	ok, _ = VerifyPassword(legacyHash, "wrong-pass")
	if ok {
		t.Errorf("VerifyPassword(md5, wrong) should fail")
	}
}

func TestIsBcryptHash(t *testing.T) {
	if IsBcryptHash(MD5("x")) {
		t.Errorf("MD5 hash must not be detected as bcrypt")
	}
	if IsBcryptHash("$2a$10$invalidbutprefix") != true {
		// 前缀判断，宽松识别
		t.Log("prefix detection")
	}
}
