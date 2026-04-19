package common

import (
	"database/sql"

	"github.com/FourWD/middleware/infra"
	"gorm.io/gorm"
)

var Database *gorm.DB
var DatabaseSql *sql.DB

func ConnectDatabaseViper(maxOpenConns int, maxIdleConns int) error {
	dns := DNS{
		Username: infra.GetEnv("DB_USER", ""),
		Password: infra.GetEnv("DB_PASSWORD", ""),
		Database: infra.GetEnv("DB_NAME", ""),
		IP:       infra.GetEnv("DB_HOST", ""),
		Instance: infra.GetEnv("DB_INSTANCE", ""),
	}

	if infra.AppInfo.Env == "local" {
		dns.Instance = ""
	}

	return connectDatabase(CreateMySqlDSN(dns), maxOpenConns, maxIdleConns)
}

func connectDatabase(dns string, maxOpenConns int, maxIdleConns int) error {
	Database, DatabaseSql = ConnectMySqlDatabase(dns, maxOpenConns, maxIdleConns)
	return nil
}

func ConnectDatabaseMySqlGoogle(DNS DNS) (*sql.DB, error) {
	if infra.AppInfo.Env == "local" {
		DNS.Instance = ""
	}

	dsn := CreateMySqlDSN(DNS)

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
