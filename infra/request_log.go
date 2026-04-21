package infra

import (
	"context"
	"encoding/json"
	"mime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel/trace"
)

const omittedBodyValue = "<omit>"

var sensitiveHeaders = map[string]struct{}{
	"authorization": {},
	"cookie":        {},
	"set-cookie":    {},
	"x-api-key":     {},
	"x-auth-token":  {},
}

type RequestEntry struct {
	Headers     map[string][]string `json:"headers"`
	FullURI     string              `json:"full_uri"`
	RelativeURI string              `json:"relative_uri"`
	Method      string              `json:"method"`
	Body        string              `json:"body"`
}

type ResponseEntry struct {
	Headers  map[string][]string `json:"headers"`
	Status   int                 `json:"status"`
	Body     string              `json:"body"`
	Duration time.Duration       `json:"duration"`
}

type Entry struct {
	Request  RequestEntry  `json:"request"`
	Response ResponseEntry `json:"response"`
}

type RequestLoggerOptions struct {
	additionalFields map[string]any
}

type RequestLoggerOption func(*RequestLoggerOptions)

type RequestLogger interface {
	Log(ctx context.Context, entry Entry, options ...RequestLoggerOption)
}

type SlogRequestLogger struct {
	logger *Logger
}

type RequestLogConfig struct {
	RequestLogger    RequestLogger
	OmitRequestBody  bool
	OmitResponseBody bool
	MaxBodyBytes     int
}

func WithRequestLogField(key string, value any) RequestLoggerOption {
	return func(o *RequestLoggerOptions) {
		if o.additionalFields == nil {
			o.additionalFields = make(map[string]any)
		}
		o.additionalFields[key] = value
	}
}

func NewSlogRequestLogger(logger *Logger) *SlogRequestLogger {
	return &SlogRequestLogger{logger: logger}
}

func NewRequestLog(cfg RequestLogConfig) fiber.Handler {
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 4096
	}

	return func(c fiber.Ctx) error {
		startedAt := time.Now()
		err := c.Next()

		if cfg.RequestLogger == nil {
			return err
		}

		entry := Entry{
			Request: RequestEntry{
				Headers:     sanitizeHeaders(c.GetReqHeaders()),
				FullURI:     c.OriginalURL(),
				RelativeURI: c.Path(),
				Method:      c.Method(),
				Body:        bodyString(c.BodyRaw(), c.Get("Content-Type"), cfg.OmitRequestBody, cfg.MaxBodyBytes),
			},
			Response: ResponseEntry{
				Headers:  sanitizeHeaders(c.GetRespHeaders()),
				Status:   c.Response().StatusCode(),
				Body:     bodyString(c.Response().Body(), c.GetRespHeader("Content-Type"), cfg.OmitResponseBody, cfg.MaxBodyBytes),
				Duration: time.Since(startedAt),
			},
		}

		options := buildRequestLogOptions(c, entry)
		cfg.RequestLogger.Log(c.Context(), entry, options...)

		return err
	}
}

func (l *SlogRequestLogger) Log(ctx context.Context, entry Entry, options ...RequestLoggerOption) {
	if l == nil || l.logger == nil {
		return
	}

	opts := MergeOptions[RequestLoggerOptions](options...)
	route := entry.Request.RelativeURI
	if value, ok := opts.additionalFields["route"].(string); ok && value != "" {
		route = value
	}
	fields := []LoggerOption{
		WithField("method", entry.Request.Method),
		WithField("route_name", route),
		WithField("raw_uri", entry.Request.FullURI),
		WithField("req_header", marshalUnfailable(entry.Request.Headers)),
		WithField("req_body", entry.Request.Body),
		WithField("resp_header", marshalUnfailable(entry.Response.Headers)),
		WithField("resp_body", entry.Response.Body),
		WithComponent("http"),
		WithOperation("request"),
		WithLogKind("request"),
		WithoutSource(),
	}

	for key, value := range opts.additionalFields {
		fields = append(fields, WithField(key, value))
	}

	message := F(
		"http request %s %s -> %d (%dms)",
		entry.Request.Method,
		route,
		entry.Response.Status,
		entry.Response.Duration.Milliseconds(),
	)

	logger := l.logger.WithContext(ctx)
	switch {
	case entry.Response.Status >= fiber.StatusInternalServerError:
		logger.Error(nil, message, fields...)
	case entry.Response.Status >= fiber.StatusBadRequest:
		logger.Warn(message, fields...)
	default:
		logger.Info(message, fields...)
	}
}

func buildRequestLogOptions(c fiber.Ctx, entry Entry) []RequestLoggerOption {
	spanCtx := trace.SpanContextFromContext(c.Context())

	options := []RequestLoggerOption{
		WithRequestLogField("route", routePath(c)),
		WithRequestLogField("path", c.Path()),
		WithRequestLogField("status", entry.Response.Status),
		WithRequestLogField("request_id", GetRequestID(c)),
		WithRequestLogField("duration_ms", entry.Response.Duration.Milliseconds()),
		WithRequestLogField("client_ip", c.IP()),
	}

	if spanCtx.IsValid() {
		options = append(
			options,
			WithRequestLogField("trace_id", spanCtx.TraceID().String()),
			WithRequestLogField("span_id", spanCtx.SpanID().String()),
		)
	}

	return options
}

func bodyString(body []byte, contentType string, omit bool, maxBytes int) string {
	if omit {
		return omittedBodyValue
	}
	if len(body) == 0 {
		return ""
	}
	if !isLoggableContentType(contentType) {
		return omittedBodyValue
	}
	if maxBytes > 0 && len(body) > maxBytes {
		return string(body[:maxBytes]) + "...<truncated>"
	}

	return string(body)
}

func marshalUnfailable(value any) string {
	b, err := json.Marshal(value)
	if err != nil {
		return "<error>"
	}

	return string(b)
}

func sanitizeHeaders(headers map[string][]string) map[string][]string {
	if len(headers) == 0 {
		return nil
	}

	result := make(map[string][]string, len(headers))
	for key, values := range headers {
		if _, ok := sensitiveHeaders[strings.ToLower(key)]; ok {
			result[key] = []string{omittedBodyValue}
			continue
		}

		copied := make([]string, len(values))
		copy(copied, values)
		result[key] = copied
	}

	return result
}

func isLoggableContentType(contentType string) bool {
	if contentType == "" {
		return true
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	if strings.HasPrefix(mediaType, "text/") {
		return true
	}

	switch mediaType {
	case
		"text/csv",
		"text/html",
		"text/plain",
		"application/json",
		"application/problem+json",
		"application/xml",
		"application/x-www-form-urlencoded":
		return true
	default:
		return false
	}
}
