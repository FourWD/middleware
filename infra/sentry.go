package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel/trace"
)

// SentryConfig holds Sentry initialization configuration.
// Use LoadSentryConfig() to populate from environment variables.
type SentryConfig struct {
	Enabled          bool
	DSN              string
	Environment      string
	Release          string
	TracesSampleRate float64
}

// LoadSentryConfig reads Sentry configuration from environment variables.
func LoadSentryConfig() SentryConfig {
	appName := GetEnv("APP_NAME", "app")
	appVersion := GetEnv("APP_VERSION", "0.1.0")

	return SentryConfig{
		Enabled:          GetEnvBool("SENTRY_ENABLED", false),
		DSN:              GetEnv("SENTRY_DSN", ""),
		Environment:      GetEnv("APP_ENV", "local"),
		Release:          appName + "@" + appVersion,
		TracesSampleRate: GetEnvFloat("SENTRY_TRACES_SAMPLE_RATE", 0),
	}
}

// SetupSentry initializes the Sentry SDK.
// Returns a shutdown function that flushes buffered events.
// Returns a no-op shutdown if Sentry is disabled.
func SetupSentry(cfg SentryConfig) (func(context.Context) error, error) {
	if !cfg.Enabled || cfg.DSN == "" {
		return func(context.Context) error { return nil }, nil
	}

	cfg.TracesSampleRate = normalizeSampleRate(cfg.TracesSampleRate)

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		EnableTracing:    cfg.TracesSampleRate > 0,
		TracesSampleRate: cfg.TracesSampleRate,
	})
	if err != nil {
		return nil, err
	}

	return func(context.Context) error {
		sentry.Flush(2 * time.Second)
		return nil
	}, nil
}

func normalizeSampleRate(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func registerSentryMiddleware(app *fiber.App, cfg StackConfig) {
	if !cfg.SentryEnabled {
		return
	}

	app.Use(func(c fiber.Ctx) error {
		defer func() {
			if recovered := recover(); recovered != nil {
				sentry.WithScope(func(scope *sentry.Scope) {
					enrichSentryScope(scope, c)
					scope.SetLevel(sentry.LevelFatal)
					sentry.CaptureException(panicToError(recovered))
				})
				panic(recovered)
			}
		}()

		err := c.Next()
		if err != nil {
			sentry.WithScope(func(scope *sentry.Scope) {
				enrichSentryScope(scope, c)
				scope.SetExtra("http.status_code", c.Response().StatusCode())
				sentry.CaptureException(err)
			})
		}

		return err
	})
}

func panicToError(value any) error {
	if err, ok := value.(error); ok {
		return err
	}
	return fmt.Errorf("panic recovered: %v", value)
}

func enrichSentryScope(scope *sentry.Scope, c fiber.Ctx) {
	scope.SetTag("request_id", GetRequestID(c))
	scope.SetTag("http.route", routePath(c))
	scope.SetTag("http.method", c.Method())

	spanCtx := trace.SpanContextFromContext(c.Context())
	if spanCtx.IsValid() {
		scope.SetTag("trace_id", spanCtx.TraceID().String())
		scope.SetTag("span_id", spanCtx.SpanID().String())
	}
}
