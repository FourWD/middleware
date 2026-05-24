package infra

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"mime"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel/trace"
)

const (
	omittedBodyValue   = "<omit>"
	redactedFieldValue = "<redacted>"
)

// sensitiveFieldNames is the canonical PII deny-list used wherever the
// log surface accepts a caller-supplied identifier — JSON body fields,
// URL query parameters, and (via WithField) structured log attributes.
//
// Match is case-insensitive. Adding a key here protects ALL of: body
// scrubber, form-body scrubber, URL query stripper, and the
// _truncated marker's _keys redaction in capDataPayload.
var sensitiveFieldNames = map[string]struct{}{
	"password":      {},
	"passwd":        {},
	"secret":        {},
	"encrypt_key":   {},
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"id_token":      {},
	"jwt":           {},
	"api_key":       {},
	"apikey":        {},
	"pin":           {},
	"otp":           {},
	"cvv":           {},
	"card_number":   {},
}

// sensitiveHeaderNames covers HTTP header names — a distinct namespace
// from field names (e.g. "authorization" is a header but never a JSON
// field, "password" is a field but never a header). Match is
// case-insensitive.
var sensitiveHeaderNames = map[string]struct{}{
	"authorization": {},
	"cookie":        {},
	"set-cookie":    {},
	"x-api-key":     {},
	"x-auth-token":  {},
}

// loggableContentTypes is the whitelist of media types whose bodies are
// safe and useful to capture in request logs. Everything else gets the
// omitted marker so we don't accidentally log binary payloads.
var loggableContentTypes = map[string]struct{}{
	"text/csv":                          {},
	"text/html":                         {},
	"text/plain":                        {},
	"application/json":                  {},
	"application/problem+json":          {},
	"application/xml":                   {},
	"application/x-www-form-urlencoded": {},
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
	RequestLogger       RequestLogger
	OmitRequestBody     bool
	OmitResponseBody    bool
	OmitRequestHeaders  bool
	OmitResponseHeaders bool
	MaxBodyBytes        int

	// SamplingRate is the probability (0–1) that a non-error response is
	// logged. 0 (default) means "log everything". 0.1 means "log 10% of
	// 2xx/3xx responses". Errors (>= AlwaysLogStatusFrom) always log.
	SamplingRate float64

	// AlwaysLogStatusFrom is the HTTP status threshold below which sampling
	// applies. Default (0) is 400, so all 4xx/5xx responses bypass sampling.
	AlwaysLogStatusFrom int
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
	if cfg.AlwaysLogStatusFrom <= 0 {
		cfg.AlwaysLogStatusFrom = fiber.StatusBadRequest
	}

	return func(c fiber.Ctx) error {
		startedAt := time.Now()
		err := c.Next()

		if cfg.RequestLogger == nil {
			return err
		}

		status := c.Response().StatusCode()
		if !shouldLog(cfg, status) {
			return err
		}

		entry := buildEntry(c, cfg, startedAt, status)
		cfg.RequestLogger.Log(c.Context(), entry, buildRequestLogOptions(c, entry)...)
		return err
	}
}

// shouldLog applies the sampling rate. Errors at or above
// AlwaysLogStatusFrom always log; others log with SamplingRate
// probability (0 = log everything).
func shouldLog(cfg RequestLogConfig, status int) bool {
	if cfg.SamplingRate <= 0 || status >= cfg.AlwaysLogStatusFrom {
		return true
	}
	return rand.Float64() < cfg.SamplingRate
}

func buildEntry(c fiber.Ctx, cfg RequestLogConfig, startedAt time.Time, status int) Entry {
	var reqHeaders, respHeaders map[string][]string
	if !cfg.OmitRequestHeaders {
		reqHeaders = sanitizeHeaders(c.GetReqHeaders())
	}
	if !cfg.OmitResponseHeaders {
		respHeaders = sanitizeHeaders(c.GetRespHeaders())
	}

	return Entry{
		Request: RequestEntry{
			Headers:     reqHeaders,
			FullURI:     sanitizeURL(c.OriginalURL()),
			RelativeURI: c.Path(),
			Method:      c.Method(),
			Body:        bodyString(c.BodyRaw(), c.Get("Content-Type"), cfg.OmitRequestBody, cfg.MaxBodyBytes),
		},
		Response: ResponseEntry{
			Headers:  respHeaders,
			Status:   status,
			Body:     bodyString(c.Response().Body(), c.GetRespHeader("Content-Type"), cfg.OmitResponseBody, cfg.MaxBodyBytes),
			Duration: time.Since(startedAt),
		},
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

	fields := buildRequestLogFields(entry, route, opts)
	message := F(
		"http request %s %s -> %d (%dms)",
		entry.Request.Method,
		route,
		entry.Response.Status,
		entry.Response.Duration.Milliseconds(),
	)

	// Route severity by HTTP status. Uses the unexported `log()` method
	// directly so this status-based dispatch does not trigger the Level-
	// method deprecation warnings — request log is the one place where
	// generic severity routing is legitimate.
	dispatchByStatus(l.logger.WithContext(ctx), entry.Response.Status, message, fields)
}

func buildRequestLogFields(entry Entry, route string, opts RequestLoggerOptions) []LoggerOption {
	fields := []LoggerOption{
		WithField("method", entry.Request.Method),
		WithField("route_name", route),
		WithField("raw_uri", entry.Request.FullURI),
		WithField("req_body", entry.Request.Body),
		WithField("resp_body", entry.Response.Body),
		WithComponent(ComponentHTTP),
		WithOperation("request"),
		WithLogKind(LogKindRequest),
		WithoutSource(),
	}

	if entry.Request.Headers != nil {
		fields = append(fields, WithField("req_header", marshalUnfailable(entry.Request.Headers)))
	}
	if entry.Response.Headers != nil {
		fields = append(fields, WithField("resp_header", marshalUnfailable(entry.Response.Headers)))
	}

	for key, value := range opts.additionalFields {
		if key == "route" {
			continue
		}
		fields = append(fields, WithField(key, value))
	}
	return fields
}

func dispatchByStatus(logger *Logger, status int, message Message, fields []LoggerOption) {
	switch {
	case status >= fiber.StatusInternalServerError:
		logger.log(slog.LevelError, nil, message, fields...)
	case status >= fiber.StatusBadRequest:
		logger.log(slog.LevelWarn, nil, message, fields...)
	default:
		logger.log(slog.LevelInfo, nil, message, fields...)
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

	scrubbed := scrubBody(body, contentType)
	if maxBytes > 0 && len(scrubbed) > maxBytes {
		return string(scrubbed[:maxBytes]) + "...<truncated>"
	}
	return string(scrubbed)
}

// scrubBody dispatches to the right scrubber based on Content-Type. Bodies
// in unknown formats are returned as-is — we never make scrubbing a
// gatekeeper because a missing media type should not block diagnostics.
func scrubBody(body []byte, contentType string) []byte {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	switch mediaType {
	case "application/json", "application/problem+json":
		return scrubJSONBody(body)
	case "application/x-www-form-urlencoded":
		return scrubFormBody(body)
	default:
		return body
	}
}

// scrubJSONBody redacts sensitive fields. Parse failures fall through to
// the raw body so a malformed payload stays inspectable.
func scrubJSONBody(body []byte) []byte {
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return body
	}
	out, err := json.Marshal(scrubValue(payload))
	if err != nil {
		return body
	}
	return out
}

// scrubFormBody redacts sensitive fields from a urlencoded form body.
// Login / OAuth token-exchange / password-reset flows typically POST
// `application/x-www-form-urlencoded` and would otherwise leak the
// credential through req_body.
func scrubFormBody(body []byte) []byte {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return body
	}
	dirty := false
	for k := range values {
		if _, sensitive := sensitiveFieldNames[strings.ToLower(k)]; sensitive {
			values.Set(k, redactedFieldValue)
			dirty = true
		}
	}
	if !dirty {
		return body
	}
	return []byte(values.Encode())
}

func scrubValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if _, sensitive := sensitiveFieldNames[strings.ToLower(k)]; sensitive {
				x[k] = redactedFieldValue
				continue
			}
			x[k] = scrubValue(val)
		}
		return x
	case []any:
		for i, val := range x {
			x[i] = scrubValue(val)
		}
		return x
	default:
		return v
	}
}

// sanitizeURL strips sensitive query parameters so a token leaked into a
// callback URL does not enter the log. The path is preserved verbatim;
// the query is re-encoded with sensitive keys replaced by the redacted
// marker. Any URL fragment is also stripped — HTTP wire format never
// carries fragments, but a fragment may sneak in via internal redirects
// or client-injected URLs, and a fragment can itself carry a token.
func sanitizeURL(raw string) string {
	if raw == "" {
		return raw
	}
	// url.Parse handles both absolute and relative refs and preserves the
	// Fragment; url.ParseRequestURI silently drops the fragment per RFC,
	// which would leave a token in the fragment uncaught.
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	dirty := false
	if u.Fragment != "" {
		u.Fragment = ""
		u.RawFragment = ""
		dirty = true
	}
	if u.RawQuery != "" {
		values := u.Query()
		for k := range values {
			if _, sensitive := sensitiveFieldNames[strings.ToLower(k)]; sensitive {
				values.Set(k, redactedFieldValue)
				dirty = true
			}
		}
		u.RawQuery = values.Encode()
	}
	if !dirty {
		return raw
	}
	return u.String()
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
		if _, ok := sensitiveHeaderNames[strings.ToLower(key)]; ok {
			result[key] = []string{redactedFieldValue}
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
	_, ok := loggableContentTypes[mediaType]
	return ok
}
