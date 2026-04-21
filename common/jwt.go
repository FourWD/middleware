package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/FourWD/middleware/infra"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func AuthenticationMiddleware(c fiber.Ctx) error {
	if isPublicPath(c) {
		return c.Next()
	}
	return checkAuth(c)
}

func isPublicPath(c fiber.Ctx) bool {
	publicPaths := infra.SplitCSV(infra.GetEnv("HTTP_PUBLIC_PATHS", ""))
	hardcodePaths := []string{"/_ah/warmup", "/wake-up", "/metrics"}
	publicPaths = append(publicPaths, hardcodePaths...)
	// return kit.StringExistsInList(c.Path(), publicPaths)
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
	// Extract token from the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(http.StatusUnauthorized).SendString("No token provided")
	}

	// Check Blacklist
	if !IsJwtValid(authHeader) {
		return c.Status(http.StatusUnauthorized).SendString("token blacklist")
	}

	// Ensure Bearer prefix exists
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(http.StatusUnauthorized).SendString("Invalid authorization header format")
	}

	tokenString := authHeader[7:]

	// Parse the token
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(infra.GetEnv("JWT_SECRET", "")), nil
	})

	if err != nil {
		// Check for specific validation errors
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

	// Token is valid, store claims in context
	c.Locals("user", claims)
	return c.Next()
}

func IsJwtValid(token string) bool {
	if infra.Mongo == nil {
		return true
	}
	collection := infra.Mongo.Database().Collection("blacklist_tokens")
	filter := bson.M{"token": token}

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false
	}

	if count > 0 {
		return false
	}

	return true
}

func DecodeJWT(ResponseJwt string, tokenString string) (map[string]interface{}, error) {
	customClaims := make(map[string]interface{})

	token, err := jwt.Parse(ResponseJwt, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenString), nil
	})

	if err != nil {
		return customClaims, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		for key, value := range claims {
			customClaims[key] = value
		}
		return customClaims, nil
	}

	return customClaims, err
}
