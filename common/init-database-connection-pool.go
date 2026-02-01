package common

import (
	"database/sql"
	"errors"
	"time"
)

func initDatabaseConnectionPool(dbSql *sql.DB, maxOpenConns int, maxIdleConns int) error {
	if dbSql == nil {
		return errors.New("database connection is nil")
	}

	dbSql.SetMaxOpenConns(maxOpenConns)
	dbSql.SetMaxIdleConns(maxIdleConns)
	dbSql.SetConnMaxLifetime(time.Hour)
	dbSql.SetConnMaxIdleTime(10 * time.Minute)

	logData := map[string]interface{}{
		"max_open_conns":     maxOpenConns,
		"max_idle_conns":     maxIdleConns,
		"conn_max_lifetime":  "1h",
		"conn_max_idle_time": "10m",
		"message":            "Database connection pool configured",
	}
	Log("DATABASE_POOL_INIT", logData, "system")

	return nil
}
