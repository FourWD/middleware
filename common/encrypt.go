package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/spf13/viper"
)

func Encrypt(plaintext string) (string, error) {
	key := viper.GetString("encrypt_key")
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

	// Create a nonce of the appropriate size
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the data
	cipherText := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(cipherText), nil
}

// func EncryptAES(text string) clause.Expr {
// 	key := "fourwd"
// 	return clause.Expr{SQL: "HEX(AES_ENCRYPT(?, ?))", Vars: []interface{}{text, key}}
// }

// func DecryptAESSql(text string) string {
// 	key := "fourwd"
// 	return fmt.Sprintf(" CAST(AES_DECRYPT(UNHEX(%s), %s) AS CHAR) ", text, key)
// }
