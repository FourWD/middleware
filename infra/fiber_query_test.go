package infra

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
)

var errFakeRows = errors.New("connection reset mid-stream")

// failingRowsDriver yields one row, then fails, to simulate a stream that
// breaks after partial results.
type failingRowsDriver struct{}

func (failingRowsDriver) Open(string) (driver.Conn, error) { return failingRowsConn{}, nil }

type failingRowsConn struct{}

func (failingRowsConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (failingRowsConn) Close() error                        { return nil }
func (failingRowsConn) Begin() (driver.Tx, error)           { return nil, errors.New("unsupported") }
func (failingRowsConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &failingRows{}, nil
}

type failingRows struct{ n int }

func (*failingRows) Columns() []string { return []string{"id"} }
func (*failingRows) Close() error      { return nil }
func (r *failingRows) Next(dest []driver.Value) error {
	r.n++
	if r.n == 1 {
		dest[0] = int64(1)
		return nil
	}
	if r.n == 2 {
		return errFakeRows
	}
	return io.EOF
}

func init() { sql.Register("mw_failing_rows", failingRowsDriver{}) }

func TestQueryToJSONReturnsRowsErr(t *testing.T) {
	db, err := sql.Open("mw_failing_rows", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	RegisterSQLDialect(db, DBDriverMySQL)

	if _, _, err := queryToJSON(context.Background(), db, "SELECT id FROM t"); !errors.Is(err, errFakeRows) {
		t.Fatalf("err = %v, want %v", err, errFakeRows)
	}
}
