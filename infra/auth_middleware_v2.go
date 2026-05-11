package infra

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// AuthSession is the per-request identity the V2 middleware attaches to
// ctx. It is a thin projection of TokenManager's Claims — handlers that
// don't need the raw JWT internals (jti, issuer, etc.) can stick to this
// shape and stay decoupled from `golang-jwt/jwt/v5`.
//
// ExpiresAt is surfaced so the logout flow can size the blacklist TTL
// to the token's remaining lifetime; past that point signature verify
// rejects on its own, so the blacklist entry is wasted storage.
type AuthSession struct {
	UserID    string
	Email     string
	Role      string
	ExpiresAt time.Time
}

type authSessionCtxKey int

const authSessionKey authSessionCtxKey = 1

// WithAuthSession attaches a verified session to the ctx. The V2
// middleware calls this on every authenticated request; handlers read
// it back through AuthSessionFrom.
func WithAuthSession(parent context.Context, s AuthSession) context.Context {
	return context.WithValue(parent, authSessionKey, s)
}

// AuthSessionFrom returns the session bound to ctx (typically by the
// V2 middleware). The bool is false when ctx carries no session —
// callers should treat that as unauthenticated.
func AuthSessionFrom(ctx context.Context) (AuthSession, bool) {
	if ctx == nil {
		return AuthSession{}, false
	}
	s, ok := ctx.Value(authSessionKey).(AuthSession)
	return s, ok
}

// AuthMiddlewareV2Config wires the V2 bearer-token gate.
type AuthMiddlewareV2Config struct {
	// TokenManager is required — it owns access-token verification.
	TokenManager *TokenManager
	// Blacklist is optional. When non-nil, every authenticated request
	// also checks IsBlacklisted(token). A blacklist lookup error does
	// NOT block the request (fail-open) — failing closed would lock
	// everyone out on a single Mongo/Redis blip. Flip the behaviour at
	// the call site if you need fail-closed semantics on prod traffic.
	Blacklist BlacklistStore
	// ExemptPaths is the project-specific list of paths to skip. The
	// middleware AUTO-MERGES this with:
	//
	//   1. Service-built-ins NewApp mounts itself — /_ah/warmup,
	//      /wake-up, /livez, /readyz, /healthz, /metrics.
	//   2. HTTP_PUBLIC_PATHS env (CSV, exact-match) so operators can
	//      open a path without a code change.
	//
	// You only need to list project routes here (e.g.
	// "/api/v1/auth/google"). Set DisableAutoExempts when a project
	// really wants to gate even the built-ins (rare).
	//
	// Matching is exact (no prefix, no regex). An entry of "/" would
	// let every request through, so prefix-style entries are
	// intentionally unsupported. The middleware skips OPTIONS so CORS
	// preflight isn't affected.
	ExemptPaths []string
	// DisableAutoExempts turns off the baseline + env merge described
	// above. Default false — keep the convenience.
	DisableAutoExempts bool
}

// baselineExemptPaths is the list of mwinfra-managed routes that NewApp
// mounts itself: GAE lifecycle endpoints + standard health/metrics
// surfaces. Auto-included by AuthenticationMiddlewareV2 unless the
// caller opts out via DisableAutoExempts.
var baselineExemptPaths = []string{
	"/_ah/warmup",
	"/wake-up",
	"/livez",
	"/readyz",
	"/healthz",
	"/metrics",
}

// AuthenticationMiddlewareV2 returns middleware that requires a valid
// Bearer access token on every request. Replaces the legacy
// AuthenticationMiddleware (which uses the old JWTClaims shape) for
// projects on TokenManager + Claims + optional blacklist.
//
// Differences vs the V1 middleware:
//
//   - Verifies via TokenManager.ParseAccessToken (rejects refresh
//     tokens explicitly via token_type claim).
//   - Optional Mongo/Redis blacklist check.
//   - Strict exact-match exempt path list (V1 also exact, but reads
//     HTTP_PUBLIC_PATHS from env and supports regex).
//   - Surfaces an AuthSession to ctx (handlers don't reach into
//     fiber.Locals or re-parse JWT).
//
// Error responses are uniform:
//
//	{ "success": false, "error": { "code": "...", "message": "..." } }
//
// with HTTP 401 for missing/invalid/expired/revoked, distinguished by
// the "code" field (MISSING_TOKEN, INVALID_TOKEN, TOKEN_REVOKED).
func AuthenticationMiddlewareV2(cfg AuthMiddlewareV2Config) fiber.Handler {
	// Compose the final exempt set once at registration time. Order of
	// the merge doesn't matter (set semantics) — overlap with the
	// caller's list collapses naturally.
	var merged []string
	if !cfg.DisableAutoExempts {
		merged = append(merged, baselineExemptPaths...)
		merged = append(merged, PublicPathsFromEnv()...)
	}
	merged = append(merged, cfg.ExemptPaths...)
	exact := make(map[string]struct{}, len(merged))
	for _, p := range merged {
		exact[p] = struct{}{}
	}
	return func(c fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions {
			return c.Next()
		}
		path := c.Path()
		if _, ok := exact[path]; ok {
			return c.Next()
		}
		header := c.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			return WriteErrorEnvelope(c, fiber.StatusUnauthorized, "MISSING_TOKEN", "missing bearer token")
		}
		token := strings.TrimSpace(header[len(prefix):])
		if token == "" {
			return WriteErrorEnvelope(c, fiber.StatusUnauthorized, "MISSING_TOKEN", "missing bearer token")
		}
		if cfg.TokenManager == nil {
			return WriteErrorEnvelope(c, fiber.StatusUnauthorized, "AUTH_NOT_CONFIGURED", "token manager not configured")
		}
		claims, err := cfg.TokenManager.ParseAccessToken(token)
		if err != nil {
			// jwt-go bundles signature/expiry/malformed under generic
			// errors. We collapse them to a single client-visible code
			// so the React side can take one branch on 401.
			code := "INVALID_TOKEN"
			if errors.Is(err, ErrInvalidTokenType) {
				code = "WRONG_TOKEN_TYPE"
			}
			return WriteErrorEnvelope(c, fiber.StatusUnauthorized, code, "invalid or expired session")
		}
		if cfg.Blacklist != nil {
			revoked, lerr := cfg.Blacklist.IsBlacklisted(c.Context(), token)
			if lerr == nil && revoked {
				return WriteErrorEnvelope(c, fiber.StatusUnauthorized, "TOKEN_REVOKED", "session has been revoked")
			}
			// On a blacklist lookup error we fall through (fail-open);
			// the session was still cryptographically valid and the
			// request log captures the lookup failure.
		}
		var exp time.Time
		if claims.ExpiresAt != nil {
			exp = claims.ExpiresAt.Time
		}
		c.SetContext(WithAuthSession(c.Context(), AuthSession{
			UserID:    claims.UserID,
			Email:     claims.Email,
			Role:      claims.Role,
			ExpiresAt: exp,
		}))
		return c.Next()
	}
}

// PublicPathsFromEnv reads HTTP_PUBLIC_PATHS (CSV, exact-match) from the
// process environment and returns the parsed list. Provided so callers
// who want V1-style env-driven exempts can opt in without re-deriving
// the env shape:
//
//	web.Use(infra.AuthenticationMiddlewareV2(infra.AuthMiddlewareV2Config{
//	    TokenManager: tm,
//	    Blacklist:    deps.Security.BlacklistStore,
//	    ExemptPaths: append(
//	        []string{"/api/v1/auth/google", "/api/v1/auth/refresh", ...},
//	        infra.PublicPathsFromEnv()...,
//	    ),
//	}))
//
// Returned slice is empty when the env var is unset — `append(x, nil...)`
// is a no-op so callers don't need a nil-check.
func PublicPathsFromEnv() []string {
	return SplitCSV(GetEnv("HTTP_PUBLIC_PATHS", ""))
}

// AdminOnly wraps a handler so it only runs for sessions with the
// supplied role. Use at route-registration time, after the V2 auth
// middleware is installed:
//
//	v1.Delete("/users/:id", infra.AdminOnly("Admin", deps.Handlers.Users.Delete))
//
// A handler running without a session in ctx returns 401 MISSING_TOKEN;
// a wrong-role session returns 403 FORBIDDEN.
func AdminOnly(role string, h fiber.Handler) fiber.Handler {
	return func(c fiber.Ctx) error {
		s, ok := AuthSessionFrom(c.Context())
		if !ok {
			return WriteErrorEnvelope(c, fiber.StatusUnauthorized, "MISSING_TOKEN", "missing bearer token")
		}
		if s.Role != role {
			return WriteErrorEnvelope(c, fiber.StatusForbidden, "FORBIDDEN", role+" role required")
		}
		return h(c)
	}
}
