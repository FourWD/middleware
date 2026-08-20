package common

import (
	"strings"
	"testing"
)

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
