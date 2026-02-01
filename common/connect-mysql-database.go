package common

import (
	"database/sql"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectMySqlDatabase(dns string, maxOpenConns int, maxIdleConns int) (*gorm.DB, *sql.DB) {
	dataGorm, errGorm := gorm.Open(mysql.Open(dns+"&loc=Asia%2FBangkok"), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	if errGorm != nil {
		LogError("DB_GORM_CONNECTION_ERROR", map[string]interface{}{"error": errGorm.Error()}, "")
		panic(errGorm)
	}

	dbSql, err := sql.Open("mysql", dns+"&loc=Asia%2FBangkok")
	if err != nil {
		LogError("DB_MYSQL_CONNECTION_ERROR", map[string]interface{}{"error": err.Error()}, "")
		panic(err)
	}

	timeZone := "Asia/Bangkok"
	dbSql.Exec("SET time_zone=?", timeZone)
	dataGorm.Exec("SET time_zone=?", timeZone)

	Log("DB_CONNECTION_SUCCESS", map[string]interface{}{}, "")
	initDatabaseConnectionPool(maxOpenConns, maxIdleConns)

	return dataGorm, dbSql
}

// func ConnectCustomDatabase(dnsConfig DNS) (*sql.DB, *gorm.DB) {
// 	dns := CreateDSN(false, dnsConfig)

// 	databaseMiddlePriceGorm, errGorm := gorm.Open(mysql.Open(dns+"&loc=Asia%2FBangkok"), &gorm.Config{
// 		SkipDefaultTransaction: true,
// 		PrepareStmt:            true,
// 	})

// 	if errGorm != nil {
// 		panic(errGorm)
// 	}

// 	database, err := sql.Open("mysql", dns+"&loc=Asia%2FBangkok")
// 	if err != nil {
// 		panic(err)
// 	}

// 	timeZone := "Asia/Bangkok"
// 	database.Exec("SET time_zone=?", timeZone)

// 	return database, databaseMiddlePriceGorm
// }
