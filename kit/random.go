package kit

import (
	"crypto/rand"
	"encoding/base64"
)

func RandomString(length int) string {
	if length <= 0 {
		length = 16
	}
	randomBytes := make([]byte, length)
	_, _ = rand.Read(randomBytes)
	randomString := base64.RawURLEncoding.EncodeToString(randomBytes)
	if len(randomString) > length {
		randomString = randomString[:length]
	}
	return randomString
}
