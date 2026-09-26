package infra

import "testing"

func TestWithMultiStatements(t *testing.T) {
	tests := map[string]string{
		"":                                     "multiStatements=true",
		"charset=utf8mb4&parseTime=True":       "charset=utf8mb4&parseTime=True&multiStatements=true",
		"?charset=utf8mb4":                     "charset=utf8mb4&multiStatements=true",
		"multiStatements=true&parseTime=True":  "multiStatements=true&parseTime=True",
		"parseTime=True&multistatements=false": "parseTime=True&multistatements=false",
	}
	for in, want := range tests {
		if got := withMultiStatements(in); got != want {
			t.Errorf("withMultiStatements(%q) = %q, want %q", in, got, want)
		}
	}
}
