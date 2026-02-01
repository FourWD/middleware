package common

import (
	"database/sql"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var Database *gorm.DB
var DatabaseSql *sql.DB

func ConnectDatabaseViper(maxOpenConns int, maxIdleConns int) error {
	dns := DNS{
		Username: viper.GetString("database.username"),
		Password: viper.GetString("database.password"),
		Database: viper.GetString("database.database"),
		IP:       viper.GetString("database.ip"),
		Instance: viper.GetString("database.instance"),
	}

	isGCP := true
	if App.Env == "local" {
		isGCP = false
	}

	return connectDatabase(CreateMySqlDSN(isGCP, dns), maxOpenConns, maxIdleConns)
}

func connectDatabase(dns string, maxOpenConns int, maxIdleConns int) error {
	Database, DatabaseSql = ConnectMySqlDatabase(dns, maxOpenConns, maxIdleConns)
	return nil
}

func ConnectDatabaseMySqlGoogle(DNS DNS) (*sql.DB, error) {
	isGCP := true
	if App.Env == "local" {
		isGCP = false
	}

	dsn := CreateMySqlDSN(isGCP, DNS)

	database, err := sql.Open("mysql", dsn+"&loc=Asia%2FBangkok")
	if err != nil {
		LogError("DB_MYSQL_GOOGLE_CONNECTION_ERROR", map[string]interface{}{"error": err.Error()}, "")
		return nil, err
	}

	timeZone := "Asia/Bangkok"
	database.Exec("SET time_zone=?", timeZone)

	Log("DB_MYSQL_CONNECTION_SUCCESS", map[string]interface{}{}, "")

	return database, nil
}
