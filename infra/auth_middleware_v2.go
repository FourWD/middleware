package infra

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// AuthSession is the per-request identity attached by AuthenticationMiddlewareV2.
// ExpiresAt is exposed so the logout flow can size the blacklist TTL to the
// token's remaining lifetime.
type AuthSession struct {
	UserID    string
	Email     string
	Role      string
	ExpiresAt time.Time
}

type authSessionCtxKey int

const authSessionKey authSessionCtxKey = 1

// WithAuthSession attaches a verified session to ctx.
func WithAuthSession(parent context.Context, s AuthSession) context.Context {
	return context.WithValue(parent, authSessionKey, s)
}

// AuthSessionFrom returns the session bound to ctx; ok=false means
// unauthenticated.
func AuthSessionFrom(ctx context.Context) (AuthSession, bool) {
	if ctx == nil {
		return AuthSession{}, false
	}
	s, ok := ctx.Value(authSessionKey).(AuthSession)
	return s, ok
}

// AuthMiddlewareV2Config wires the V2 bearer-token gate.
type AuthMiddlewareV2Config struct {
	// TokenManager owns access-token verification. Required.
	TokenManager *TokenManager

	// Blacklist, if set, is consulted on every authenticated request.
	// Lookup errors fail-open (request proceeds) — failing closed would
	// lock everyone out on a single Mongo/Redis blip.
	Blacklist BlacklistStore

	// ExemptPaths are project-specific routes to skip auth. Auto-merged
	// with: baseline paths (/_ah/warmup, /wake-up, /livez, /readyz,
	// /healthz, /metrics) and HTTP_PUBLIC_PATHS env (CSV). Match is exact;
	// OPTIONS requests are always skipped for CORS preflight.
	ExemptPaths []string

	// DisableAutoExempts turns off the baseline + env merge (rare).
	DisableAutoExempts bool
}

var baselineExemptPaths = []string{
	"/_ah/warmup",
	"/wake-up",
	"/livez",
	"/readyz",
	"/healthz",
	"/metrics",
}

const bearerSchemePrefix = "Bearer "

// AuthenticationMiddlewareV2 requires a valid Bearer access token on every
// non-exempt request. Returns 401 with a uniform error envelope; the "code"
// field distinguishes failures (MISSING_TOKEN, INVALID_TOKEN, WRONG_TOKEN_TYPE,
// TOKEN_REVOKED, AUTH_NOT_CONFIGURED).
func AuthenticationMiddlewareV2(cfg AuthMiddlewareV2Config) fiber.Handler {
	exempt := composeExemptSet(cfg)

	return func(c fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions {
			return c.Next()
		}
		if _, ok := exempt[c.Path()]; ok {
			return c.Next()
		}

		token, errResp := extractBearerToken(c)
		if errResp != nil {
			return errResp
		}

		if cfg.TokenManager == nil {
			return WriteErrorEnvelope(c, fiber.StatusUnauthorized, "AUTH_NOT_CONFIGURED", "token manager not configured")
		}

		claims, err := cfg.TokenManager.ParseAccessToken(token)
		if err != nil {
			code := "INVALID_TOKEN"
			if errors.Is(err, ErrInvalidTokenType) {
				code = "WRONG_TOKEN_TYPE"
			}
			return WriteErrorEnvelope(c, fiber.StatusUnauthorized, code, "invalid or expired session")
		}

		if cfg.Blacklist != nil {
			if revoked, lerr := cfg.Blacklist.IsBlacklisted(c.Context(), token); lerr == nil && revoked {
				return WriteErrorEnvelope(c, fiber.StatusUnauthorized, "TOKEN_REVOKED", "session has been revoked")
			}
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

// composeExemptSet builds the exact-match exempt path set from baseline +
// env + caller config.
func composeExemptSet(cfg AuthMiddlewareV2Config) map[string]struct{} {
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
	return exact
}

// extractBearerToken pulls the token from the Authorization header. Returns
// (token, nil) on success or ("", errorResponse) so the caller can return
// the response directly.
func extractBearerToken(c fiber.Ctx) (string, error) {
	header := c.Get("Authorization")
	if !strings.HasPrefix(header, bearerSchemePrefix) {
		return "", WriteErrorEnvelope(c, fiber.StatusUnauthorized, "MISSING_TOKEN", "missing bearer token")
	}
	token := strings.TrimSpace(header[len(bearerSchemePrefix):])
	if token == "" {
		return "", WriteErrorEnvelope(c, fiber.StatusUnauthorized, "MISSING_TOKEN", "missing bearer token")
	}
	return token, nil
}

// PublicPathsFromEnv reads HTTP_PUBLIC_PATHS (CSV, exact-match). Returns an
// empty slice when unset so callers can append unconditionally.
func PublicPathsFromEnv() []string {
	return SplitCSV(GetEnv("HTTP_PUBLIC_PATHS", ""))
}

// AdminOnly wraps a handler so it runs only for sessions with the given role.
// Returns 401 MISSING_TOKEN when no session in ctx, 403 FORBIDDEN on role
// mismatch.
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
