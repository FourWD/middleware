package common

import (
	"crypto/rand"
	"encoding/base64"
)

func RandomString(length int) string {
	if length <= 0 {
		length = 16
	}
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		LogError("RANDOM_GENERATE_ERROR", map[string]interface{}{"error": err.Error()}, "")
	}
	// Encode the random bytes as a base64 string.
	randomString := base64.RawURLEncoding.EncodeToString(randomBytes)
	// Trim the string to the desired length if necessary.
	if len(randomString) > length {
		randomString = randomString[:length]
	}
	return randomString
}
