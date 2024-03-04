package common

import (
	"fmt"

	"gorm.io/gorm/clause"
)

func EncryptAES(text string) clause.Expr {
	key := "fourwd"
	return clause.Expr{SQL: "HEX(AES_ENCRYPT(?, ?))", Vars: []interface{}{text, key}}
}

func DecryptAESSql(text string) string {
	key := "fourwd"
	return fmt.Sprintf(" CAST(AES_DECRYPT(UNHEX(%s), %s) AS CHAR) ", text, key)
}