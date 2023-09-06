package common

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Database *gorm.DB
var DatabaseMysql *sql.DB

type DSN struct {
	Username string
	Password string
	Database string
	IP       string
	Instance string
}

func CreateDSN(isGCP bool, dsn DSN) string {
	var protocol string
	setting := "?charset=utf8mb4&parseTime=True&loc=Asia%%2FBangkok"
	if isGCP {
		protocol = fmt.Sprintf("unix(/cloudsql/%s)", dsn.Instance)
	} else {
		protocol = fmt.Sprintf("tcp(%s:3306)", dsn.IP)
		setting += "&loc=Local"
	}
	return fmt.Sprintf("%s:%s@%s/%s%s", dsn.Username, dsn.Password, protocol, dsn.Database, setting)
}

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

	timeZone := "Asia/Bangkok"
	Database.Raw("SET time_zone=?", timeZone)
	DatabaseMysql.Exec("SET time_zone=?", timeZone)

	return nil
}
