package infra

import (
	"fmt"
	"strings"
	"sync"

	"cloud.google.com/go/cloudsqlconn"
	connmysql "cloud.google.com/go/cloudsqlconn/mysql/mysql"
	connpgx "cloud.google.com/go/cloudsqlconn/postgres/pgxv5"
)

// database/sql driver names registered for Cloud SQL connector dialing. The
// MySQL one doubles as the DSN network, e.g. user:pass@cloudsql-mysql(inst)/db.
const (
	CloudSQLMySQLDriver    = "cloudsql-mysql"
	CloudSQLPostgresDriver = "cloudsql-postgres"
)

const instanceModeSocket = "socket"

// UseCloudSQLConnector reports whether DB_INSTANCE should be dialed through the
// Cloud SQL Go connector instead of the /cloudsql unix socket. The socket only
// exists when the platform mounts it — on App Engine that needs a deploy-time
// `beta_settings: cloud_sql_instances` directive, which cannot come from .env.
// The connector dials the Cloud SQL Admin API itself, so DB_INSTANCE alone is
// enough; DB_INSTANCE_MODE=socket opts back out.
//
// Requires the Cloud SQL Admin API enabled on the running project and
// roles/cloudsql.client for the service account on the instance's project.
func UseCloudSQLConnector(cfg DatabaseConfig) bool {
	if cfg.Instance == "" {
		return false
	}
	return !strings.EqualFold(GetEnv("DB_INSTANCE_MODE", "connector"), instanceModeSocket)
}

// CloudSQLDriverName returns the connector driver for cfg, or "" when cfg is
// dialed over tcp or the unix socket.
func CloudSQLDriverName(cfg DatabaseConfig) string {
	if !UseCloudSQLConnector(cfg) {
		return ""
	}
	if cfg.Driver == DBDriverPostgres {
		return CloudSQLPostgresDriver
	}
	return CloudSQLMySQLDriver
}

var (
	cloudSQLMu      sync.Mutex
	cloudSQLDrivers = map[string]error{}
)

// EnsureCloudSQLDriver registers the connector-backed driver once per process.
// sql.Register panics on a duplicate name, so the outcome — error included — is
// cached and replayed to later callers.
func EnsureCloudSQLDriver(name string) error {
	cloudSQLMu.Lock()
	defer cloudSQLMu.Unlock()

	if err, ok := cloudSQLDrivers[name]; ok {
		return err
	}

	// Lazy refresh keeps no background goroutine alive between requests, which
	// is what App Engine / Cloud Run need on instances that get frozen.
	opts := []cloudsqlconn.Option{cloudsqlconn.WithLazyRefresh()}

	var err error
	switch name {
	case CloudSQLMySQLDriver:
		// The dialer lives for the life of the process; its cleanup is dropped
		// because nothing outlives it to call.
		_, err = connmysql.RegisterDriver(name, opts...)
	case CloudSQLPostgresDriver:
		_, err = connpgx.RegisterDriver(name, opts...)
	default:
		err = fmt.Errorf("unknown driver %q", name)
	}
	if err != nil {
		err = fmt.Errorf("register cloud sql driver %q: %w", name, err)
	}

	cloudSQLDrivers[name] = err
	return err
}

// SQLDriverDSN returns the database/sql driver name and DSN for cfg, registering
// the connector driver first when the instance is dialed through it.
func SQLDriverDSN(cfg DatabaseConfig) (driver, dsn string, err error) {
	driver = CloudSQLDriverName(cfg)
	if driver != "" {
		if err := EnsureCloudSQLDriver(driver); err != nil {
			return "", "", err
		}
	}

	if cfg.Driver == DBDriverPostgres {
		if driver == "" {
			driver = "pgx"
		}
		return driver, BuildPostgresDSN(cfg), nil
	}

	if driver == "" {
		driver = DBDriverMySQL
	}
	return driver, BuildMySQLDSN(cfg), nil
}

// connectorSSLMode forces sslmode=disable on a Postgres DSN. The connector
// already wraps the connection in mTLS, so pgx negotiating its own TLS on top
// would be a second, pointless handshake.
func connectorSSLMode(params string) string {
	kept := make([]string, 0, strings.Count(params, "=")+1)
	for _, field := range strings.Fields(params) {
		if strings.HasPrefix(strings.ToLower(field), "sslmode=") {
			continue
		}
		kept = append(kept, field)
	}
	return strings.Join(append(kept, "sslmode=disable"), " ")
}

// mysqlConnectorDSN reports whether a raw MySQL DSN was built for the connector.
// The legacy Connect helpers take a DSN string rather than a DatabaseConfig, so
// the network name is the only signal they have.
func mysqlConnectorDSN(dsn string) bool {
	return strings.Contains(dsn, "@"+CloudSQLMySQLDriver+"(")
}
