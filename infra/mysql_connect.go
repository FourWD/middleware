package infra

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ConnectMySqlDatabase opens MySQL using both gorm and database/sql, sets the
// session timezone to Asia/Bangkok, and applies the connection pool limits.
// Panics on connection failure (callers expect this at boot).
func ConnectMySqlDatabase(dns string, maxOpenConns int, maxIdleConns int) (*gorm.DB, *sql.DB) {
	dsn := dns + "&loc=Asia%2FBangkok"

	driverName := DBDriverMySQL
	if mysqlConnectorDSN(dsn) {
		if err := EnsureCloudSQLDriver(CloudSQLMySQLDriver); err != nil {
			AppLog.EventError(err, "DB_CLOUDSQL_REGISTER_FAILURE", nil, "",
				WithComponent(ComponentDB),
				WithOperation("connect"),
				WithLogKind(LogKindError))
			panic(err)
		}
		driverName = CloudSQLMySQLDriver
	}

	dataGorm, errGorm := gorm.Open(mysql.New(mysql.Config{DriverName: driverName, DSN: dsn}), &gorm.Config{
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

	dbSql, err := sql.Open(driverName, dsn)
	if err != nil {
		AppLog.EventError(err, "DB_MYSQL_CONNECT_FAILURE", nil, "",
			WithComponent(ComponentDB),
			WithOperation("connect"),
			WithLogKind(LogKindError))
		panic(err)
	}

	RegisterSQLDialect(dbSql, DBDriverMySQL)

	timeZone := "Asia/Bangkok"
	if _, err := dbSql.Exec("SET time_zone=?", timeZone); err != nil {
		AppLog.EventError(err, "DB_SET_TIMEZONE_FAILURE", nil, "",
			WithComponent(ComponentDB),
			WithOperation("set_timezone"),
			WithLogKind(LogKindError))
	}
	if err := dataGorm.Exec("SET time_zone=?", timeZone).Error; err != nil {
		AppLog.EventError(err, "DB_GORM_SET_TIMEZONE_FAILURE", nil, "",
			WithComponent(ComponentDB),
			WithOperation("set_timezone"),
			WithLogKind(LogKindError))
	}

	LogInfraEvent("DB_MYSQL_CONNECT_SUCCESS", ComponentDB, "connect", nil)
	initDatabaseConnectionPool(dbSql, maxOpenConns, maxIdleConns)

	return dataGorm, dbSql
}

func initDatabaseConnectionPool(dbSql *sql.DB, maxOpenConns int, maxIdleConns int) error {
	if dbSql == nil {
		return errors.New("database connection is nil")
	}

	dbSql.SetMaxOpenConns(maxOpenConns)
	dbSql.SetMaxIdleConns(maxIdleConns)
	dbSql.SetConnMaxLifetime(time.Hour)
	dbSql.SetConnMaxIdleTime(10 * time.Minute)

	LogInfraEvent("DB_POOL_INIT", ComponentDB, "init_pool", map[string]any{
		"max_open_conns":     maxOpenConns,
		"max_idle_conns":     maxIdleConns,
		"conn_max_lifetime":  "1h",
		"conn_max_idle_time": "10m",
	})

	return nil
}
