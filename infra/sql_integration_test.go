package infra

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// These tests need a live database. They are skipped unless MW_IT_PG_* /
// MW_IT_MY_* point at one, so `go test ./...` stays green without docker.
//
//	MW_IT_PG_HOST=127.0.0.1 MW_IT_PG_PORT=55432 go test ./infra/ -run Integration

func itConfig(t *testing.T, prefix, driver string) DatabaseConfig {
	t.Helper()

	host := os.Getenv("MW_IT_" + prefix + "_HOST")
	if host == "" {
		t.Skipf("set MW_IT_%s_HOST to run integration tests", prefix)
	}

	t.Setenv("DB_DRIVER", driver)
	t.Setenv("DB_HOST", host)
	t.Setenv("DB_PORT", os.Getenv("MW_IT_"+prefix+"_PORT"))
	t.Setenv("DB_USER", os.Getenv("MW_IT_"+prefix+"_USER"))
	t.Setenv("DB_PASSWORD", os.Getenv("MW_IT_"+prefix+"_PASSWORD"))
	t.Setenv("DB_NAME", os.Getenv("MW_IT_"+prefix+"_NAME"))

	return LoadDatabaseConfig()
}

func itOpen(t *testing.T, cfg DatabaseConfig) *sql.DB {
	t.Helper()

	AppLog = NewLoggerWith(CommonConfig{})
	db, err := OpenDB(cfg, AppLog)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("resolve sql.DB: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	return sqlDB
}

func itRows(t *testing.T, db *sql.DB, query string, args ...any) []map[string]any {
	t.Helper()

	jsonBytes, _, err := queryToJSON(db, query, args...)
	if err != nil {
		t.Fatalf("queryToJSON(%s): %v", query, err)
	}
	var out []map[string]any
	if err := json.Unmarshal(jsonBytes, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return out
}

// --- Postgres ---------------------------------------------------------------

func TestIntegrationPostgresSessionTimeZone(t *testing.T) {
	cfg := itConfig(t, "PG", DBDriverPostgres)
	db := itOpen(t, cfg)

	rows := itRows(t, db, "SHOW timezone")
	if len(rows) != 1 {
		t.Fatalf("SHOW timezone returned %d rows", len(rows))
	}
	if got := rows[0]["TimeZone"]; got != "Asia/Bangkok" {
		t.Fatalf("session timezone = %#v, want Asia/Bangkok (row: %#v)", got, rows[0])
	}
}

func TestIntegrationPostgresDialectDetected(t *testing.T) {
	cfg := itConfig(t, "PG", DBDriverPostgres)
	db := itOpen(t, cfg)

	if got := SQLDialect(db); got != DBDriverPostgres {
		t.Fatalf("SQLDialect = %q, want %q", got, DBDriverPostgres)
	}
}

func TestIntegrationPostgresPlaceholders(t *testing.T) {
	cfg := itConfig(t, "PG", DBDriverPostgres)
	db := itOpen(t, cfg)

	rows := itRows(t,
		db,
		"SELECT ?::int AS a, ?::text AS b, 'lit?eral' AS c -- trailing ? comment",
		7, "hello",
	)
	if len(rows) != 1 {
		t.Fatalf("got %d rows", len(rows))
	}
	if rows[0]["a"] != "7" || rows[0]["b"] != "hello" || rows[0]["c"] != "lit?eral" {
		t.Fatalf("row = %#v", rows[0])
	}
}

func TestIntegrationPostgresColumnTypesAndValues(t *testing.T) {
	cfg := itConfig(t, "PG", DBDriverPostgres)
	db := itOpen(t, cfg)

	if _, err := db.Exec(`
		DROP TABLE IF EXISTS mw_types;
		CREATE TABLE mw_types (
			id          serial PRIMARY KEY,
			ts          timestamp,
			tstz        timestamptz,
			flag        boolean,
			amount      numeric(12,2),
			uid         uuid,
			payload     jsonb,
			note        text,
			legacy_date date
		);
		INSERT INTO mw_types (ts, tstz, flag, amount, uid, payload, note, legacy_date)
		VALUES (
			'2026-08-20 10:00:00',
			'2026-08-20 03:00:00+00',
			true,
			1234.56,
			'11111111-2222-3333-4444-555555555555',
			'{"k":"v"}',
			'plain',
			'1900-01-01'
		);`); err != nil {
		t.Fatalf("seed: %v", err)
	}

	rows := itRows(t, db, "SELECT * FROM mw_types WHERE id = ?", 1)
	if len(rows) != 1 {
		t.Fatalf("got %d rows", len(rows))
	}
	got := rows[0]

	checks := map[string]any{
		// wall clock, must NOT be shifted
		"ts": "2026-08-20 10:00:00",
		// absolute instant 03:00Z rendered in Bangkok, matching MySQL
		"tstz": "2026-08-20 10:00:00",
		// same shape as MySQL TINYINT(1)
		"flag":   "1",
		"amount": "1234.56",
		"uid":    "11111111-2222-3333-4444-555555555555",
		"note":   "plain",
		// 1900-01-01 sentinel surfaces as null
		"legacy_date": nil,
	}
	for col, want := range checks {
		if got[col] != want {
			t.Errorf("%s = %#v, want %#v", col, got[col], want)
		}
	}
	if got["payload"] == nil {
		t.Errorf("payload = nil, want jsonb text")
	}

	t.Logf("full row: %#v", got)
}

func TestIntegrationPostgresReportsDatabaseTypeNames(t *testing.T) {
	cfg := itConfig(t, "PG", DBDriverPostgres)
	db := itOpen(t, cfg)

	rows, err := db.Query("SELECT now() AS tstz, now()::timestamp AS ts")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	names := SQLColumnTypes(rows, 2, DBDriverPostgres)
	if names[0] != "TIMESTAMPTZ" || names[1] != "TIMESTAMP" {
		t.Fatalf("DatabaseTypeName = %#v, want [TIMESTAMPTZ TIMESTAMP]", names)
	}
}

func TestIntegrationPostgresPaginationShape(t *testing.T) {
	cfg := itConfig(t, "PG", DBDriverPostgres)
	db := itOpen(t, cfg)

	// The exact wrapper FiberPaginatedQuery builds.
	base := "SELECT ?::int AS n"
	paginated := fmt.Sprintf(
		"SELECT *, count(*) OVER() AS full_count FROM (%s) AS sub LIMIT %d OFFSET %d",
		base, 10, 0,
	)

	rows := itRows(t, db, paginated, 5)
	if len(rows) != 1 {
		t.Fatalf("got %d rows", len(rows))
	}
	if rows[0]["n"] != "5" || rows[0]["full_count"] != "1" {
		t.Fatalf("row = %#v", rows[0])
	}
}

// --- MySQL: proves the existing behaviour is untouched -----------------------

func TestIntegrationMySQLUnchanged(t *testing.T) {
	cfg := itConfig(t, "MY", DBDriverMySQL)
	db := itOpen(t, cfg)

	if got := SQLDialect(db); got != DBDriverMySQL {
		t.Fatalf("SQLDialect = %q, want %q", got, DBDriverMySQL)
	}

	if _, err := db.Exec(`DROP TABLE IF EXISTS mw_types`); err != nil {
		t.Fatalf("drop: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE mw_types (
			id     int PRIMARY KEY,
			dt     datetime(3),
			flag   tinyint(1),
			amount decimal(12,2),
			note   varchar(50),
			legacy date
		)`); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO mw_types VALUES
		(1, '2026-08-20 10:00:00.123', 1, 1234.56, 'plain', '1900-01-01')`); err != nil {
		t.Fatalf("seed: %v", err)
	}

	rows := itRows(t, db, "SELECT * FROM mw_types WHERE id = ?", 1)
	if len(rows) != 1 {
		t.Fatalf("got %d rows", len(rows))
	}
	got := rows[0]

	checks := map[string]any{
		"dt":     "2026-08-20 10:00:00.123", // fractional seconds preserved
		"flag":   "1",
		"amount": "1234.56",
		"note":   "plain",
		// parseTime=True makes DATE a time.Time, so the 1900-01-01 sentinel
		// applies here too — same as before, and same as Postgres.
		"legacy": nil,
	}
	for col, want := range checks {
		if got[col] != want {
			t.Errorf("%s = %#v, want %#v", col, got[col], want)
		}
	}

	t.Logf("full row: %#v", got)
}

func TestIntegrationMySQLQuestionMarkInLiteral(t *testing.T) {
	cfg := itConfig(t, "MY", DBDriverMySQL)
	db := itOpen(t, cfg)

	rows := itRows(t, db, "SELECT ? AS a, 'lit?eral' AS b", 7)
	if len(rows) != 1 || rows[0]["a"] != "7" || rows[0]["b"] != "lit?eral" {
		t.Fatalf("rows = %#v", rows)
	}
}
