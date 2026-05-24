package infra

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
// Environment is normalized the same way as CommonConfig.AppEnv (lowercased,
// trimmed) so Sentry's environment tag matches the app's canonical env value
// regardless of LoadSentryConfig vs NewApp initialization order.
//
// Release identifier: if APP_VERSION is unset, the release tag rotates daily
// as "dev-YYYYMMDD" so missing-version configs are visible in Sentry release
// health rather than silently grouping every dev build under "0.1.0".
func LoadSentryConfig() SentryConfig {
	appID := GetEnv("APP_ID", "app")
	appVersion := GetEnv("APP_VERSION", "")
	if appVersion == "" {
		appVersion = "dev-" + time.Now().UTC().Format("20060102")
	}

	return SentryConfig{
		Enabled:          GetEnvBool("SENTRY_ENABLED", false),
		DSN:              GetEnv("SENTRY_DSN", ""),
		Environment:      strings.ToLower(strings.TrimSpace(GetEnv("APP_ENV", "local"))),
		Release:          appID + "@" + appVersion,
		TracesSampleRate: GetEnvFloat("SENTRY_TRACES_SAMPLE_RATE", 0),
	}
}

// noopShutdown is a safe shutdown function used when Sentry is disabled or
// init fails — callers can invoke it unconditionally without nil-checking.
func noopShutdown(context.Context) error { return nil }

// SetupSentry initializes the Sentry SDK and returns a shutdown function
// that flushes buffered events. The shutdown respects the caller's ctx
// deadline (falls back to 2s when no deadline is set). Both the disabled
// case and the init-error case return a safe no-op shutdown so the caller
// never has to nil-check.
func SetupSentry(cfg SentryConfig) (func(context.Context) error, error) {
	if !cfg.Enabled || cfg.DSN == "" {
		return noopShutdown, nil
	}

	cfg.TracesSampleRate = normalizeSampleRate(cfg.TracesSampleRate)

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		EnableTracing:    cfg.TracesSampleRate > 0,
		TracesSampleRate: cfg.TracesSampleRate,
	}); err != nil {
		return noopShutdown, err
	}

	return func(ctx context.Context) error {
		timeout := 2 * time.Second
		if deadline, ok := ctx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining < 0 {
				remaining = 0
			}
			timeout = remaining
		}
		sentry.Flush(timeout)
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
		// Each request gets its own Hub clone so concurrent requests don't
		// share scope state (tags, user, breadcrumbs). The local hub is
		// passed to all capture calls in this middleware rather than the
		// global sentry.* package functions.
		hub := sentry.CurrentHub().Clone()

		defer func() {
			if recovered := recover(); recovered != nil {
				safelyCaptureToSentry(func() {
					hub.WithScope(func(scope *sentry.Scope) {
						enrichSentryScope(scope, c)
						scope.SetLevel(sentry.LevelFatal)
						hub.CaptureException(panicToError(recovered))
					})
				})
				panic(recovered)
			}
		}()

		err := c.Next()
		if err != nil {
			safelyCaptureToSentry(func() {
				hub.WithScope(func(scope *sentry.Scope) {
					enrichSentryScope(scope, c)
					scope.SetTag("http.status_code", strconv.Itoa(c.Response().StatusCode()))
					hub.CaptureException(err)
				})
			})
		}

		return err
	})
}

// safelyCaptureToSentry runs the given Sentry instrumentation under its own
// recover. If anything in the Sentry path panics — enrichment, scope setup,
// the SDK itself — the panic is swallowed here rather than masking the
// original error or panic value that the caller is trying to record. The
// swallowed panic is logged as a lifecycle warning so broken Sentry
// instrumentation is visible in the app's own logs.
func safelyCaptureToSentry(fn func()) {
	defer func() {
		r := recover()
		if r == nil || AppLog == nil {
			return
		}
		AppLog.LifecycleWarn("SENTRY_INSTRUMENTATION_PANIC",
			map[string]any{"panic": fmt.Sprintf("%v", r)},
			WithComponent(ComponentApp),
			WithOperation("sentry_capture"))
	}()
	fn()
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

	// Peek session claims directly instead of GetSession() to avoid the
	// SESSION_INVALID_SIGNATURE warning side-effect on public/unauth paths.
	if claims := fiber.Locals[*JWTClaims](c, "user"); claims != nil && claims.UserID != "" {
		// Only ID is set — name/email are PII per CLAUDE.md deny-list.
		scope.SetUser(sentry.User{ID: claims.UserID})
	}

	spanCtx := trace.SpanContextFromContext(c.Context())
	if spanCtx.IsValid() {
		scope.SetTag("trace_id", spanCtx.TraceID().String())
		scope.SetTag("span_id", spanCtx.SpanID().String())
	}
}
