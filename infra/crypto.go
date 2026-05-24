package infra

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
)

// aesGCM derives the AES-GCM cipher from APP_ENCRYPT_KEY once on first use.
// Reusing the AEAD avoids re-hashing the key and re-creating the cipher
// block on every Encrypt/Decrypt call.
var aesGCM = sync.OnceValues(func() (cipher.AEAD, error) {
	key := sha256.Sum256([]byte(GetEnv("APP_ENCRYPT_KEY", "")))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
})

// Encrypt AES-GCM-encrypts plaintext with APP_ENCRYPT_KEY and returns
// hex-encoded (nonce || ciphertext).
func Encrypt(plaintext string) (string, error) {
	gcm, err := aesGCM()
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(cipherText), nil
}

// Decrypt reverses Encrypt: hex-decode (nonce || ciphertext) and AES-GCM-open.
func Decrypt(cipherText string) (string, error) {
	gcm, err := aesGCM()
	if err != nil {
		return "", err
	}

	cipherTextBytes, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherTextBytes) < nonceSize {
		return "", fmt.Errorf("cipherText too short")
	}
	nonce, payload := cipherTextBytes[:nonceSize], cipherTextBytes[nonceSize:]

	plainText, err := gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
