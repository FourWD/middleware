package infra

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// legacyConvert is the pre-Postgres implementation, kept verbatim as the
// oracle for the MySQL branch. Any divergence is a regression.
func legacyConvert(val any) any {
	if b, ok := val.([]byte); ok {
		return string(b)
	}
	if val == nil {
		return nil
	}
	temp := strings.Replace(fmt.Sprintf("%v", val), " +0700 +07", "", -1)
	if len(temp) >= 10 && temp[0:10] == "1900-01-01" {
		return nil
	}
	return temp
}

func bangkok(t *testing.T) *time.Location {
	t.Helper()
	loc, err := bangkokLocation()
	if err != nil {
		t.Skipf("Asia/Bangkok tzdata unavailable: %v", err)
	}
	return loc
}

func TestConvertSQLValueMySQLMatchesLegacy(t *testing.T) {
	loc := bangkok(t)

	values := []any{
		nil,
		[]byte("hello"),
		[]byte("1900-01-01"),
		"2026-05-25 14:30:00 +0700 +07",
		"1900-01-01",
		"1900-01-01 00:00:00 +0700 +07",
		"2024-01-15",
		42,
		int64(42),
		float64(1.5),
		true,
		time.Date(2026, 8, 20, 10, 0, 0, 0, loc),
		time.Date(2026, 8, 20, 10, 0, 0, 123000000, loc),
		time.Date(1900, 1, 1, 0, 0, 0, 0, loc),
	}

	for i, v := range values {
		t.Run(fmt.Sprintf("%d_%T", i, v), func(t *testing.T) {
			want := legacyConvert(v)
			// dbType is populated on MySQL too and must never change the result.
			for _, dbType := range []string{"", "DATETIME", "TIMESTAMP", "TIMESTAMPTZ"} {
				if got := ConvertSQLValue(v, DBDriverMySQL, dbType); got != want {
					t.Fatalf("dbType=%q: got %#v, want %#v", dbType, got, want)
				}
			}
		})
	}
}

func TestConvertSQLValueMySQLKeepsFractionalSeconds(t *testing.T) {
	loc := bangkok(t)
	v := time.Date(2026, 8, 20, 10, 0, 0, 123000000, loc)

	if got := ConvertSQLValue(v, DBDriverMySQL, "DATETIME"); got != "2026-08-20 10:00:00.123" {
		t.Fatalf("got %#v — DATETIME(3) precision lost", got)
	}
}

func TestConvertSQLValuePostgresTimestampIsNotShifted(t *testing.T) {
	// pgx hands back a plain TIMESTAMP as a wall clock labelled UTC. Moving it
	// into Bangkok would invent a 7-hour shift.
	v := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)

	if got := ConvertSQLValue(v, DBDriverPostgres, "TIMESTAMP"); got != "2026-08-20 10:00:00" {
		t.Fatalf("got %#v, want %q", got, "2026-08-20 10:00:00")
	}
}

func TestConvertSQLValuePostgresTimestamptzMovesToBangkok(t *testing.T) {
	bangkok(t)
	// An absolute instant: 03:00 UTC is 10:00 in Bangkok, matching MySQL.
	v := time.Date(2026, 8, 20, 3, 0, 0, 0, time.UTC)

	if got := ConvertSQLValue(v, DBDriverPostgres, "TIMESTAMPTZ"); got != "2026-08-20 10:00:00" {
		t.Fatalf("got %#v, want %q", got, "2026-08-20 10:00:00")
	}
}

func TestConvertSQLValuePostgresBoolMatchesMySQLShape(t *testing.T) {
	if got := ConvertSQLValue(true, DBDriverPostgres, "BOOL"); got != "1" {
		t.Fatalf("true -> %#v, want %q", got, "1")
	}
	if got := ConvertSQLValue(false, DBDriverPostgres, "BOOL"); got != "0" {
		t.Fatalf("false -> %#v, want %q", got, "0")
	}
}

func TestConvertSQLValuePostgresSentinelStillNulls(t *testing.T) {
	v := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

	if got := ConvertSQLValue(v, DBDriverPostgres, "TIMESTAMP"); got != nil {
		t.Fatalf("got %#v, want nil", got)
	}
}

func TestConvertSQLValuePostgresUnknownColumnTypeDoesNotShift(t *testing.T) {
	// An empty dbType means the driver could not report one — treat it as a
	// wall clock rather than guessing and shifting.
	v := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)

	if got := ConvertSQLValue(v, DBDriverPostgres, ""); got != "2026-08-20 10:00:00" {
		t.Fatalf("got %#v, want %q", got, "2026-08-20 10:00:00")
	}
}

// pgx attaches a location to a plain TIMESTAMP but never shifts the clock:
// gorm registers a Bangkok ScanLocation on the direct path and skips it on the
// Cloud SQL connector path, leaving UTC. Both must render identically.
func TestConvertSQLValuePostgresTimestampIsLabelIndependent(t *testing.T) {
	loc := bangkok(t)

	labelled := map[string]time.Time{
		"connector (UTC label)":  time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC),
		"direct (Bangkok label)": time.Date(2026, 8, 20, 10, 0, 0, 0, loc),
		"neither (New York label)": time.Date(2026, 8, 20, 10, 0, 0, 0,
			time.FixedZone("EST", -5*60*60)),
	}

	for name, v := range labelled {
		t.Run(name, func(t *testing.T) {
			if got := ConvertSQLValue(v, DBDriverPostgres, "TIMESTAMP"); got != "2026-08-20 10:00:00" {
				t.Fatalf("got %#v, want %q", got, "2026-08-20 10:00:00")
			}
		})
	}
}

// A TIMESTAMPTZ is an absolute instant, so the location pgx happens to attach
// (time.Local, i.e. the container's TZ) must not reach the response.
func TestConvertSQLValuePostgresTimestamptzIgnoresProcessTimezone(t *testing.T) {
	bangkok(t)
	instant := time.Date(2026, 8, 20, 3, 0, 0, 0, time.UTC)

	for _, zone := range []*time.Location{
		time.UTC,
		time.FixedZone("EST", -5*60*60),
		time.FixedZone("+14", 14*60*60),
	} {
		t.Run(zone.String(), func(t *testing.T) {
			got := ConvertSQLValue(instant.In(zone), DBDriverPostgres, "TIMESTAMPTZ")
			if got != "2026-08-20 10:00:00" {
				t.Fatalf("got %#v, want %q", got, "2026-08-20 10:00:00")
			}
		})
	}
}
