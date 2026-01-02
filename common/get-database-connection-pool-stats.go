package common

func GetDatabaseConnectionPoolStats() map[string]interface{} {
	sqlDB, err := Database.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,          // Maximum number of open connections allowed
		"open_connections":     stats.OpenConnections,             // Number of currently open connections
		"in_use":               stats.InUse,                       // Number of connections currently in use
		"idle":                 stats.Idle,                        // Number of idle connections
		"wait_count":           stats.WaitCount,                   // Number of times waited for a connection
		"wait_duration_ms":     stats.WaitDuration.Milliseconds(), // Total wait duration in milliseconds
		"max_idle_closed":      stats.MaxIdleClosed,               // Number of connections closed due to exceeding MaxIdleConns
		"max_idle_time_closed": stats.MaxIdleTimeClosed,           // Number of connections closed due to idle timeout
		"max_lifetime_closed":  stats.MaxLifetimeClosed,           // Number of connections closed due to exceeding MaxLifetime
	}
}
