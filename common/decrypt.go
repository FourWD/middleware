package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/spf13/viper"
)

func Decrypt(cipherText string) (string, error) {
	key := viper.GetString("encrypt_key")
	cipherTextBytes, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	// Create a new AES cipher block from the key
	keyBytes := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return "", err
	}

	// Create a GCM (Galois/Counter Mode) cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Separate the nonce and the ciphertext
	nonceSize := gcm.NonceSize()
	if len(cipherTextBytes) < nonceSize {
		return "", fmt.Errorf("cipherText too short")
	}
	nonce, ciphertextBytes := cipherTextBytes[:nonceSize], cipherTextBytes[nonceSize:]

	// Decrypt the data
	plainText, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
