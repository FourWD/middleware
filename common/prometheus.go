package common

import (
	"reflect"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

var (
	httpRequestsTotal *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
)

func fiberPrometheus(c fiber.Ctx) error {
	if prometheusMiddleware == nil {
		return c.Next()
	}
	return prometheusMiddleware(c)
}

func registerPrometheus() {
	// 1. Register counters
	registerMetrics(prometheusName, "critical", CRITICAL)
	registerMetrics(prometheusName, "warning", WARNING)
	if prometheusLogic != nil {
		registerMetrics(prometheusName, "logic", prometheusLogic)
	}

	// 2. Register HTTP request metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: prometheusName + "_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: prometheusName + "_http_request_duration_seconds",
			Help: "Duration of HTTP requests in seconds",
		},
		[]string{"method", "path", "status"},
	)
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)

	// 3. Register middleware to collect request metrics
	prometheusMiddleware = func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		httpRequestsTotal.WithLabelValues(c.Method(), c.Path(), status).Inc()
		httpRequestDuration.WithLabelValues(c.Method(), c.Path(), status).Observe(duration)
		return err
	}
	fiberApp.Use(fiberPrometheus)

	// 4. Register /metrics endpoint
	fiberApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
}
