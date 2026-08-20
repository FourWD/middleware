package infra

import (
	"database/sql"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectPostgresDatabase opens Postgres using both gorm and database/sql and
// applies the connection pool limits. The session timezone comes from the DSN.
// Panics on connection failure (callers expect this at boot).
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

	RegisterSQLDialect(dbSql, DBDriverPostgres)

	// Timezone comes from the caller's DSN alone. A hardcoded
	// "SET TIME ZONE" here would only bind one pooled connection, and it
	// previously contradicted the LoadDatabaseConfig default.
	LogInfraEvent("DB_POSTGRES_CONNECT_SUCCESS", ComponentDB, "connect", nil)
	initDatabaseConnectionPool(dbSql, maxOpenConns, maxIdleConns)

	return dataGorm, dbSql
}
