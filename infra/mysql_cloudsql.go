package infra

import "database/sql"

// ConnectDatabaseMySqlGoogle opens a sql.DB to Cloud SQL MySQL (no gorm).
// Forces dsn.Instance to "" in local env so the driver falls back to tcp.
// Sets session timezone to Asia/Bangkok. Returns the bare *sql.DB.
func ConnectDatabaseMySqlGoogle(dsn DNS) (*sql.DB, error) {
	if AppInfo.Env == "local" {
		dsn.Instance = ""
	}

	rawDSN := CreateMySqlDSN(dsn)

	database, err := sql.Open("mysql", rawDSN+"&loc=Asia%2FBangkok")
	if err != nil {
		AppLog.EventError(err, "DB_MYSQL_CLOUDSQL_CONNECT_FAILURE", nil, "",
			WithComponent(ComponentDB),
			WithOperation("connect"),
			WithLogKind(LogKindError))
		return nil, err
	}

	database.Exec("SET time_zone=?", "Asia/Bangkok")

	LogInfraEvent("DB_MYSQL_CLOUDSQL_CONNECT_SUCCESS", ComponentDB, "connect", nil)

	return database, nil
}
