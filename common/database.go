package common

import (
	"database/sql"

	"github.com/FourWD/middleware/infra"
	"gorm.io/gorm"
)

var Database *gorm.DB
var DatabaseSql *sql.DB

func ConnectDatabaseViper(maxOpenConns int, maxIdleConns int) error {
	dns := infra.DNS{
		Username: infra.GetEnv("DB_USER", ""),
		Password: infra.GetEnv("DB_PASSWORD", ""),
		Database: infra.GetEnv("DB_NAME", ""),
		IP:       infra.GetEnv("DB_HOST", ""),
		Instance: infra.GetEnv("DB_INSTANCE", ""),
	}

	if infra.AppInfo.Env == "local" {
		dns.Instance = ""
	}

	Database, DatabaseSql = infra.ConnectMySqlDatabase(infra.CreateMySqlDSN(dns), maxOpenConns, maxIdleConns)
	return nil
}
