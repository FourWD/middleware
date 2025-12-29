package common

import (
	"time"

	"github.com/google/uuid"
)

func InitDatabaseConnectionPool(maxOpenConns int, maxIdleConns int) error {
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

// เริ่ม monitoring connection pool ทุก 30 วินาที
func MonitorDatabasePool() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := GetDatabasePoolStats()

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

func GetDatabasePoolStats() map[string]interface{} {
	sqlDB, err := Database.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,          // จำนวน connections สูงสุดที่กำหนด
		"open_connections":     stats.OpenConnections,             // จำนวน connections ที่เปิดอยู่
		"in_use":               stats.InUse,                       // จำนวน connections ที่กำลังใช้งาน
		"idle":                 stats.Idle,                        // จำนวน connections ที่ idle
		"wait_count":           stats.WaitCount,                   // จำนวนครั้งที่ต้องรอ connection
		"wait_duration_ms":     stats.WaitDuration.Milliseconds(), // เวลารอรวม (มิลลิวินาที)
		"max_idle_closed":      stats.MaxIdleClosed,               // จำนวน connections ที่ปิดเพราะเกิน MaxIdleConns
		"max_idle_time_closed": stats.MaxIdleTimeClosed,           // จำนวน connections ที่ปิดเพราะ idle timeout
		"max_lifetime_closed":  stats.MaxLifetimeClosed,           // จำนวน connections ที่ปิดเพราะเกิน MaxLifetime
	}
}
