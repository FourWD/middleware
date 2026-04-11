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

	ProjectName     string
	ApplicationName string
	CompanyName     string
	Environment     string
	ServiceVersion  string
	Hostname        string

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

func InjectCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDContextKey, correlationID)
}

func GetCorrelationID(ctx context.Context) (string, bool) {
	value := ctx.Value(correlationIDContextKey)
	correlationID, ok := value.(string)
	return correlationID, ok
}

func NewLoggerWith(appName, env string) *Logger {
	hostname, _ := os.Hostname()
	return NewSlogAPILogger(os.Stdout, SlogAPILoggerConfig{
		LogLevel:        strings.ToLower(GetEnv("LOG_LEVEL", "info")),
		ProjectName:     appName,
		ApplicationName: appName,
		CompanyName:     GetEnv("COMPANY_NAME", "company"),
		Environment:     env,
		ServiceVersion:  GetEnv("SERVICE_VERSION", ""),
		Hostname:        hostname,
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
		"app_name", cfg.ApplicationName,
		"company", cfg.CompanyName,
		"env", cfg.Environment,
		"log_type", "api",
	}
	if cfg.ServiceVersion != "" {
		baseAttrs = append(baseAttrs, "service_version", cfg.ServiceVersion)
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

func (l *Logger) log(level slog.Level, err error, msg Message, opts ...LoggerOption) {
	logCtx := context.Background()
	if l.hasCtx {
		logCtx = l.ctx
		opts = append([]LoggerOption{WithContext(l.ctx)}, opts...)
	}

	options := MergeOptions[LoggerOptions](opts...)
	if !options.DisableHook() {
		addSource(&options, l.rootPath)
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

func addSource(options *LoggerOptions, rootPath string) {
	if options.additionalFields == nil {
		options.additionalFields = make(map[string]any)
	}

	_, file, line, ok := runtime.Caller(3)
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
