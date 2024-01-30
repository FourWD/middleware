package common

import (
	"encoding/base64"
	"fmt"
	"math/rand"
)

// var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// // n is the length of random string we want to generate
// func RandStr(n int) string {
// 	b := make([]byte, n)
// 	for i := range b {
// 		// randomly select 1 character from given charset
// 		b[i] = charset[rand.Intn(len(charset))]
// 	}
// 	return string(b)
// }

func RandomString(length int) string {
	if length <= 0 {
		length = 16
	}
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Printf("Failed to generate random bytes: %v\n", err)
	}
	// Encode the random bytes as a base64 string.
	randomString := base64.RawURLEncoding.EncodeToString(randomBytes)
	return randomString
}
