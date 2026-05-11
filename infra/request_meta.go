package infra

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

// RequestMeta is the per-request network identity propagated across the
// use-case boundary via context.Context. The request log already captures
// these for its own logger; exposing them through ctx lets audit-log
// writers (and anything else that records "who called from where") read
// the same values without re-plumbing fiber.Ctx all the way down.
//
// Auto-populated on every request by NewApp's pre-stack middleware —
// callers just read from ctx.
type RequestMeta struct {
	IP        string
	UserAgent string
}

type requestMetaCtxKey int

const requestMetaKey requestMetaCtxKey = 1

// WithRequestMeta attaches request metadata to a child ctx. The auto
// middleware in NewApp calls this once per request; manual call sites
// can use it for tests or for cases where a non-HTTP entrypoint wants
// to carry IP/UA forward.
func WithRequestMeta(parent context.Context, m RequestMeta) context.Context {
	return context.WithValue(parent, requestMetaKey, m)
}

// RequestMetaFrom returns the metadata bound to ctx. The bool is false
// when no middleware has attached one — callers default to empty.
func RequestMetaFrom(ctx context.Context) (RequestMeta, bool) {
	if ctx == nil {
		return RequestMeta{}, false
	}
	m, ok := ctx.Value(requestMetaKey).(RequestMeta)
	return m, ok
}

// NewRequestMetaMiddleware returns a fiber middleware that stamps the
// client IP + User-Agent onto every request's ctx. Mounted very early in
// the stack — before auth — so even public endpoints (login, refresh)
// can read who called from where.
func NewRequestMetaMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.SetContext(WithRequestMeta(c.Context(), RequestMeta{
			IP:        c.IP(),
			UserAgent: string(c.Request().Header.UserAgent()),
		}))
		return c.Next()
	}
}
