package infra

import (
	"database/sql"
	"reflect"
	"strings"
	"sync"

	"github.com/FourWD/middleware/kit"
)

// sqlDialects maps an open *sql.DB to the dialect it speaks. Entries are never
// evicted — a *sql.DB is expected to live as long as the process.
var sqlDialects sync.Map

// postgresDriverPkgs is an allowlist of driver package paths, matched by
// prefix. Anything unrecognised is treated as MySQL so an unknown driver
// degrades to "Postgres unsupported" rather than "MySQL broken".
var postgresDriverPkgs = []string{
	"github.com/jackc/pgx",
	"github.com/lib/pq",
	"cloud.google.com/go/cloudsqlconn/postgres",
}

// RegisterSQLDialect records the dialect db speaks. The connect helpers call
// this; callers that build their own *sql.DB — notably one wrapped by an
// instrumentation driver, which defeats detection — should call it too.
// Never fails: an unrecognised driver is recorded as MySQL.
func RegisterSQLDialect(db *sql.DB, driver string) {
	if db == nil {
		return
	}
	if driver != DBDriverPostgres {
		driver = DBDriverMySQL
	}
	sqlDialects.Store(db, driver)
}

// SQLDialect returns the dialect for db, falling back to the driver's package
// path when db was never registered. Defaults to MySQL.
func SQLDialect(db *sql.DB) string {
	if db == nil {
		return DBDriverMySQL
	}
	if v, ok := sqlDialects.Load(db); ok {
		return v.(string)
	}

	dialect := detectSQLDialect(db)
	sqlDialects.Store(db, dialect)
	return dialect
}

func detectSQLDialect(db *sql.DB) string {
	driver := db.Driver()
	if driver == nil {
		return DBDriverMySQL
	}

	t := reflect.TypeOf(driver)
	if t == nil {
		return DBDriverMySQL
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	pkg := t.PkgPath()
	for _, p := range postgresDriverPkgs {
		if strings.HasPrefix(pkg, p) {
			return DBDriverPostgres
		}
	}
	return DBDriverMySQL
}

// PrepareSQL returns sqlText in the placeholder style db expects. MySQL keeps
// the `?` form untouched; Postgres gets positional `$N`.
func PrepareSQL(db *sql.DB, sqlText string) string {
	if SQLDialect(db) != DBDriverPostgres {
		return sqlText
	}
	return kit.ToPostgresPlaceholders(sqlText)
}
