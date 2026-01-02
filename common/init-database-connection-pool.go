package common

import "time"

func initDatabaseConnectionPool(maxOpenConns int, maxIdleConns int) error {
	sqlDB, err := Database.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

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
