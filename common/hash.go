package common

import (
	"crypto/md5"
	"encoding/hex"
)

func HashPassword(password string) string {
	hashPassword := md5.New()
	hashPassword.Write([]byte(password))
	Print(" md5 pass : ", hex.EncodeToString(hashPassword.Sum(nil)))
	return hex.EncodeToString(hashPassword.Sum(nil))
}
