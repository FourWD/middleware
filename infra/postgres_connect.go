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

	// Postgres SET TIME ZONE does not accept bind parameters — the value
	// must be embedded as a literal.
	const setTZ = "SET TIME ZONE 'Asia/Bangkok'"
	if _, err := dbSql.Exec(setTZ); err != nil {
		AppLog.EventError(err, "DB_SET_TIMEZONE_FAILURE", nil, "",
			WithComponent(ComponentDB),
			WithOperation("set_timezone"),
			WithLogKind(LogKindError))
	}
	if err := dataGorm.Exec(setTZ).Error; err != nil {
		AppLog.EventError(err, "DB_GORM_SET_TIMEZONE_FAILURE", nil, "",
			WithComponent(ComponentDB),
			WithOperation("set_timezone"),
			WithLogKind(LogKindError))
	}

	LogInfraEvent("DB_POSTGRES_CONNECT_SUCCESS", ComponentDB, "connect", nil)
	initDatabaseConnectionPool(dbSql, maxOpenConns, maxIdleConns)

	return dataGorm, dbSql
}
