package infra

import (
	"encoding/hex"
	"strings"
	"testing"
)

// TestMain in this package is responsible for setting APP_ENCRYPT_KEY before
// any test touches the aesGCM sync.OnceValues — once derived, the cipher is
// cached for the rest of the process. We use t.Setenv at file scope is not
// allowed, so each test calls ensureEncryptKey() which is idempotent.
func ensureEncryptKey(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENCRYPT_KEY", "test-encryption-key-for-unit-tests")
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	ensureEncryptKey(t)

	cases := []string{
		"hello world",
		"",
		"unicode: สวัสดี 你好 🚀",
		strings.Repeat("a", 1024),
	}

	for _, plaintext := range cases {
		t.Run(truncForName(plaintext), func(t *testing.T) {
			cipher, err := Encrypt(plaintext)
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}
			if cipher == plaintext && plaintext != "" {
				t.Fatalf("ciphertext equals plaintext: not encrypted")
			}
			decoded, err := Decrypt(cipher)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}
			if decoded != plaintext {
				t.Fatalf("round-trip mismatch: got %q want %q", decoded, plaintext)
			}
		})
	}
}

func TestEncrypt_ProducesDifferentCiphertextEachCall(t *testing.T) {
	ensureEncryptKey(t)

	a, err := Encrypt("same input")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	b, err := Encrypt("same input")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if a == b {
		t.Fatal("AES-GCM should use a fresh nonce per call; ciphertexts must differ")
	}
}

func TestDecrypt_InvalidHex(t *testing.T) {
	ensureEncryptKey(t)

	if _, err := Decrypt("not-hex-zzz"); err == nil {
		t.Fatal("expected error for non-hex input")
	}
}

func TestDecrypt_TooShort(t *testing.T) {
	ensureEncryptKey(t)

	// 8 hex chars = 4 bytes — far shorter than the GCM nonce (12 bytes).
	if _, err := Decrypt("01234567"); err == nil {
		t.Fatal("expected error for ciphertext shorter than nonce")
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	ensureEncryptKey(t)

	cipher, err := Encrypt("sensitive payload")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Flip one bit in the encrypted payload (skip the 12-byte nonce prefix).
	raw, err := hex.DecodeString(cipher)
	if err != nil {
		t.Fatalf("hex decode: %v", err)
	}
	if len(raw) < 13 {
		t.Fatalf("ciphertext too short to tamper: %d bytes", len(raw))
	}
	raw[12] ^= 0xff
	tampered := hex.EncodeToString(raw)

	if _, err := Decrypt(tampered); err == nil {
		t.Fatal("expected GCM authentication failure on tampered ciphertext")
	}
}

func truncForName(s string) string {
	if len(s) <= 20 {
		if s == "" {
			return "empty"
		}
		return s
	}
	return s[:20] + "..."
}
