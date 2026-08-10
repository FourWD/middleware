package infra

import "context"

// Scoped is a Logger view with its component pinned, so a long-lived worker
// or engine names itself once at construction instead of on every call.
//
// Each method stamps WithComponent and WithCallerSkip(1) as defaults — the
// extra frame this wrapper adds would otherwise make every `source` field
// point at logger_scoped.go instead of the caller. Both are prepended, so a
// per-call option for the same key still wins.
type Scoped struct {
	l         *Logger
	component string
}

// Scoped returns a view of l with component pinned. Pass a Component*
// constant from log_taxonomy.go, or a sub-scope of one ("fanout:target_a")
// when a service runs many instances of the same engine.
func (l *Logger) Scoped(component string) Scoped {
	return Scoped{l: l, component: component}
}

// Component is the pinned component string, useful when a caller needs to
// derive a sub-scope from an existing Scoped.
func (s Scoped) Component() string { return s.component }

// Logger is the underlying logger, for the rare call that needs an API this
// wrapper does not expose.
func (s Scoped) Logger() *Logger { return s.l }

func (s Scoped) Info(ctx context.Context, label string, data map[string]any, opts ...LoggerOption) {
	s.l.EventCtx(ctx, label, data, s.defaults(opts)...)
}

func (s Scoped) Warn(ctx context.Context, label string, data map[string]any, opts ...LoggerOption) {
	s.l.WarnCtx(ctx, label, data, s.defaults(opts)...)
}

func (s Scoped) Error(ctx context.Context, err error, label string, data map[string]any, opts ...LoggerOption) {
	s.l.ErrorCtx(ctx, err, label, data, s.defaults(opts)...)
}

func (s Scoped) Lifecycle(label string, data map[string]any, opts ...LoggerOption) {
	s.l.LifecycleEvent(label, data, s.defaults(opts)...)
}

func (s Scoped) LifecycleWarn(label string, data map[string]any, opts ...LoggerOption) {
	s.l.LifecycleWarn(label, data, s.defaults(opts)...)
}

func (s Scoped) LifecycleError(err error, label string, data map[string]any, opts ...LoggerOption) {
	s.l.LifecycleError(err, label, data, s.defaults(opts)...)
}

func (s Scoped) defaults(opts []LoggerOption) []LoggerOption {
	out := make([]LoggerOption, 0, len(opts)+2)
	out = append(out, WithComponent(s.component), WithCallerSkip(1))
	return append(out, opts...)
}
