package util

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("AdminPass123!")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if len(hash) < 4 || hash[:2] != "$2" {
		t.Fatalf("expected bcrypt hash, got %q", hash)
	}
	if hash == "AdminPass123!" {
		t.Fatal("password stored in plaintext")
	}

	matched, upgrade := VerifyPassword(hash, "AdminPass123!")
	if !matched || upgrade {
		t.Fatalf("bcrypt verify: matched=%v upgrade=%v", matched, upgrade)
	}
	matched, _ = VerifyPassword(hash, "wrong-password")
	if matched {
		t.Fatal("wrong password accepted")
	}

	// 同样的口令两次哈希必须不同（每用户随机盐）
	hash2, err := HashPassword("AdminPass123!")
	if err != nil {
		t.Fatalf("HashPassword(2): %v", err)
	}
	if hash == hash2 {
		t.Fatal("identical password produced identical hash (missing salt)")
	}
}

func TestVerifyLegacyMD5PasswordTriggersUpgrade(t *testing.T) {
	legacy := MD5("legacy-pass")
	matched, upgrade := VerifyPassword(legacy, "legacy-pass")
	if !matched || !upgrade {
		t.Fatalf("legacy MD5 verify: matched=%v upgrade=%v", matched, upgrade)
	}
	matched, _ = VerifyPassword(legacy, "other-pass")
	if matched {
		t.Fatal("legacy MD5 accepted wrong password")
	}
}

func TestVerifyPasswordRejectsUnknownFormat(t *testing.T) {
	matched, upgrade := VerifyPassword("not-a-recognized-hash", "not-a-recognized-hash")
	if matched || upgrade {
		t.Fatalf("unknown hash format must not authenticate: matched=%v upgrade=%v", matched, upgrade)
	}
	if _, err := HashPassword(""); err == nil {
		t.Fatal("empty password must be rejected")
	}
}
