package infra

import (
	"testing"
)

func TestValidateCommonConfig_OK(t *testing.T) {
	cfg := baseCommonConfig()
	if err := validateCommonConfig(cfg); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestValidateCommonConfig_InvalidAppEnv(t *testing.T) {
	cfg := baseCommonConfig()
	cfg.AppEnv = "production"

	if err := validateCommonConfig(cfg); err == nil {
		t.Fatal("expected error for invalid APP_ENV")
	}
}

func TestValidateCommonConfig_DebugAuthTokenTooShort(t *testing.T) {
	cfg := baseCommonConfig()
	cfg.DebugAuthToken = "short-token"

	if err := validateCommonConfig(cfg); err == nil {
		t.Fatal("expected error for short HTTP_DEBUG_AUTH_TOKEN")
	}
}

func TestValidateCommonConfig_InvalidRedisAddress(t *testing.T) {
	cfg := baseCommonConfig()
	cfg.Redis.Addr = ""

	if err := validateCommonConfig(cfg); err == nil {
		t.Fatal("expected error for invalid REDIS_ADDR")
	}
}

func TestValidateCommonConfig_InvalidRedisDB(t *testing.T) {
	cfg := baseCommonConfig()
	cfg.Redis.DB = -1

	if err := validateCommonConfig(cfg); err == nil {
		t.Fatal("expected error for invalid REDIS_DB")
	}
}

func TestValidateCommonConfig_InvalidPrimaryDatabaseDriver(t *testing.T) {
	cfg := baseCommonConfig()
	cfg.Database.Driver = "sqlite"

	if err := validateCommonConfig(cfg); err == nil {
		t.Fatal("expected error for invalid DB_DRIVER")
	}
}

func TestValidateCommonConfig_InvalidSecondaryDatabaseDriver(t *testing.T) {
	cfg := baseCommonConfig()
	cfg.SecondaryDBEnabled = true
	cfg.SecondaryDatabase.Driver = "sqlite"

	if err := validateCommonConfig(cfg); err == nil {
		t.Fatal("expected error for invalid DB2_DRIVER")
	}
}

func TestValidateCommonConfig_RefreshTTLTooShort(t *testing.T) {
	cfg := baseCommonConfig()
	cfg.Auth.AccessTokenTTLMinutes = 120
	cfg.Auth.RefreshTokenTTLMinutes = 60

	if err := validateCommonConfig(cfg); err == nil {
		t.Fatal("expected error for refresh ttl less than access ttl")
	}
}

func TestValidateCommonConfig_BootstrapAdminEmailWithoutPassword(t *testing.T) {
	cfg := baseCommonConfig()
	cfg.Auth.BootstrapAdmin.Email = "admin@example.com"
	cfg.Auth.BootstrapAdmin.Password = ""

	if err := validateCommonConfig(cfg); err == nil {
		t.Fatal("expected error when email set without password")
	}
}

func TestValidateCommonConfig_BootstrapAdminPasswordWithoutEmail(t *testing.T) {
	cfg := baseCommonConfig()
	cfg.Auth.BootstrapAdmin.Email = ""
	cfg.Auth.BootstrapAdmin.Password = "secret123"

	if err := validateCommonConfig(cfg); err == nil {
		t.Fatal("expected error when password set without email")
	}
}

func baseCommonConfig() CommonConfig {
	return CommonConfig{
		AppName:      "pakkad-service",
		AppEnv:       "dev",
		HTTPAddress:  ":8080",
		RedisEnabled: true,
		Database: DatabaseConfig{
			Driver:       DBDriverMySQL,
			Host:         "127.0.0.1",
			Port:         3306,
			User:         "root",
			Name:         "db",
			MaxIdleConns: 10,
			MaxOpenConns: 25,
			MaxLifetime:  30,
		},
		SecondaryDatabase: DatabaseConfig{
			Driver:       DBDriverPostgres,
			Host:         "127.0.0.1",
			Port:         5432,
			User:         "postgres",
			Name:         "db2",
			MaxIdleConns: 10,
			MaxOpenConns: 25,
			MaxLifetime:  30,
		},
		Redis: RedisConfig{
			Addr: "127.0.0.1:6379",
			DB:   0,
		},
		Auth: AuthConfig{
			JWTSecret:              "secret",
			JWTIssuer:              "issuer",
			AccessTokenTTLMinutes:  60,
			RefreshTokenTTLMinutes: 10080,
			BcryptCost:             12,
		},
	}
}
