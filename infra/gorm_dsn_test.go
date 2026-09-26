package infra

import (
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestBuildPostgresDSNQuotesSpecialValues(t *testing.T) {
	t.Setenv("DB_INSTANCE_MODE", "")

	password := `p a'ss\word`
	cfg := postgresConfig()
	cfg.Password = password
	cfg.Name = ""

	dsn := BuildPostgresDSN(cfg)
	parsed, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("ParseConfig(%q): %v", dsn, err)
	}
	if parsed.Password != password {
		t.Errorf("password = %q, want %q", parsed.Password, password)
	}
	if parsed.Database != "" {
		t.Errorf("database = %q, want empty", parsed.Database)
	}
	if parsed.User != "app" || parsed.Host != "127.0.0.1" || parsed.Port != 5432 {
		t.Errorf("user/host/port = %s %s:%d", parsed.User, parsed.Host, parsed.Port)
	}
	if got := parsed.RuntimeParams["TimeZone"]; got != "UTC" {
		t.Errorf("timezone = %q, want UTC", got)
	}
}
