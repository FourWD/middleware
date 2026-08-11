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
		// sslmode=prefer คือ default ของ libpq เอง: ลอง TLS ก่อน ถ้าเซิร์ฟเวอร์ไม่รองรับ
		// จึงถอยไป plaintext เดิมเป็น disable ซึ่งถูกปฏิเสธทันทีโดย Cloud SQL ที่ตั้ง
		// sslMode=ENCRYPTED_ONLY ("pg_hba.conf rejects connection ... no encryption")
		//
		// TimeZone ยังเป็น UTC โดยตั้งใจ เปลี่ยนเมื่อไรคือเลื่อนเวลาของทุก service ที่ไม่ได้
		// ตั้ง DB_PARAMS เอง service ที่ต้องการเวลาไทยให้ตั้ง DB_PARAMS ระบุมาเอง
		defaultParams = "sslmode=prefer TimeZone=UTC"
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

// OpenDialector builds the gorm dialector for cfg, registering the Cloud SQL
// connector driver when DB_INSTANCE is dialed through it.
func OpenDialector(cfg DatabaseConfig) (gorm.Dialector, error) {
	// Left empty outside connector mode so gorm keeps its own default driver.
	driverName := CloudSQLDriverName(cfg)
	if driverName != "" {
		if err := EnsureCloudSQLDriver(driverName); err != nil {
			return nil, err
		}
	}

	if cfg.Driver == DBDriverPostgres {
		return postgres.New(postgres.Config{DriverName: driverName, DSN: BuildPostgresDSN(cfg)}), nil
	}
	return mysql.New(mysql.Config{DriverName: driverName, DSN: BuildMySQLDSN(cfg)}), nil
}

func OpenDB(cfg DatabaseConfig, appLogger *Logger) (*gorm.DB, error) {
	dialector, err := OpenDialector(cfg)
	if err != nil {
		return nil, err
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

	appLogger.LifecycleEvent("DB_CONNECT_SUCCESS", map[string]any{
		"driver":    cfg.Driver,
		"host":      cfg.Host,
		"instance":  cfg.Instance,
		"database":  cfg.Name,
		"connector": UseCloudSQLConnector(cfg),
	},
		WithComponent(ComponentDB),
		WithOperation("connect"),
	)

	return db, nil
}

func BuildMySQLDSN(cfg DatabaseConfig) string {
	if cfg.Instance != "" {
		network := fmt.Sprintf("unix(/cloudsql/%s)", cfg.Instance)
		if UseCloudSQLConnector(cfg) {
			// The connector registers its dialer under the driver name, so the
			// DSN addresses the instance directly instead of a socket path.
			network = fmt.Sprintf("%s(%s)", CloudSQLMySQLDriver, cfg.Instance)
		}

		return fmt.Sprintf(
			"%s:%s@%s/%s?%s",
			cfg.User, cfg.Password, network, cfg.Name, cfg.Params,
		)
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.Params,
	)
}

func BuildPostgresDSN(cfg DatabaseConfig) string {
	// Instance ชนะ Host เหมือนฝั่ง BuildMySQLDSN — libpq/pgx ถือว่า host ที่ขึ้นต้นด้วย /
	// คือ unix socket directory ซึ่งเป็นวิธีที่ App Engine / Cloud Run ต่อ Cloud SQL
	host := cfg.Host
	params := cfg.Params
	if cfg.Instance != "" {
		host = "/cloudsql/" + cfg.Instance
		if UseCloudSQLConnector(cfg) {
			// The connector driver reads the instance connection name out of
			// host and replaces the dial address itself.
			host = cfg.Instance
			params = connectorSSLMode(params)
		}
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d %s",
		host, cfg.User, cfg.Password, cfg.Name, cfg.Port, params,
	)
}
