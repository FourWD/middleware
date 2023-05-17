package common

import (
	"database/sql"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Database *gorm.DB
var DatabaseMysql *sql.DB

func ConnectDatabase(dsn string) error {
	var err error

	Database, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	if err != nil {
		PrintError(`Gorm`, `Connection Error !`)
		panic(err)
	}

	DatabaseMysql, err = sql.Open("mysql", dsn)
	if err != nil {
		PrintError(`DB Mysql`, `Connection Error !`)
		panic(err)
	}

	return nil
}
