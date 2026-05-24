package infra

import (
	"strings"
	"testing"
)

func TestHashBlacklistToken_Deterministic(t *testing.T) {
	a := hashBlacklistToken("eyJhbGciOi...")
	b := hashBlacklistToken("eyJhbGciOi...")
	if a != b {
		t.Fatalf("hash should be deterministic for the same input: %q vs %q", a, b)
	}
}

func TestHashBlacklistToken_DifferentTokensProduceDifferentHashes(t *testing.T) {
	a := hashBlacklistToken("token-a")
	b := hashBlacklistToken("token-b")
	if a == b {
		t.Fatal("different tokens must produce different hashes")
	}
}

func TestHashBlacklistToken_EmptyTokenStillHashes(t *testing.T) {
	h := hashBlacklistToken("")
	// SHA-256 of empty string is a known constant
	const wantEmptyHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if h != wantEmptyHash {
		t.Fatalf("empty hash mismatch: got %q want %q", h, wantEmptyHash)
	}
}

func TestHashBlacklistToken_HexLength(t *testing.T) {
	h := hashBlacklistToken("any-token")
	// SHA-256 -> 32 bytes -> 64 hex chars
	if len(h) != 64 {
		t.Fatalf("hash length: got %d want 64 (hex SHA-256)", len(h))
	}
	if strings.ContainsFunc(h, func(r rune) bool {
		return !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'))
	}) {
		t.Fatalf("hash contains non-hex chars: %q", h)
	}
}

func TestHashBlacklistToken_DoesNotContainOriginalToken(t *testing.T) {
	const secret = "super-secret-jwt-payload"
	if h := hashBlacklistToken(secret); strings.Contains(h, secret) {
		t.Fatal("hash must not leak the original token as a substring")
	}
}

func TestRedisBlacklistKey_UsesHashNotRaw(t *testing.T) {
	const token = "secret-token"
	key := redisBlacklistKey(token)

	if strings.Contains(key, token) {
		t.Fatal("Redis key must not contain raw token (PII leak via MONITOR/RDB)")
	}
	if !strings.HasPrefix(key, "blacklist:") {
		t.Fatalf("expected blacklist: prefix, got %q", key)
	}
	if !strings.Contains(key, hashBlacklistToken(token)) {
		t.Fatal("Redis key should contain the token hash")
	}
}
