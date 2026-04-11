package infra

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// OTelConfig holds OpenTelemetry initialization configuration.
// Use LoadOTelConfig() to populate from environment variables.
type OTelConfig struct {
	Enabled           bool
	OTLPTraceEndpoint string
	OTLPTraceInsecure bool
	ServiceName       string
	ServiceVersion    string
	Environment       string
}

// LoadOTelConfig reads OpenTelemetry configuration from environment variables.
func LoadOTelConfig() OTelConfig {
	return OTelConfig{
		Enabled:           GetEnvBool("OTEL_ENABLED", false),
		OTLPTraceEndpoint: GetEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTLPTraceInsecure: GetEnvBool("OTEL_EXPORTER_OTLP_INSECURE", false),
		ServiceName:       GetEnv("APP_NAME", "app"),
		ServiceVersion:    GetEnv("APP_VERSION", "0.1.0"),
		Environment:       GetEnv("APP_ENV", "local"),
	}
}

// SetupOTel initializes the OpenTelemetry tracer provider.
// Returns a shutdown function. Returns a no-op shutdown if OTel is disabled.
func SetupOTel(ctx context.Context, cfg OTelConfig) (func(context.Context) error, error) {
	if !cfg.Enabled || cfg.OTLPTraceEndpoint == "" {
		return func(context.Context) error { return nil }, nil
	}
	if strings.TrimSpace(cfg.ServiceName) == "" {
		cfg.ServiceName = "app"
	}

	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.OTLPTraceEndpoint),
	}
	if cfg.OTLPTraceInsecure {
		options = append(options, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("deployment.environment.name", cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider.Shutdown, nil
}

func registerOTelTrace(app *fiber.App, cfg StackConfig) {
	tracer := otel.Tracer(cfg.ServiceName + "/http")

	app.Use(func(c fiber.Ctx) error {
		ctx, span := tracer.Start(c.Context(), c.Method()+" "+c.Path(), trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		c.SetContext(ctx)
		err := c.Next()

		route := routePath(c)
		span.SetName(c.Method() + " " + route)
		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.route", route),
			attribute.Int("http.status_code", c.Response().StatusCode()),
			attribute.String("http.status_class", statusCodeClass(c.Response().StatusCode())),
			attribute.String("app.request_id", GetRequestID(c)),
		)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		if c.Response().StatusCode() >= fiber.StatusInternalServerError {
			span.SetStatus(codes.Error, "http 5xx response")
		}

		return nil
	})
}
