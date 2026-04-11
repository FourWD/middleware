package infra

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	recovermw "github.com/gofiber/fiber/v3/middleware/recover"
)

// StackConfig configures the standard middleware stack.
// Use LoadStackConfig() to populate from environment variables,
// then set Logger and RateLimitStore before calling RegisterStack().
type StackConfig struct {
	// ServiceName used for otel tracer and metrics namespace.
	ServiceName string

	// Logger for request logging. If nil, request logging is skipped.
	Logger *Logger

	// Request log
	RequestLogOmitRequestBody  bool
	RequestLogOmitResponseBody bool
	RequestLogMaxBodyBytes     int

	// CORS
	AllowOrigins string

	// Rate limit
	RateLimitEnabled       bool
	RateLimitMaxRequests   int
	RateLimitWindowSeconds int
	RateLimitKeyPrefix     string
	RateLimitExemptHealth  bool
	RateLimitStore         RateLimitStore

	// Sentry
	SentryEnabled bool

	// Metrics
	MetricsNamespace string
}

// LoadStackConfig reads standard middleware configuration from environment variables.
// After calling this, set Logger and RateLimitStore manually as they are project-specific.
func LoadStackConfig() StackConfig {
	return StackConfig{
		ServiceName:                GetEnv("APP_NAME", "app"),
		RequestLogOmitRequestBody:  GetEnvBool("HTTP_REQUEST_LOG_OMIT_REQUEST_BODY", true),
		RequestLogOmitResponseBody: GetEnvBool("HTTP_REQUEST_LOG_OMIT_RESPONSE_BODY", true),
		RequestLogMaxBodyBytes:     GetEnvInt("HTTP_REQUEST_LOG_MAX_BODY_BYTES", 4096),
		AllowOrigins:               GetEnv("HTTP_ALLOW_ORIGINS", "*"),
		RateLimitEnabled:           GetEnvBool("RATE_LIMIT_ENABLED", true),
		RateLimitMaxRequests:       GetEnvInt("RATE_LIMIT_MAX_REQUESTS", 60),
		RateLimitWindowSeconds:     GetEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60),
		RateLimitKeyPrefix:         GetEnv("RATE_LIMIT_KEY_PREFIX", "rate_limit"),
		RateLimitExemptHealth:      GetEnvBool("RATE_LIMIT_EXEMPT_HEALTH", true),
		SentryEnabled:              GetEnvBool("SENTRY_ENABLED", false),
		MetricsNamespace:           GetEnv("METRICS_NAMESPACE", "app"),
	}
}

func (c StackConfig) normalized() StackConfig {
	cfg := c
	if strings.TrimSpace(cfg.ServiceName) == "" {
		cfg.ServiceName = "app"
	}
	if strings.TrimSpace(cfg.AllowOrigins) == "" {
		cfg.AllowOrigins = "*"
	}
	if strings.TrimSpace(cfg.MetricsNamespace) == "" {
		cfg.MetricsNamespace = "app"
	}
	if cfg.RequestLogMaxBodyBytes <= 0 {
		cfg.RequestLogMaxBodyBytes = 4096
	}
	if cfg.RateLimitMaxRequests <= 0 {
		cfg.RateLimitMaxRequests = 60
	}
	if cfg.RateLimitWindowSeconds <= 0 {
		cfg.RateLimitWindowSeconds = 60
	}
	if strings.TrimSpace(cfg.RateLimitKeyPrefix) == "" {
		cfg.RateLimitKeyPrefix = "rate_limit"
	}

	return cfg
}

func validateStackConfig(cfg StackConfig) error {
	if cfg.RateLimitEnabled && cfg.RateLimitStore == nil {
		return fmt.Errorf("rate limit enabled but store is nil")
	}

	return nil
}

// RegisterStack registers the full middleware stack in the correct order:
// RequestID → CORS → Sentry → Recover → Envelope → RateLimit → OTel → Metrics → RequestLog
//
// For projects with WebSocket or SSE routes, use RegisterBaseStack + RegisterHTTPStack
// with realtime routes registered in between to exempt them from rate limiting and tracing.
func RegisterStack(app *fiber.App, cfg StackConfig) {
	RegisterBaseStack(app, cfg)
	RegisterHTTPStack(app, cfg)
}

// RegisterBaseStack registers essential middleware that applies to all routes including
// long-lived connections (WebSocket, SSE): RequestID → CORS → Sentry → Recover.
// Call this before registering WebSocket/SSE routes.
func RegisterBaseStack(app *fiber.App, cfg StackConfig) {
	cfg = cfg.normalized()
	app.Use(NewRequestID())
	registerCORS(app, cfg)
	registerSentryMiddleware(app, cfg)
	app.Use(recovermw.New())
}

// RegisterHTTPStack registers HTTP-only middleware: RateLimit → OTel → Metrics → RequestLog.
// Call this after registering WebSocket/SSE routes and before registering REST routes.
// Note: Sentry is registered in RegisterBaseStack (before Recover) so it captures panics with full stack traces.
func RegisterHTTPStack(app *fiber.App, cfg StackConfig) {
	cfg = cfg.normalized()

	if cfg.AllowOrigins == "*" && GetEnv("APP_ENV", "local") == "prod" && cfg.Logger != nil {
		cfg.Logger.Warn(
			M("CORS AllowOrigins is wildcard (*) in production — set HTTP_ALLOW_ORIGINS explicitly"),
			WithComponent("middleware"),
			WithOperation("cors_check"),
			WithLogKind("security"),
		)
	}

	if err := validateStackConfig(cfg); err != nil && cfg.Logger != nil {
		cfg.Logger.Warn(
			M("invalid middleware stack config, disabling rate limit"),
			WithComponent("middleware"),
			WithOperation("rate_limit_init"),
			WithLogKind("configuration"),
			WithField("error", err),
		)
		cfg.RateLimitEnabled = false
	}

	app.Use(NewEnvelopeWrapper())
	registerRateLimit(app, cfg)
	registerOTelTrace(app, cfg)
	registerMetrics(app, cfg)
	if cfg.Logger != nil {
		app.Use(NewRequestLog(RequestLogConfig{
			RequestLogger:    NewSlogRequestLogger(cfg.Logger),
			OmitRequestBody:  cfg.RequestLogOmitRequestBody,
			OmitResponseBody: cfg.RequestLogOmitResponseBody,
			MaxBodyBytes:     cfg.RequestLogMaxBodyBytes,
		}))
	}
}

func registerCORS(app *fiber.App, cfg StackConfig) {
	app.Use(cors.New(cors.Config{
		AllowOrigins:  splitOrigins(cfg.AllowOrigins),
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization", RequestIDHeader},
		ExposeHeaders: []string{RequestIDHeader},
	}))
}

func splitOrigins(origins string) []string {
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
