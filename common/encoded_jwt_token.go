package common

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

// extractTokenFromHeader extracts JWT token string from Authorization header
func extractTokenFromHeader(c *fiber.Ctx) (string, error) {
	authorizeToken := c.Get("Authorization")
	if authorizeToken == "" {
		return "", errors.New("authorization header is empty")
	}

	if len(authorizeToken) < 7 || authorizeToken[:7] != "Bearer " {
		return "", errors.New("invalid authorization header format")
	}

	tokenString := authorizeToken[7:]
	if tokenString == "" {
		return "", errors.New("token is empty")
	}

	return tokenString, nil
}

// extractClaimFromToken parses JWT and extracts a claim value.
// If allowExpired is true, it will return the claim even if the token is expired
// (useful for token refresh scenarios).
func extractClaimFromToken(tokenString string, claimKey string, allowExpired bool) (string, error) {
	secretKeyToken := []byte(viper.GetString("jwt_secret_key"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKeyToken, nil
	})

	// Handle parsing errors
	if err != nil {
		// If allowExpired, check if error is only about expiration
		if allowExpired {
			if ve, ok := err.(*jwt.ValidationError); ok && ve.Errors == jwt.ValidationErrorExpired {
				// Token is expired but signature is valid, we can read claims
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if claimValue, ok := claims[claimKey].(string); ok {
						return claimValue, nil
					}
					return "", errors.New("claim not found or not a string")
				}
			}
		}
		return "", err
	}

	// Token is valid - extract claim
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claimValue, ok := claims[claimKey].(string); ok {
			return claimValue, nil
		}
		return "", errors.New("claim not found or not a string")
	}

	return "", errors.New("invalid token claims")
}

// EncodedJwtToken extracts a claim from a valid JWT token
func EncodedJwtToken(c *fiber.Ctx, claimKey string) (string, error) {
	tokenString, err := extractTokenFromHeader(c)
	if err != nil {
		return "", err
	}
	return extractClaimFromToken(tokenString, claimKey, false)
}

// EncodedJwtTokenExpired extracts a claim from JWT token even if expired.
// This is useful for token refresh scenarios where you need user info from expired token.
// WARNING: Only use this for reading claims, NOT for authentication decisions.
func EncodedJwtTokenExpired(c *fiber.Ctx, claimKey string) (string, error) {
	tokenString, err := extractTokenFromHeader(c)
	if err != nil {
		return "", err
	}
	return extractClaimFromToken(tokenString, claimKey, true)
}
