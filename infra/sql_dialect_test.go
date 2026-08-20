package infra

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
)

// wrappedDriver stands in for an instrumentation driver (otelsql, sqlmw),
// whose package path tells us nothing about the underlying database.
type wrappedDriver struct{}

func (wrappedDriver) Open(string) (driver.Conn, error) { return nil, driver.ErrSkip }

func openLazy(t *testing.T, name string) *sql.DB {
	t.Helper()
	db, err := sql.Open(name, "")
	if err != nil {
		t.Skipf("driver %q unavailable: %v", name, err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSQLDialectDetectsByDriver(t *testing.T) {
	cases := []struct {
		driverName string
		want       string
	}{
		{"mysql", DBDriverMySQL},
		{"pgx", DBDriverPostgres},
		{"pgx/v5", DBDriverPostgres},
		{"postgres", DBDriverPostgres},
		{"clickhouse", DBDriverMySQL}, // uses `?` placeholders like MySQL
	}

	for _, tc := range cases {
		t.Run(tc.driverName, func(t *testing.T) {
			if got := SQLDialect(openLazy(t, tc.driverName)); got != tc.want {
				t.Fatalf("SQLDialect(%s) = %q, want %q", tc.driverName, got, tc.want)
			}
		})
	}
}

func TestSQLDialectNilDB(t *testing.T) {
	if got := SQLDialect(nil); got != DBDriverMySQL {
		t.Fatalf("SQLDialect(nil) = %q, want %q", got, DBDriverMySQL)
	}
}

func TestSQLDialectUnknownDriverFallsBackToMySQL(t *testing.T) {
	db := sql.OpenDB(wrappedConnector{})
	t.Cleanup(func() { db.Close() })

	if got := SQLDialect(db); got != DBDriverMySQL {
		t.Fatalf("SQLDialect(wrapped) = %q, want %q — detection must bias to MySQL", got, DBDriverMySQL)
	}
}

func TestRegisterSQLDialectWinsOverDetection(t *testing.T) {
	db := sql.OpenDB(wrappedConnector{})
	t.Cleanup(func() { db.Close() })

	RegisterSQLDialect(db, DBDriverPostgres)
	if got := SQLDialect(db); got != DBDriverPostgres {
		t.Fatalf("SQLDialect after register = %q, want %q", got, DBDriverPostgres)
	}
}

func TestRegisterSQLDialectRejectsUnknownName(t *testing.T) {
	db := sql.OpenDB(wrappedConnector{})
	t.Cleanup(func() { db.Close() })

	RegisterSQLDialect(db, "oracle")
	if got := SQLDialect(db); got != DBDriverMySQL {
		t.Fatalf("SQLDialect after bogus register = %q, want %q", got, DBDriverMySQL)
	}
}

func TestRegisterSQLDialectNilIsNoop(t *testing.T) {
	RegisterSQLDialect(nil, DBDriverPostgres) // must not panic
}

func TestPrepareSQL(t *testing.T) {
	mysqlDB := openLazy(t, "mysql")
	pgDB := openLazy(t, "pgx")

	const in = "SELECT * FROM t WHERE a = ? AND b = ?"

	if got := PrepareSQL(mysqlDB, in); got != in {
		t.Fatalf("MySQL statement rewritten: %s", got)
	}

	want := "SELECT * FROM t WHERE a = $1 AND b = $2"
	if got := PrepareSQL(pgDB, in); got != want {
		t.Fatalf("\n got: %s\nwant: %s", got, want)
	}
}

type wrappedConnector struct{}

func (wrappedConnector) Connect(context.Context) (driver.Conn, error) { return nil, driver.ErrSkip }
func (wrappedConnector) Driver() driver.Driver                        { return wrappedDriver{} }
