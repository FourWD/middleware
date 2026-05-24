package infra

import "fmt"

// DNS holds the fields needed to build a MySQL connection string.
// The name is historical (was originally meant to be DSN).
type DNS struct {
	Username string
	Password string
	Database string
	IP       string
	Instance string
	IsMySql5 bool
}

// CreateMySqlDSN builds a MySQL DSN. When Instance is set the connection
// uses the Cloud SQL Unix socket; otherwise it uses tcp. Sets &loc=Local
// on tcp connections only.
func CreateMySqlDSN(dsn DNS) string {
	var protocol string
	setting := "?charset=utf8mb4&parseTime=True"

	if dsn.IsMySql5 {
		setting = "?charset=utf8&parseTime=True"
	}

	if dsn.Instance != "" {
		protocol = fmt.Sprintf("unix(/cloudsql/%s)", dsn.Instance)
	} else {
		protocol = fmt.Sprintf("tcp(%s:3306)", dsn.IP)
		setting += "&loc=Local"
	}
	LogInfraEvent("DB_DSN_BUILT", ComponentDB, "build_dsn", map[string]any{"database": dsn.Database})
	return fmt.Sprintf("%s:%s@%s/%s%s", dsn.Username, dsn.Password, protocol, dsn.Database, setting)
}
