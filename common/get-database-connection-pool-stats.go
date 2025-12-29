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
