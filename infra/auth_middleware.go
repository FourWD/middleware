package infra

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// JWTClaims is the legacy claims shape used by GenerateJWTToken and
// AuthenticationMiddleware. New code should prefer Claims + TokenManager.
type JWTClaims struct {
	UserID string            `json:"user_id"`
	Role   string            `json:"role"`
	Remark map[string]string `json:"remark"`
	jwt.RegisteredClaims
}

var baselinePublicPaths = []string{"/_ah/warmup", "/wake-up", "/metrics"}

// publicPathMatcher classifies an incoming path against HTTP_PUBLIC_PATHS +
// the baseline list. Exact strings hit an O(1) map; regex-looking patterns
// are compiled once and walked sequentially.
type publicPathMatcher struct {
	exact   map[string]struct{}
	regexes []*regexp.Regexp
}

func (m *publicPathMatcher) matches(path string) bool {
	if _, ok := m.exact[path]; ok {
		return true
	}
	for _, re := range m.regexes {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

var publicPaths = sync.OnceValue(func() *publicPathMatcher {
	patterns := append(SplitCSV(GetEnv("HTTP_PUBLIC_PATHS", "")), baselinePublicPaths...)
	m := &publicPathMatcher{exact: map[string]struct{}{}}
	for _, raw := range patterns {
		p := strings.TrimSpace(raw)
		if p == "" {
			continue
		}
		if !looksLikeRegexPattern(p) {
			m.exact[p] = struct{}{}
			continue
		}
		if re, err := regexp.Compile(p); err == nil {
			m.regexes = append(m.regexes, re)
		}
	}
	return m
})

var jwtSecret = sync.OnceValue(func() []byte {
	return []byte(GetEnv("JWT_SECRET", ""))
})

// AuthenticationMiddleware validates a Bearer JWT on the Authorization header
// and stores the parsed claims under c.Locals("user"). Routes matching
// HTTP_PUBLIC_PATHS (or the hardcoded /_ah/warmup, /wake-up, /metrics) skip
// the check. NewApp registers this automatically.
func AuthenticationMiddleware(c fiber.Ctx) error {
	if publicPaths().matches(c.Path()) {
		return c.Next()
	}
	return checkAuth(c)
}

func looksLikeRegexPattern(pattern string) bool {
	return strings.ContainsAny(pattern, `\.^$*+?()[]{}|`)
}

func checkAuth(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(http.StatusUnauthorized).SendString("No token provided")
	}

	if !IsJwtValid(authHeader) {
		return c.Status(http.StatusUnauthorized).SendString("token blacklist")
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(http.StatusUnauthorized).SendString("Invalid authorization header format")
	}

	tokenString := authHeader[7:]
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret(), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return c.Status(http.StatusUnauthorized).SendString("Invalid token signature")
		}
		if errors.Is(err, jwt.ErrTokenExpired) {
			return c.Status(http.StatusUnauthorized).SendString("Token expired")
		}
		return c.Status(http.StatusUnauthorized).SendString("Invalid token")
	}

	if !token.Valid {
		return c.Status(http.StatusUnauthorized).SendString("Invalid token")
	}

	c.Locals("user", claims)
	return c.Next()
}

// IsJwtValid returns false if the token is in the blacklist stored on the
// dedicated middleware Mongo cluster (MongoMiddleware).
// Returns true when MongoMiddleware is not initialized (blacklist disabled).
// Fails closed: returns false on lookup error (treats token as invalid) —
// safer default than letting potentially-revoked tokens through when Mongo
// is unreachable.
func IsJwtValid(token string) bool {
	if MongoMiddleware == nil {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	collection := MongoMiddleware.Database().Collection("blacklist_tokens")
	count, err := collection.CountDocuments(ctx, bson.M{"token_hash": hashBlacklistToken(token)})
	if err != nil {
		// NOTE: token value intentionally NOT logged (PII deny-list).
		AppLog.EventError(err, "JWT_BLACKLIST_LOOKUP_FAILURE", nil, "",
			WithComponent(ComponentAuth),
			WithOperation("blacklist_check"),
			WithLogKind(LogKindError))
		return false
	}
	return count == 0
}
