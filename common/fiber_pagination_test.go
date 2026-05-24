package common

import (
	"strings"
	"testing"
)

func TestConvertSQLValue_ByteSliceToString(t *testing.T) {
	got := convertSQLValue([]byte("hello"))
	if got != "hello" {
		t.Fatalf("got %v want %q", got, "hello")
	}
}

func TestConvertSQLValue_NilPassesThrough(t *testing.T) {
	if got := convertSQLValue(nil); got != nil {
		t.Fatalf("got %v want nil", got)
	}
}

func TestConvertSQLValue_StripsTimezoneSuffix(t *testing.T) {
	got := convertSQLValue("2026-05-25 14:30:00 +0700 +07")
	if got != "2026-05-25 14:30:00" {
		t.Fatalf("got %q — tzSuffix not stripped", got)
	}
}

func TestConvertSQLValue_NullDateMarkerBecomesNil(t *testing.T) {
	cases := []string{
		"1900-01-01",
		"1900-01-01 00:00:00",
		"1900-01-01 00:00:00 +0700 +07",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			if got := convertSQLValue(in); got != nil {
				t.Fatalf("got %v want nil for null-date sentinel %q", got, in)
			}
		})
	}
}

func TestConvertSQLValue_NormalDateUnchanged(t *testing.T) {
	got := convertSQLValue("2024-01-15")
	if got != "2024-01-15" {
		t.Fatalf("got %v want %q", got, "2024-01-15")
	}
}

func TestConvertSQLValue_NumericStringifies(t *testing.T) {
	if got := convertSQLValue(42); got != "42" {
		t.Fatalf("got %v want %q", got, "42")
	}
}

func TestParseFullCount_AllTypes(t *testing.T) {
	cases := []struct {
		name string
		in   interface{}
		want int
	}{
		{"string numeric", "42", 42},
		{"string empty", "", 0},
		{"string non-numeric", "abc", 0},
		{"int64", int64(99), 99},
		{"int", 7, 7},
		{"unsupported type", 3.14, 0},
		{"nil", nil, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseFullCount(tc.in); got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
}

func TestRawSql_InterpolatesValues(t *testing.T) {
	got := rawSql("SELECT * FROM users WHERE id = ? AND name = ?", 7, "alice")
	want := "SELECT * FROM users WHERE id = '7' AND name = 'alice'"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRawSql_MismatchedValuesCountReturnsMessage(t *testing.T) {
	got := rawSql("SELECT * FROM users WHERE id = ? AND name = ?", 7)
	if !strings.Contains(got, "SQL and values count mismatch") {
		t.Fatalf("expected mismatch message, got %q", got)
	}
}

func TestRawSql_StripsNewlinesAndTabs(t *testing.T) {
	got := rawSql("SELECT *\nFROM users\tWHERE id = ?", 1)
	if strings.Contains(got, "\n") || strings.Contains(got, "\t") {
		t.Fatalf("newlines/tabs not stripped: %q", got)
	}
}
