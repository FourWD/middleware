package infra

import "testing"

func TestLoadDatabaseConfigDefaultParams(t *testing.T) {
	cases := []struct {
		name   string
		driver string
		want   string
	}{
		{"mysql", DBDriverMySQL, "charset=utf8mb4&parseTime=True&loc=Local"},
		{"postgres", DBDriverPostgres, "sslmode=prefer TimeZone=Asia/Bangkok"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DB_DRIVER", tc.driver)
			if got := LoadDatabaseConfig().Params; got != tc.want {
				t.Fatalf("Params = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLoadDatabaseConfigParamsOverride(t *testing.T) {
	t.Setenv("DB_DRIVER", DBDriverPostgres)
	t.Setenv("DB_PARAMS", "sslmode=require TimeZone=UTC")

	if got := LoadDatabaseConfig().Params; got != "sslmode=require TimeZone=UTC" {
		t.Fatalf("DB_PARAMS override ignored, got %q", got)
	}
}

// The connector path strips sslmode but must keep TimeZone — it is the only
// place the Thai business-day timezone reaches Cloud SQL.
func TestBuildPostgresDSNConnectorKeepsTimeZone(t *testing.T) {
	t.Setenv("DB_INSTANCE_MODE", "connector")

	dsn := BuildPostgresDSN(DatabaseConfig{
		Instance: "proj:asia-southeast1:pg-prod-01",
		User:     "app",
		Password: "secret",
		Name:     "pakwan",
		Port:     5432,
		Params:   "sslmode=prefer TimeZone=Asia/Bangkok",
	})

	want := "host=proj:asia-southeast1:pg-prod-01 user=app password=secret dbname=pakwan port=5432 TimeZone=Asia/Bangkok sslmode=disable"
	if dsn != want {
		t.Fatalf("BuildPostgresDSN()\n got: %s\nwant: %s", dsn, want)
	}
}
