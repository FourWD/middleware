package infra

import (
	"database/sql"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// dbStatsCollector implements prometheus.Collector and reads sql.DBStats on every scrape.
// Scope and naming follow the de-facto Go convention (go_sql_*), compatible with common
// Grafana dashboards (e.g. search: "golang db stats").
type dbStatsCollector struct {
	db   *sql.DB
	name string

	maxOpen        *prometheus.Desc
	open           *prometheus.Desc
	inUse          *prometheus.Desc
	idle           *prometheus.Desc
	waitCount      *prometheus.Desc
	waitDuration   *prometheus.Desc
	idleClosed     *prometheus.Desc
	idleTimeClosed *prometheus.Desc
	lifetimeClosed *prometheus.Desc
}

func newDBStatsCollector(name string, db *sql.DB) *dbStatsCollector {
	labels := prometheus.Labels{"db_name": name}
	desc := func(suffix, help string) *prometheus.Desc {
		return prometheus.NewDesc("go_sql_"+suffix, help, nil, labels)
	}
	return &dbStatsCollector{
		db:             db,
		name:           name,
		maxOpen:        desc("max_open_connections", "Maximum number of open connections to the database."),
		open:           desc("open_connections", "Number of established connections both in use and idle."),
		inUse:          desc("in_use", "Number of connections currently in use."),
		idle:           desc("idle", "Number of idle connections."),
		waitCount:      desc("wait_count_total", "Total number of connections waited for."),
		waitDuration:   desc("wait_duration_seconds_total", "Total time blocked waiting for a new connection."),
		idleClosed:     desc("max_idle_closed_total", "Total number of connections closed due to SetMaxIdleConns."),
		idleTimeClosed: desc("max_idle_time_closed_total", "Total number of connections closed due to SetConnMaxIdleTime."),
		lifetimeClosed: desc("max_lifetime_closed_total", "Total number of connections closed due to SetConnMaxLifetime."),
	}
}

func (c *dbStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxOpen
	ch <- c.open
	ch <- c.inUse
	ch <- c.idle
	ch <- c.waitCount
	ch <- c.waitDuration
	ch <- c.idleClosed
	ch <- c.idleTimeClosed
	ch <- c.lifetimeClosed
}

func (c *dbStatsCollector) Collect(ch chan<- prometheus.Metric) {
	s := c.db.Stats()
	ch <- prometheus.MustNewConstMetric(c.maxOpen, prometheus.GaugeValue, float64(s.MaxOpenConnections))
	ch <- prometheus.MustNewConstMetric(c.open, prometheus.GaugeValue, float64(s.OpenConnections))
	ch <- prometheus.MustNewConstMetric(c.inUse, prometheus.GaugeValue, float64(s.InUse))
	ch <- prometheus.MustNewConstMetric(c.idle, prometheus.GaugeValue, float64(s.Idle))
	ch <- prometheus.MustNewConstMetric(c.waitCount, prometheus.CounterValue, float64(s.WaitCount))
	ch <- prometheus.MustNewConstMetric(c.waitDuration, prometheus.CounterValue, s.WaitDuration.Seconds())
	ch <- prometheus.MustNewConstMetric(c.idleClosed, prometheus.CounterValue, float64(s.MaxIdleClosed))
	ch <- prometheus.MustNewConstMetric(c.idleTimeClosed, prometheus.CounterValue, float64(s.MaxIdleTimeClosed))
	ch <- prometheus.MustNewConstMetric(c.lifetimeClosed, prometheus.CounterValue, float64(s.MaxLifetimeClosed))
}

// registerDBMetrics registers Prometheus collectors for each configured GORM database.
// Metrics are scraped on-demand from sql.DBStats — no goroutine required.
func registerDBMetrics(dbs Databases, logger *Logger) {
	if dbs.Primary != nil {
		if err := registerOne("primary", dbs.Primary, logger); err != nil && logger != nil {
			logger.Warn(M("register db metrics failed"),
				WithField("db_name", "primary"),
				WithField("error", err.Error()),
				WithComponent("metrics"), WithOperation("register_db_metrics"), WithLogKind("startup"))
		}
	}
	if dbs.Secondary != nil {
		if err := registerOne("secondary", dbs.Secondary, logger); err != nil && logger != nil {
			logger.Warn(M("register db metrics failed"),
				WithField("db_name", "secondary"),
				WithField("error", err.Error()),
				WithComponent("metrics"), WithOperation("register_db_metrics"), WithLogKind("startup"))
		}
	}
}

func registerOne(name string, gormDB interface{ DB() (*sql.DB, error) }, logger *Logger) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("resolve sql.DB: %w", err)
	}
	collector := newDBStatsCollector(name, sqlDB)
	if err := prometheus.Register(collector); err != nil {
		// AlreadyRegisteredError is not fatal — likely a re-init in tests.
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return nil
		}
		return err
	}
	return nil
}
