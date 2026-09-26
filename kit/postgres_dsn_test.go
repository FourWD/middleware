package kit

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestCreatePostgresDSNPlainUnchanged(t *testing.T) {
	got := CreatePostgresDSN(PostgresDSN{Username: "app", Password: "secret", Database: "db", Host: "127.0.0.1"})
	want := "host=127.0.0.1 user=app password=secret dbname=db port=5432 sslmode=prefer"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestCreatePostgresDSNQuotesSpecialValues(t *testing.T) {
	password := `p a'ss\word`
	dsn := CreatePostgresDSN(PostgresDSN{Username: "app", Password: password, Host: "127.0.0.1"})

	cfg, err := pgconn.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("ParseConfig(%q): %v", dsn, err)
	}
	if cfg.Password != password {
		t.Errorf("password = %q, want %q", cfg.Password, password)
	}
	if cfg.User != "app" {
		t.Errorf("user = %q, want app", cfg.User)
	}
	if cfg.Database != "" {
		t.Errorf("database = %q, want empty", cfg.Database)
	}
	if cfg.Host != "127.0.0.1" || cfg.Port != 5432 {
		t.Errorf("host/port = %s:%d", cfg.Host, cfg.Port)
	}
}

func TestQuotePostgresDSNValue(t *testing.T) {
	tests := map[string]string{
		"plain":           "plain",
		"":                "''",
		"a b":             "'a b'",
		`a'b`:             `'a\'b'`,
		`a\b`:             `'a\\b'`,
		"k=v":             "'k=v'",
		"/cloudsql/p:r:i": "/cloudsql/p:r:i",
	}
	for in, want := range tests {
		if got := QuotePostgresDSNValue(in); got != want {
			t.Errorf("QuotePostgresDSNValue(%q) = %q, want %q", in, got, want)
		}
	}
}
