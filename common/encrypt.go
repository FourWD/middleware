package common

import "gorm.io/gorm/clause"

func EncryptAES(text string) clause.Expr {
	key := "fourwd"
	return clause.Expr{SQL: "HEX(AES_ENCRYPT(?, ?))", Vars: []interface{}{text, key}}
}
