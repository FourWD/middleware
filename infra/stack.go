package infra

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	recovermw "github.com/gofiber/fiber/v3/middleware/recover"
)

// StackConfig configures the standard middleware stack.
// Use LoadStackConfig() to populate from environment variables,
// then set Logger before calling RegisterStack().
//
// Rate limiting is NOT part of this stack — use AppDeps.Runtime.RateLimit
// tiered middleware (Strict/Default) applied to route groups instead.
type StackConfig struct {
	// ServiceName used for otel tracer and metrics namespace.
	ServiceName string

	// Logger for request logging. If nil, request logging is skipped.
	Logger *Logger

	// Request log
	RequestLogOmitRequestBody     bool
	RequestLogOmitResponseBody    bool
	RequestLogOmitRequestHeaders  bool
	RequestLogOmitResponseHeaders bool
	RequestLogMaxBodyBytes        int

	// CORS
	AllowOrigins string

	// Envelope wrapper
	EnvelopeEnabled bool

	// Sentry
	SentryEnabled bool

	// Metrics
	MetricsNamespace string
}

// LoadStackConfig reads standard middleware configuration from environment variables.
// After calling this, set Logger manually as it is project-specific.
func LoadStackConfig() StackConfig {
	return StackConfig{
		ServiceName:                   GetEnv("APP_ID", "app"),
		RequestLogOmitRequestBody:     GetEnvBool("HTTP_REQUEST_LOG_OMIT_REQUEST_BODY", true),
		RequestLogOmitResponseBody:    GetEnvBool("HTTP_REQUEST_LOG_OMIT_RESPONSE_BODY", true),
		RequestLogOmitRequestHeaders:  GetEnvBool("HTTP_REQUEST_LOG_OMIT_REQUEST_HEADERS", true),
		RequestLogOmitResponseHeaders: GetEnvBool("HTTP_REQUEST_LOG_OMIT_RESPONSE_HEADERS", true),
		RequestLogMaxBodyBytes:        GetEnvInt("HTTP_REQUEST_LOG_MAX_BODY_BYTES", 4096),
		AllowOrigins:               GetEnv("HTTP_ALLOW_ORIGINS", "*"),
		EnvelopeEnabled:            GetEnvBool("HTTP_ENVELOPE_ENABLED", false),
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

	return cfg
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

// RegisterHTTPStack registers HTTP-only middleware: OTel → Metrics → RequestLog.
// Call this after registering WebSocket/SSE routes and before registering REST routes.
// Note: Sentry is registered in RegisterBaseStack (before Recover) so it captures panics with full stack traces.
// Rate limiting is NOT part of this stack — apply AppDeps.Runtime.RateLimit.Strict()/Default()
// to route groups in your Register function instead.
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

	if cfg.EnvelopeEnabled {
		app.Use(NewEnvelopeWrapper())
	}
	registerOTelTrace(app, cfg)
	registerMetrics(app, cfg)
	if cfg.Logger != nil {
		app.Use(NewRequestLog(RequestLogConfig{
			RequestLogger:       NewSlogRequestLogger(cfg.Logger),
			OmitRequestBody:     cfg.RequestLogOmitRequestBody,
			OmitResponseBody:    cfg.RequestLogOmitResponseBody,
			OmitRequestHeaders:  cfg.RequestLogOmitRequestHeaders,
			OmitResponseHeaders: cfg.RequestLogOmitResponseHeaders,
			MaxBodyBytes:        cfg.RequestLogMaxBodyBytes,
		}))
	}
}

func registerCORS(app *fiber.App, cfg StackConfig) {
	app.Use(cors.New(cors.Config{
		AllowOrigins:  splitCSV(cfg.AllowOrigins),
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization", RequestIDHeader},
		ExposeHeaders: []string{RequestIDHeader},
	}))
}

func splitCSV(value string) []string {
	return SplitCSV(value)
}

// SplitCSV splits a comma-separated string, trims each element, and drops empties.
func SplitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
