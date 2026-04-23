package infra

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

const (
	DBDriverMySQL    = "mysql"
	DBDriverPostgres = "postgres"
)

type DatabaseConfig struct {
	Driver       string
	Host         string
	Port         int
	Instance     string
	User         string
	Password     string
	Name         string
	Params       string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  int // minutes
}

func LoadDatabaseConfig() DatabaseConfig {
	return loadDatabaseConfigWithPrefix("DB")
}

func LoadSecondaryDatabaseConfig() DatabaseConfig {
	return loadDatabaseConfigWithPrefix("DB2")
}

func loadDatabaseConfigWithPrefix(prefix string) DatabaseConfig {
	driver := GetEnv(prefix+"_DRIVER", DBDriverMySQL)

	defaultParams := "charset=utf8mb4&parseTime=True&loc=Local"
	if driver == DBDriverPostgres {
		defaultParams = "sslmode=disable TimeZone=UTC"
	}

	return DatabaseConfig{
		Driver:       driver,
		Host:         GetEnv(prefix+"_HOST", "127.0.0.1"),
		Port:         GetEnvInt(prefix+"_PORT", 3306),
		Instance:     GetEnv(prefix+"_INSTANCE", ""),
		User:         GetEnv(prefix+"_USER", "root"),
		Password:     GetEnv(prefix+"_PASSWORD", "root"),
		Name:         GetEnv(prefix+"_NAME", ""),
		Params:       GetEnv(prefix+"_PARAMS", defaultParams),
		MaxIdleConns: GetEnvInt(prefix+"_MAX_IDLE_CONNS", 10),
		MaxOpenConns: GetEnvInt(prefix+"_MAX_OPEN_CONNS", 25),
		MaxLifetime:  GetEnvInt(prefix+"_MAX_LIFETIME_MINUTES", 30),
	}
}

type Databases struct {
	Primary   *gorm.DB
	Secondary *gorm.DB // nil when DB2_NAME is empty
}

// BindDatabase validates the primary GORM DB and returns it with its underlying *sql.DB.
// Use this in Register to wire legacy package-level globals:
//
//	db, sqlDB, err := infra.BindDatabase(deps.Data.Databases)
//	if err != nil { return err }
//	common.Database = db
//	common.DatabaseSql = sqlDB
func BindDatabase(dbs Databases) (*gorm.DB, *sql.DB, error) {
	if dbs.Primary == nil {
		return nil, nil, fmt.Errorf("bind database: primary is nil")
	}
	sqlDB, err := dbs.Primary.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("bind database: resolve sql.DB: %w", err)
	}
	return dbs.Primary, sqlDB, nil
}

func OpenDB(cfg DatabaseConfig, appLogger *Logger) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case DBDriverPostgres:
		dialector = postgres.Open(BuildPostgresDSN(cfg))
	default:
		dialector = mysql.Open(BuildMySQLDSN(cfg))
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", cfg.Driver, err)
	}

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		return nil, fmt.Errorf("register gorm otel plugin: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("resolve sql db: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Minute)

	appLogger.Info(
		M("database connected"),
		WithComponent("database"),
		WithOperation("connect"),
		WithLogKind("infrastructure"),
		WithField("driver", cfg.Driver),
		WithField("host", cfg.Host),
		WithField("instance", cfg.Instance),
		WithField("database", cfg.Name),
	)

	return db, nil
}

func BuildMySQLDSN(cfg DatabaseConfig) string {
	if cfg.Instance != "" {
		return fmt.Sprintf(
			"%s:%s@unix(/cloudsql/%s)/%s?%s",
			cfg.User, cfg.Password, cfg.Instance, cfg.Name, cfg.Params,
		)
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.Params,
	)
}

func BuildPostgresDSN(cfg DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d %s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.Params,
	)
}
