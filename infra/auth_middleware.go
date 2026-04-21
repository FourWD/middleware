package infra

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

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

// AuthenticationMiddleware validates a Bearer JWT on the Authorization header
// and stores the parsed claims under c.Locals("user"). Routes matching
// HTTP_PUBLIC_PATHS (or the hardcoded /_ah/warmup, /wake-up, /metrics) skip
// the check. NewApp registers this automatically.
func AuthenticationMiddleware(c fiber.Ctx) error {
	if isPublicPath(c) {
		return c.Next()
	}
	return checkAuth(c)
}

func isPublicPath(c fiber.Ctx) bool {
	publicPaths := SplitCSV(GetEnv("HTTP_PUBLIC_PATHS", ""))
	hardcodePaths := []string{"/_ah/warmup", "/wake-up", "/metrics"}
	publicPaths = append(publicPaths, hardcodePaths...)
	for _, pattern := range publicPaths {
		if matchesPublicPathPattern(pattern, c.Path()) {
			return true
		}
	}
	return false
}

func matchesPublicPathPattern(pattern string, path string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}

	if !looksLikeRegexPattern(pattern) {
		return pattern == path
	}

	matched, err := regexp.MatchString(pattern, path)
	return err == nil && matched
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
		return []byte(GetEnv("JWT_SECRET", "")), nil
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

// IsJwtValid returns false if the token is in the Mongo blacklist. When Mongo
// is not initialized (blacklist disabled), it returns true.
func IsJwtValid(token string) bool {
	if Mongo == nil {
		return true
	}
	collection := Mongo.Database().Collection("blacklist_tokens")
	filter := bson.M{"token": token}

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false
	}
	return count == 0
}
