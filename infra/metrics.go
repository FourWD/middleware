package infra

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var (
	metricsOnce                       sync.Once
	httpRequestsTotal                 *prometheus.CounterVec
	httpRequestDuration               *prometheus.HistogramVec
	httpInFlight                      *prometheus.GaugeVec
	httpServerErrors                  *prometheus.CounterVec
	backgroundHeartbeatRunsTotal      *prometheus.CounterVec
	backgroundHeartbeatRetriesTotal   prometheus.Counter
	backgroundCircuitBreakerOpenTotal prometheus.Counter
	metricsNamespace                  string
)

func initPrometheus(namespace string) {
	metricsOnce.Do(func() {
		metricsNamespace = namespace

		httpRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests.",
			},
			[]string{"method", "route", "status"},
		)

		httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "route", "status"},
		)

		httpInFlight = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "in_flight_requests",
				Help:      "Current number of in-flight HTTP requests.",
			},
			[]string{"method", "route"},
		)

		httpServerErrors = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "server_errors_total",
				Help:      "Total number of HTTP 5xx responses.",
			},
			[]string{"method", "route", "status"},
		)

		backgroundHeartbeatRunsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "background",
				Name:      "heartbeat_runs_total",
				Help:      "Total number of heartbeat runs by result.",
			},
			[]string{"result"},
		)

		backgroundHeartbeatRetriesTotal = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "background",
				Name:      "heartbeat_retries_total",
				Help:      "Total number of heartbeat retry attempts.",
			},
		)

		backgroundCircuitBreakerOpenTotal = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "background",
				Name:      "circuit_breaker_open_total",
				Help:      "Total number of times the heartbeat circuit breaker reached the open state.",
			},
		)

		prometheus.MustRegister(
			httpRequestsTotal,
			httpRequestDuration,
			httpInFlight,
			httpServerErrors,
			backgroundHeartbeatRunsTotal,
			backgroundHeartbeatRetriesTotal,
			backgroundCircuitBreakerOpenTotal,
		)
	})
}

func incrementHeartbeatRetryMetric() {
	if backgroundHeartbeatRetriesTotal != nil {
		backgroundHeartbeatRetriesTotal.Inc()
	}
}

func incrementHeartbeatRunMetric(result string) {
	if backgroundHeartbeatRunsTotal != nil {
		backgroundHeartbeatRunsTotal.WithLabelValues(result).Inc()
	}
}

func incrementBackgroundCircuitBreakerOpenMetric() {
	if backgroundCircuitBreakerOpenTotal != nil {
		backgroundCircuitBreakerOpenTotal.Inc()
	}
}

func registerMetrics(app *fiber.App, cfg StackConfig) {
	initPrometheus(cfg.MetricsNamespace)

	meter := otel.Meter(cfg.ServiceName + "/http")
	var logger *Logger = cfg.Logger

	requests, err := meter.Int64Counter("http.server.requests")
	if err != nil {
		if logger != nil {
			logger.Error(err, M("failed to create http.server.requests counter"), WithComponent("metrics"), WithOperation("create_request_counter"), WithLogKind("infrastructure"))
		}
		requests = noop.Int64Counter{}
	}
	duration, err := meter.Float64Histogram("http.server.duration.ms")
	if err != nil {
		if logger != nil {
			logger.Error(err, M("failed to create http.server.duration.ms histogram"), WithComponent("metrics"), WithOperation("create_duration_histogram"), WithLogKind("infrastructure"))
		}
		duration = noop.Float64Histogram{}
	}

	app.Use(func(c fiber.Ctx) error {
		route := routePath(c)
		httpInFlight.WithLabelValues(c.Method(), route).Inc()
		defer httpInFlight.WithLabelValues(c.Method(), route).Dec()

		startedAt := time.Now()
		err := c.Next()
		elapsed := time.Since(startedAt)
		status := c.Response().StatusCode()
		statusClass := statusCodeClass(status)

		attrs := metric.WithAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.route", route),
			attribute.Int("http.status_code", status),
			attribute.String("http.status_class", statusClass),
		)
		requests.Add(c.Context(), 1, attrs)
		duration.Record(c.Context(), float64(elapsed.Milliseconds()), attrs)

		statusStr := strconv.Itoa(status)
		httpRequestsTotal.WithLabelValues(c.Method(), route, statusStr).Inc()
		httpRequestDuration.WithLabelValues(c.Method(), route, statusStr).Observe(elapsed.Seconds())
		if status >= fiber.StatusInternalServerError {
			httpServerErrors.WithLabelValues(c.Method(), route, statusStr).Inc()
		}

		return err
	})
}

// PrometheusHandler returns an http.Handler for the /metrics endpoint.
// Call this after RegisterStack() so that prometheus collectors are initialized.
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
