package kit

import "fmt"

// PostgresDSN holds the fields needed to build a Postgres connection string.
type PostgresDSN struct {
	Username string
	Password string
	Database string
	Host     string
}

// CreatePostgresDSN builds a Postgres DSN string with port 5432 and
// sslmode=disable. Mirrors the prior common.CreatePostgresDSN behavior.
func CreatePostgresDSN(dsn PostgresDSN) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		dsn.Host, dsn.Username, dsn.Password, dsn.Database, 5432)
}
