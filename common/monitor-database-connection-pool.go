package common

import (
	"time"

	"github.com/google/uuid"
)

func MonitorDatabaseConnectionPool() {
	go func() {
		monitorDatabaseConnectionPool()
	}()
}

// เริ่ม monitoring connection pool ทุก 30 วินาที
func monitorDatabaseConnectionPool() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := GetDatabaseConnectionPoolStats()

		logData := map[string]interface{}{
			"stats": stats,
		}
		Log("DB_POOL_STATS", logData, uuid.NewString())

		// ถ้า wait_count สูง แสดงว่า connection pool เต็ม
		if waitCount, ok := stats["wait_count"].(int64); ok && waitCount > 100 {
			logData["message"] = "High wait count - consider increasing MaxOpenConns"
			logData["wait_count"] = waitCount
			LogError("DB_POOL_WARNING", logData, uuid.NewString())
		}

		// ถ้า in_use ใกล้เคียง max_open_connections แสดงว่าใกล้เต็ม
		if inUse, ok := stats["in_use"].(int); ok {
			if maxOpen, ok := stats["max_open_connections"].(int); ok {
				if inUse >= int(float64(maxOpen)*0.9) { // ใช้ไป 90%
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
