package common

import (
	"database/sql"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgresDatabase(dns string) (*gorm.DB, *sql.DB) {
	dbGorm, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	Log("DB_POSTGRES_CONNECTION_SUCCESS", map[string]interface{}{}, "")

	db, err := dbGorm.DB()
	if err != nil {
		log.Fatal("Failed to get DB from GORM: ", err)
	}

	return dbGorm, db
}
