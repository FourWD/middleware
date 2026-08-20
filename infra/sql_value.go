package infra

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// nullDateMarker is the sentinel "default" date some legacy MySQL columns use
// in place of NULL. Rows starting with it are surfaced to clients as JSON null.
const nullDateMarker = "1900-01-01"

// bangkokTzSuffix is what %v appends to a time.Time in Asia/Bangkok, which is
// how the MySQL driver hands back DATETIME columns (loc=Asia/Bangkok).
const bangkokTzSuffix = " +0700 +07"

// timeLayoutTrimmed matches what %v prints for the date and clock portion:
// fractional seconds appear only when non-zero.
const timeLayoutTrimmed = "2006-01-02 15:04:05.999999999"

// ConvertSQLValue normalises a raw driver value for JSON output.
//
// dbType is the column's driver type name (see SQLColumnTypes). It is only
// consulted on Postgres, to tell TIMESTAMPTZ — an absolute instant, safe to
// move into Bangkok — from TIMESTAMP, a wall clock that pgx merely labels UTC
// and which must not be shifted.
//
// The MySQL branch is the original behaviour verbatim; do not route new
// normalisation through it.
func ConvertSQLValue(val any, dialect, dbType string) any {
	if b, ok := val.([]byte); ok {
		return string(b)
	}
	if val == nil {
		return nil
	}

	var s string
	if dialect == DBDriverPostgres {
		s = postgresValueString(val, dbType)
	} else {
		s = strings.ReplaceAll(fmt.Sprintf("%v", val), bangkokTzSuffix, "")
	}

	if strings.HasPrefix(s, nullDateMarker) {
		return nil
	}
	return s
}

func postgresValueString(val any, dbType string) string {
	switch v := val.(type) {
	case time.Time:
		if strings.EqualFold(dbType, "TIMESTAMPTZ") {
			if loc, err := bangkokLocation(); err == nil {
				v = v.In(loc)
			}
		}
		return v.Format(timeLayoutTrimmed)
	case bool:
		// Match MySQL's TINYINT(1), so a service switching backends keeps the
		// same response shape.
		if v {
			return "1"
		}
		return "0"
	}
	return fmt.Sprintf("%v", val)
}

// SQLColumnTypes returns the driver type name per column, for the dialects
// that consult them. ConvertSQLValue ignores dbType on MySQL, so nothing is
// fetched there and the result is nil — read it with SQLColumnTypeAt.
func SQLColumnTypes(rows *sql.Rows, columns int, dialect string) []string {
	if dialect != DBDriverPostgres {
		return nil
	}

	types, err := rows.ColumnTypes()
	if err != nil {
		return nil
	}

	names := make([]string, columns)
	for i := 0; i < columns && i < len(types); i++ {
		names[i] = types[i].DatabaseTypeName()
	}
	return names
}

// SQLColumnTypeAt is a bounds-safe read of an SQLColumnTypes result, which is
// nil for dialects that do not report types.
func SQLColumnTypeAt(types []string, i int) string {
	if i < len(types) {
		return types[i]
	}
	return ""
}
