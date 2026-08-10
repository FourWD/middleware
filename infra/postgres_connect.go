package infra

import (
	"database/sql"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectPostgresDatabase opens Postgres using both gorm and database/sql,
// sets the session timezone to Asia/Bangkok, and applies the connection pool
// limits. Panics on connection failure (callers expect this at boot).
// Mirrors the structure of ConnectMySqlDatabase.
//
// Deprecated: ใช้ infra.OpenDB แทน ตัวนี้ใช้ lib/pq ขณะที่ OpenDB ใช้ pgx และไม่มี
// ผู้เรียกภายในโมดูลแล้ว
func ConnectPostgresDatabase(dns string, maxOpenConns int, maxIdleConns int) (*gorm.DB, *sql.DB) {
	dataGorm, errGorm := gorm.Open(postgres.Open(dns), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	if errGorm != nil {
		AppLog.EventError(errGorm, "DB_GORM_CONNECT_FAILURE", nil, "",
			WithComponent(ComponentDB),
			WithOperation("connect"),
			WithLogKind(LogKindError))
		panic(errGorm)
	}

	dbSql, err := sql.Open("postgres", dns)
	if err != nil {
		AppLog.EventError(err, "DB_POSTGRES_CONNECT_FAILURE", nil, "",
			WithComponent(ComponentDB),
			WithOperation("connect"),
			WithLogKind(LogKindError))
		panic(err)
	}

	// เดิมที่นี่ hardcode "SET TIME ZONE 'Asia/Bangkok'" ซึ่งขัดกับ default ของ
	// LoadDatabaseConfig (TimeZone=UTC) — โมดูลเดียวกันบอกคนละอย่าง ตอนนี้ปล่อยให้
	// timezone มาจากพารามิเตอร์ TimeZone ใน DSN ที่ผู้เรียกกำหนดเองแหล่งเดียว
	LogInfraEvent("DB_POSTGRES_CONNECT_SUCCESS", ComponentDB, "connect", nil)
	initDatabaseConnectionPool(dbSql, maxOpenConns, maxIdleConns)

	return dataGorm, dbSql
}
