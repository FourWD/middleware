package infra

import "testing"

func mysqlConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver:   DBDriverMySQL,
		Host:     "127.0.0.1",
		Port:     3306,
		User:     "app",
		Password: "secret",
		Name:     "limousine",
		Params:   "charset=utf8mb4&parseTime=True&loc=Local",
	}
}

func postgresConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver:   DBDriverPostgres,
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "app",
		Password: "secret",
		Name:     "pakwan",
		Params:   "sslmode=prefer TimeZone=UTC",
	}
}

func TestBuildMySQLDSN(t *testing.T) {
	tests := []struct {
		name     string
		instance string
		mode     string
		want     string
	}{
		{
			name: "tcp when no instance",
			want: "app:secret@tcp(127.0.0.1:3306)/limousine?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name:     "connector by default",
			instance: "proj:asia-southeast1:mysql-dev-01",
			want:     "app:secret@cloudsql-mysql(proj:asia-southeast1:mysql-dev-01)/limousine?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name:     "socket when opted out",
			instance: "proj:asia-southeast1:mysql-dev-01",
			mode:     "socket",
			want:     "app:secret@unix(/cloudsql/proj:asia-southeast1:mysql-dev-01)/limousine?charset=utf8mb4&parseTime=True&loc=Local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DB_INSTANCE_MODE", tt.mode)

			cfg := mysqlConfig()
			cfg.Instance = tt.instance

			if got := BuildMySQLDSN(cfg); got != tt.want {
				t.Errorf("BuildMySQLDSN()\n got: %s\nwant: %s", got, tt.want)
			}
		})
	}
}

func TestBuildPostgresDSN(t *testing.T) {
	tests := []struct {
		name     string
		instance string
		mode     string
		want     string
	}{
		{
			name: "tcp when no instance",
			want: "host=127.0.0.1 user=app password=secret dbname=pakwan port=5432 sslmode=prefer TimeZone=UTC",
		},
		{
			// The connector reads the instance out of host and supplies mTLS, so
			// pgx must not negotiate SSL of its own.
			name:     "connector by default",
			instance: "proj:asia-southeast1:pg-prod-01",
			want:     "host=proj:asia-southeast1:pg-prod-01 user=app password=secret dbname=pakwan port=5432 TimeZone=UTC sslmode=disable",
		},
		{
			name:     "socket when opted out",
			instance: "proj:asia-southeast1:pg-prod-01",
			mode:     "socket",
			want:     "host=/cloudsql/proj:asia-southeast1:pg-prod-01 user=app password=secret dbname=pakwan port=5432 sslmode=prefer TimeZone=UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DB_INSTANCE_MODE", tt.mode)

			cfg := postgresConfig()
			cfg.Instance = tt.instance

			if got := BuildPostgresDSN(cfg); got != tt.want {
				t.Errorf("BuildPostgresDSN()\n got: %s\nwant: %s", got, tt.want)
			}
		})
	}
}

func TestCloudSQLDriverName(t *testing.T) {
	t.Setenv("DB_INSTANCE_MODE", "")

	mysqlCfg := mysqlConfig()
	if got := CloudSQLDriverName(mysqlCfg); got != "" {
		t.Errorf("no instance: got %q, want empty", got)
	}

	mysqlCfg.Instance = "proj:region:inst"
	if got := CloudSQLDriverName(mysqlCfg); got != CloudSQLMySQLDriver {
		t.Errorf("mysql: got %q, want %q", got, CloudSQLMySQLDriver)
	}

	pgCfg := postgresConfig()
	pgCfg.Instance = "proj:region:inst"
	if got := CloudSQLDriverName(pgCfg); got != CloudSQLPostgresDriver {
		t.Errorf("postgres: got %q, want %q", got, CloudSQLPostgresDriver)
	}

	t.Setenv("DB_INSTANCE_MODE", "socket")
	if got := CloudSQLDriverName(pgCfg); got != "" {
		t.Errorf("socket mode: got %q, want empty", got)
	}
}

func TestCreateMySqlDSNUsesConnector(t *testing.T) {
	t.Setenv("DB_INSTANCE_MODE", "")

	dsn := CreateMySqlDSN(DNS{
		Username: "app",
		Password: "secret",
		Database: "limousine",
		Instance: "proj:region:inst",
	})

	want := "app:secret@cloudsql-mysql(proj:region:inst)/limousine?charset=utf8mb4&parseTime=True"
	if dsn != want {
		t.Errorf("CreateMySqlDSN()\n got: %s\nwant: %s", dsn, want)
	}
	if !mysqlConnectorDSN(dsn) {
		t.Error("mysqlConnectorDSN() = false, want true")
	}
	if mysqlConnectorDSN("app:secret@tcp(127.0.0.1:3306)/limousine") {
		t.Error("mysqlConnectorDSN() = true for a tcp DSN")
	}
}
