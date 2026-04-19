package infra

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

const (
	UserIDKey = "user_id"
	ActorKey  = "actor"
)

const (
	CorrelationIDKey = "correlation_id"
	TraceIDKey       = "trace_id"
	SpanIDKey        = "span_id"
	ComponentKey     = "component"
	OperationKey     = "operation"
	LogKindKey       = "log_kind"
	LevelCritical    = slog.Level(12)
)

type loggerContextKey string

const correlationIDContextKey loggerContextKey = "correlation_id"

type LoggerOptions struct {
	disableHook      bool
	callerSkip       int
	additionalFields map[string]any
}

var topLevelLogFields = map[string]struct{}{
	"source":         {},
	"error":          {},
	"stack_trace":    {},
	CorrelationIDKey: {},
	TraceIDKey:       {},
	SpanIDKey:        {},
	ComponentKey:     {},
	OperationKey:     {},
	LogKindKey:       {},
	"request_id":     {},
	"method":         {},
	"route_name":     {},
	"path":           {},
	"status":         {},
	"duration_ms":    {},
	"client_ip":      {},
}

func (o LoggerOptions) DisableHook() bool {
	return o.disableHook
}

type LoggerOption func(*LoggerOptions)

type Message struct {
	format string
	args   []any
}

type Logger struct {
	logger   *slog.Logger
	rootPath string
	ctx      context.Context
	hasCtx   bool
}

type SlogAPILoggerConfig struct {
	LogLevel string

	ProjectName string
	AppID       string
	Environment string
	AppVersion  string
	Hostname    string

	RootPath string
}

func M(msg string) Message {
	return Message{format: msg}
}

func F(format string, args ...any) Message {
	return Message{format: format, args: args}
}

func WithField(key string, value any) LoggerOption {
	return func(o *LoggerOptions) {
		if o.additionalFields == nil {
			o.additionalFields = make(map[string]any)
		}
		o.additionalFields[key] = value
	}
}

func WithDataField(key string, value any) LoggerOption {
	return func(o *LoggerOptions) {
		if o.additionalFields == nil {
			o.additionalFields = make(map[string]any)
		}

		data, _ := o.additionalFields["data"].(map[string]any)
		if data == nil {
			data = make(map[string]any)
		}
		data[key] = value
		o.additionalFields["data"] = data
	}
}

func WithDataFields(fields map[string]any) LoggerOption {
	return func(o *LoggerOptions) {
		if len(fields) == 0 {
			return
		}
		if o.additionalFields == nil {
			o.additionalFields = make(map[string]any)
		}

		data, _ := o.additionalFields["data"].(map[string]any)
		if data == nil {
			data = make(map[string]any, len(fields))
		}
		for key, value := range fields {
			data[key] = value
		}
		o.additionalFields["data"] = data
	}
}

func WithTraceID(id string) LoggerOption {
	return WithField(TraceIDKey, id)
}

func WithSpanID(id string) LoggerOption {
	return WithField(SpanIDKey, id)
}

func WithCorrelationID(id string) LoggerOption {
	return WithField(CorrelationIDKey, id)
}

func WithUserID(id string) LoggerOption {
	return WithField(UserIDKey, id)
}

func WithActor(actor string) LoggerOption {
	return WithField(ActorKey, actor)
}

func WithComponent(component string) LoggerOption {
	return WithField(ComponentKey, component)
}

func WithOperation(operation string) LoggerOption {
	return WithField(OperationKey, operation)
}

func WithLogKind(kind string) LoggerOption {
	return WithField(LogKindKey, kind)
}

func WithContext(ctx context.Context) LoggerOption {
	return func(o *LoggerOptions) {
		if cid, ok := GetCorrelationID(ctx); ok {
			WithCorrelationID(cid)(o)
		}

		span := trace.SpanFromContext(ctx)
		if span == nil {
			return
		}

		spanCtx := span.SpanContext()
		if spanCtx.IsValid() {
			WithTraceID(spanCtx.TraceID().String())(o)
			WithSpanID(spanCtx.SpanID().String())(o)
		}
	}
}

func WithHookDisabled() LoggerOption {
	return func(o *LoggerOptions) {
		o.disableHook = true
	}
}

func WithCallerSkip(n int) LoggerOption {
	return func(o *LoggerOptions) {
		o.callerSkip = n
	}
}

func InjectCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDContextKey, correlationID)
}

func GetCorrelationID(ctx context.Context) (string, bool) {
	value := ctx.Value(correlationIDContextKey)
	correlationID, ok := value.(string)
	return correlationID, ok
}

func NewLoggerWith(cfg CommonConfig) *Logger {
	hostname, _ := os.Hostname()
	return NewSlogAPILogger(os.Stdout, SlogAPILoggerConfig{
		LogLevel:    cfg.LogLevel,
		ProjectName: cfg.AppID,
		AppID:       cfg.AppID,
		Environment: cfg.AppEnv,
		AppVersion:  cfg.AppVersion,
		Hostname:    hostname,
	})
}

func NewSlogAPILogger(w io.Writer, cfg SlogAPILoggerConfig) *Logger {
	if strings.TrimSpace(cfg.LogLevel) == "" {
		cfg.LogLevel = "info"
	}

	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: resolveLogLevel(cfg.LogLevel),
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			switch attr.Key {
			case "msg":
				attr.Key = "message"
			case "time":
				attr.Key = "timestamp"
			case "level":
				attr.Key = "log_level"
				if level, ok := attr.Value.Any().(slog.Level); ok && level == LevelCritical {
					attr.Value = slog.StringValue("critical")
				} else {
					attr.Value = slog.StringValue(strings.ToLower(attr.Value.String()))
				}
			}

			return attr
		},
	})

	baseAttrs := []any{
		"project_name", cfg.ProjectName,
		"app_id", cfg.AppID,
		"env", cfg.Environment,
		"log_type", "api",
	}
	if cfg.AppVersion != "" {
		baseAttrs = append(baseAttrs, "app_version", cfg.AppVersion)
	}
	if cfg.Hostname != "" {
		baseAttrs = append(baseAttrs, "hostname", cfg.Hostname)
	}

	base := slog.New(handler).With(baseAttrs...)

	return &Logger{
		logger:   base,
		rootPath: cfg.RootPath,
	}
}

func (l *Logger) Slog() *slog.Logger {
	return l.logger
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		logger:   l.logger,
		rootPath: l.rootPath,
		ctx:      ctx,
		hasCtx:   true,
	}
}

func (l *Logger) Debug(msg Message, opts ...LoggerOption) {
	l.log(slog.LevelDebug, nil, msg, opts...)
}

func (l *Logger) Info(msg Message, opts ...LoggerOption) {
	l.log(slog.LevelInfo, nil, msg, opts...)
}

func (l *Logger) Warn(msg Message, opts ...LoggerOption) {
	l.log(slog.LevelWarn, nil, msg, opts...)
}

func (l *Logger) Error(err error, msg Message, opts ...LoggerOption) {
	l.log(slog.LevelError, err, msg, opts...)
}

func (l *Logger) CriticalWarning(msg Message, opts ...LoggerOption) {
	l.log(LevelCritical, nil, msg, opts...)
}

func (l *Logger) CriticalError(err error, msg Message, opts ...LoggerOption) {
	l.log(LevelCritical, err, msg, opts...)
}

// Event logs an info-level business event without a context.
// Use for init, cron jobs, or when request-scoped correlation is unavailable.
func (l *Logger) Event(label string, data map[string]any, requestID string, opts ...LoggerOption) {
	all := buildEventOpts(data, requestID, opts)
	l.log(slog.LevelInfo, nil, M(label), all...)
}

// EventWarn logs a warn-level business event without a context.
func (l *Logger) EventWarn(label string, data map[string]any, requestID string, opts ...LoggerOption) {
	all := buildEventOpts(data, requestID, opts)
	l.log(slog.LevelWarn, nil, M(label), all...)
}

// EventError logs an error-level business event without a context.
// Pass err to capture the message and a stack trace.
func (l *Logger) EventError(err error, label string, data map[string]any, requestID string, opts ...LoggerOption) {
	all := buildEventOpts(data, requestID, opts)
	l.log(slog.LevelError, err, M(label), all...)
}

// EventCtx logs an info-level business event using the request context.
// request_id, correlation_id, trace_id, and span_id are extracted automatically.
func (l *Logger) EventCtx(ctx context.Context, label string, data map[string]any, opts ...LoggerOption) {
	all := buildEventCtxOpts(ctx, data, opts)
	l.log(slog.LevelInfo, nil, M(label), all...)
}

// WarnCtx logs a warn-level business event using the request context.
func (l *Logger) WarnCtx(ctx context.Context, label string, data map[string]any, opts ...LoggerOption) {
	all := buildEventCtxOpts(ctx, data, opts)
	l.log(slog.LevelWarn, nil, M(label), all...)
}

// ErrorCtx logs an error-level business event using the request context.
// Pass err to capture the message and a stack trace.
func (l *Logger) ErrorCtx(ctx context.Context, err error, label string, data map[string]any, opts ...LoggerOption) {
	all := buildEventCtxOpts(ctx, data, opts)
	l.log(slog.LevelError, err, M(label), all...)
}

func buildEventOpts(data map[string]any, requestID string, extra []LoggerOption) []LoggerOption {
	all := make([]LoggerOption, 0, len(extra)+2)
	if requestID != "" {
		all = append(all, WithField("request_id", requestID))
	}
	if len(data) > 0 {
		all = append(all, WithDataFields(data))
	}
	all = append(all, extra...)
	return all
}

func buildEventCtxOpts(ctx context.Context, data map[string]any, extra []LoggerOption) []LoggerOption {
	all := make([]LoggerOption, 0, len(extra)+3)
	if ctx != nil {
		all = append(all, WithContext(ctx))
		if rid, ok := GetCorrelationID(ctx); ok {
			all = append(all, WithField("request_id", rid))
		}
	}
	if len(data) > 0 {
		all = append(all, WithDataFields(data))
	}
	all = append(all, extra...)
	return all
}

func (l *Logger) log(level slog.Level, err error, msg Message, opts ...LoggerOption) {
	logCtx := context.Background()
	if l.hasCtx {
		logCtx = l.ctx
		opts = append([]LoggerOption{WithContext(l.ctx)}, opts...)
	}

	options := MergeOptions[LoggerOptions](opts...)
	if !options.DisableHook() {
		addSource(&options, l.rootPath, options.callerSkip)
	}

	attrs := toAttrs(options)
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		if level >= slog.LevelError {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			attrs = append(attrs, slog.String("stack_trace", string(buf[:n])))
		}
	}

	l.logger.LogAttrs(logCtx, level, fmt.Sprintf(msg.format, msg.args...), attrs...)
}

func toAttrs(options LoggerOptions) []slog.Attr {
	if len(options.additionalFields) == 0 {
		return nil
	}

	dataFields := map[string]any{}
	attrs := make([]slog.Attr, 0, len(options.additionalFields))
	for key, value := range options.additionalFields {
		if key == "data" {
			if data, ok := value.(map[string]any); ok {
				for dataKey, dataValue := range data {
					dataFields[dataKey] = dataValue
				}
			}
			continue
		}

		if _, ok := topLevelLogFields[key]; ok {
			attrs = append(attrs, slog.Any(key, value))
			continue
		}

		dataFields[key] = value
	}

	if len(dataFields) > 0 {
		attrs = append(attrs, slog.Any("data", dataFields))
	}

	return attrs
}

func addSource(options *LoggerOptions, rootPath string, skip int) {
	if options.additionalFields == nil {
		options.additionalFields = make(map[string]any)
	}

	_, file, line, ok := runtime.Caller(3 + skip)
	if !ok {
		return
	}

	if rootPath != "" {
		file = strings.TrimPrefix(file, rootPath)
		file = strings.TrimPrefix(file, `\`)
		file = strings.TrimPrefix(file, `/`)
	}

	options.additionalFields["source"] = fmt.Sprintf("%s:%d", file, line)
}

func resolveLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "critical":
		return LevelCritical
	default:
		return slog.LevelInfo
	}
}
