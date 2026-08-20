package common

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FourWD/middleware/infra"
	"github.com/gofiber/fiber/v3"
)

// Drives FiberPaginatedQuery end to end against a live database. Skipped
// unless MW_IT_PG_* / MW_IT_MY_* are set, so `go test ./...` needs no docker.

func itPaginationDB(t *testing.T, prefix, driver string) *sql.DB {
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

	infra.AppLog = infra.NewLoggerWith(infra.CommonConfig{})
	db, err := infra.OpenDB(infra.LoadDatabaseConfig(), infra.AppLog)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("resolve sql.DB: %v", err)
	}

	prev := DatabaseSql
	DatabaseSql = sqlDB
	t.Cleanup(func() {
		DatabaseSql = prev
		sqlDB.Close()
	})
	return sqlDB
}

type paginationResponse struct {
	Status     int              `json:"status"`
	Data       []map[string]any `json:"data"`
	Pagination struct {
		CurrentPage int `json:"currentPage"`
		PageSize    int `json:"pageSize"`
		TotalItems  int `json:"totalItems"`
		TotalPages  int `json:"totalPages"`
	} `json:"pagination"`
}

func runPagination(t *testing.T, target, baseSQL string, values ...any) paginationResponse {
	t.Helper()

	app := fiber.New()
	app.Get("/list", func(c fiber.Ctx) error {
		return FiberPaginatedQuery(c, baseSQL, values...)
	})

	resp, err := app.Test(httptest.NewRequest("GET", target, nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("status %d: %s", resp.StatusCode, body)
	}

	var out paginationResponse
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal %s: %v", body, err)
	}
	return out
}

func seedNumbers(t *testing.T, db *sql.DB, driver string) {
	t.Helper()

	if _, err := db.Exec(`DROP TABLE IF EXISTS mw_page`); err != nil {
		t.Fatalf("drop: %v", err)
	}

	ddl := `CREATE TABLE mw_page (id int PRIMARY KEY, grp varchar(10), note varchar(50))`
	if driver == infra.DBDriverPostgres {
		ddl = `CREATE TABLE mw_page (id int PRIMARY KEY, grp text, note text)`
	}
	if _, err := db.Exec(ddl); err != nil {
		t.Fatalf("create: %v", err)
	}

	for i := 1; i <= 25; i++ {
		grp := "a"
		if i%2 == 0 {
			grp = "b"
		}
		if _, err := db.Exec(`INSERT INTO mw_page VALUES (`+ph(driver, 1)+`, `+ph(driver, 2)+`, `+ph(driver, 3)+`)`,
			i, grp, "note?mark"); err != nil {
			t.Fatalf("insert %d: %v", i, err)
		}
	}
}

// ph builds a driver-native placeholder — this helper seeds directly, without
// going through PrepareSQL, so the test does not depend on what it verifies.
func ph(driver string, n int) string {
	if driver == infra.DBDriverPostgres {
		return "$" + string(rune('0'+n))
	}
	return "?"
}

func assertPagination(t *testing.T, driver string) {
	t.Helper()

	db := itPaginationDB(t, map[string]string{
		infra.DBDriverPostgres: "PG",
		infra.DBDriverMySQL:    "MY",
	}[driver], driver)
	seedNumbers(t, db, driver)

	// Two placeholders, plus a literal '?' the scanner must not touch.
	base := "SELECT id, note FROM mw_page WHERE grp = ? AND id <= ? ORDER BY id"

	got := runPagination(t, "/list?page=2&limit=5", base, "a", 25)

	// grp='a' is the odd ids: 1,3,..,25 => 13 rows. Page 2 of 5 => ids 11..19.
	if got.Pagination.TotalItems != 13 {
		t.Errorf("totalItems = %d, want 13", got.Pagination.TotalItems)
	}
	if got.Pagination.TotalPages != 3 {
		t.Errorf("totalPages = %d, want 3", got.Pagination.TotalPages)
	}
	if got.Pagination.CurrentPage != 2 || got.Pagination.PageSize != 5 {
		t.Errorf("page/limit = %d/%d, want 2/5", got.Pagination.CurrentPage, got.Pagination.PageSize)
	}
	if len(got.Data) != 5 {
		t.Fatalf("len(data) = %d, want 5: %#v", len(got.Data), got.Data)
	}
	if got.Data[0]["id"] != "11" || got.Data[4]["id"] != "19" {
		t.Errorf("ids = %v .. %v, want 11 .. 19", got.Data[0]["id"], got.Data[4]["id"])
	}
	if got.Data[0]["note"] != "note?mark" {
		t.Errorf("note = %#v, want %q", got.Data[0]["note"], "note?mark")
	}
	if _, leaked := got.Data[0]["full_count"]; !leaked {
		t.Errorf("full_count column missing from row: %#v", got.Data[0])
	}
}

func TestIntegrationPaginationPostgres(t *testing.T) {
	assertPagination(t, infra.DBDriverPostgres)
}

func TestIntegrationPaginationMySQL(t *testing.T) {
	assertPagination(t, infra.DBDriverMySQL)
}
