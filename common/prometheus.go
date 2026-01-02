package common

import (
	"reflect"

	fiberprometheus "github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var Prometheus = map[string]prometheus.Counter{}
var prometheusMiddleware fiber.Handler

// ===== Critical Errors =====
type criticalTypes struct {
	// Database
	DB_CONNECTION string
	DB_QUERY      string
	DB_TIMEOUT    string

	// Firebase
	FIREBASE_GET    string
	FIREBASE_SET    string
	FIREBASE_UPDATE string
	FIREBASE_DELETE string

	// Redis
	REDIS_CONNECTION string
	REDIS_GET        string
	REDIS_SET        string

	// External API
	API_CONNECTION string
	API_TIMEOUT    string
}

var CRITICAL = criticalTypes{
	DB_CONNECTION:    "db_connection",
	DB_QUERY:         "db_query",
	DB_TIMEOUT:       "db_timeout",
	FIREBASE_GET:     "firebase_get",
	FIREBASE_SET:     "firebase_set",
	FIREBASE_UPDATE:  "firebase_update",
	FIREBASE_DELETE:  "firebase_delete",
	REDIS_CONNECTION: "redis_connection",
	REDIS_GET:        "redis_get",
	REDIS_SET:        "redis_set",
	API_CONNECTION:   "api_connection",
	API_TIMEOUT:      "api_timeout",
}

// ===== Warning Errors =====
type warningTypes struct {
	// Performance
	SLOW_QUERY      string
	SLOW_CONNECTION string
	SLOW_RESPONSE   string

	// Resource
	HIGH_MEMORY     string
	HIGH_CPU        string
	HIGH_CONNECTION string

	// Queue
	QUEUE_HIGH    string
	QUEUE_TIMEOUT string

	// Rate Limit
	RATE_LIMIT_NEAR string
}

var WARNING = warningTypes{
	SLOW_QUERY:      "slow_query",
	SLOW_CONNECTION: "slow_connection",
	SLOW_RESPONSE:   "slow_response",
	HIGH_MEMORY:     "high_memory",
	HIGH_CPU:        "high_cpu",
	HIGH_CONNECTION: "high_connection",
	QUEUE_HIGH:      "queue_high",
	QUEUE_TIMEOUT:   "queue_timeout",
	RATE_LIMIT_NEAR: "rate_limit_near",
}

func registerMetrics(prefix string, category string, data interface{}) {
	v := reflect.ValueOf(data)
	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i).String()
		Prometheus[val] = prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_" + category + "_" + val + "_total",
			Help: category + " error: " + val,
		})
		prometheus.MustRegister(Prometheus[val])
	}
}

func fiberPrometheus(c *fiber.Ctx) error {
	if prometheusMiddleware == nil {
		return c.Next()
	}
	return prometheusMiddleware(c)
}

func RegisterPrometheus(app *fiber.App, name string, logic interface{}) {
	p := fiberprometheus.New(name)
	p.RegisterAt(app, "/metrics")
	prometheusMiddleware = p.Middleware
	app.Use(fiberPrometheus)

	registerMetrics(name, "critical", CRITICAL)
	registerMetrics(name, "warning", WARNING)
	registerMetrics(name, "logic", logic)
}
