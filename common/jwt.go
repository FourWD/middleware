package common

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
)

func AuthenticationMiddleware(c *fiber.Ctx) error {
	if isPublicPath(c) {
		return c.Next()
	}
	return checkAuth(c)
}

func isPublicPath(c *fiber.Ctx) bool {
	publicPaths := viper.GetStringSlice("public_path")
	hardcodePaths := []string{"/_ah/warmup", "/wake-up", "/metrics"}
	publicPaths = append(publicPaths, hardcodePaths...)
	return StringExistsInList(c.Path(), publicPaths)
}

func checkAuth(c *fiber.Ctx) error {
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
		return []byte(viper.GetString("jwt_secret_key")), nil
	})

	if err != nil {
		// Check for specific validation errors
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				return c.Status(http.StatusUnauthorized).SendString("Invalid token signature")
			}
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return c.Status(http.StatusUnauthorized).SendString("Token expired")
			}
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
	collection := DatabaseMongoMiddleware.Database.Collection("blacklist_tokens")
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
