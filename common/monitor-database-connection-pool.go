package common

import (
	"time"

	"github.com/google/uuid"
)

func monitorDatabaseConnectionPool() {
	go func() {
		monitorDatabaseConnectionPoolLoop()
	}()
}

// Start monitoring connection pool every 30 seconds
func monitorDatabaseConnectionPoolLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := GetDatabaseConnectionPoolStats()

		logData := map[string]interface{}{
			"stats": stats,
		}
		Log("DB_POOL_STATS", logData, uuid.NewString())

		// If wait_count is high, connection pool is full
		if waitCount, ok := stats["wait_count"].(int64); ok && waitCount > 100 {
			logData["message"] = "High wait count - consider increasing MaxOpenConns"
			logData["wait_count"] = waitCount
			LogError("DB_POOL_WARNING", logData, uuid.NewString())
		}

		// If in_use is close to max_open_connections, pool is almost full
		if inUse, ok := stats["in_use"].(int); ok {
			if maxOpen, ok := stats["max_open_connections"].(int); ok {
				if inUse >= int(float64(maxOpen)*0.9) { // 90% usage
					logData["message"] = "Connection pool almost full"
					logData["in_use"] = inUse
					logData["max_open"] = maxOpen
					logData["usage_percent"] = float64(inUse) / float64(maxOpen) * 100
					LogError("DB_POOL_WARNING", logData, uuid.NewString())
				}
			}
		}
	}
}
