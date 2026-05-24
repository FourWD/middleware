package infra

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const bearerPrefix = "Bearer "

// GenerateJWTToken signs a JWTClaims envelope with HS256 and JWT_SECRET.
// The signing key is read once via the cached jwtSecret() loader.
func GenerateJWTToken(userID string, role string, remark map[string]string, expiresIn time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		Remark: remark,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

func extractTokenFromHeader(c fiber.Ctx) (string, error) {
	header := c.Get("Authorization")
	if header == "" {
		return "", errors.New("authorization header is empty")
	}
	if !strings.HasPrefix(header, bearerPrefix) {
		return "", errors.New("invalid authorization header format")
	}
	token := header[len(bearerPrefix):]
	if token == "" {
		return "", errors.New("token is empty")
	}
	return token, nil
}

// extractClaimFromToken parses a JWT and returns a string claim. When
// allowExpired is true the claim is still returned for an otherwise-valid
// token whose only failure was expiry — used by the token-refresh path.
func extractClaimFromToken(tokenString string, claimKey string, allowExpired bool) (string, error) {
	secret := jwtSecret()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		if allowExpired && errors.Is(err, jwt.ErrTokenExpired) {
			return claimAsString(token, claimKey)
		}
		return "", err
	}
	if !token.Valid {
		return "", errors.New("invalid token claims")
	}
	return claimAsString(token, claimKey)
}

func claimAsString(token *jwt.Token, claimKey string) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	value, ok := claims[claimKey].(string)
	if !ok {
		return "", errors.New("claim not found or not a string")
	}
	return value, nil
}

// EncodedJwtToken extracts a claim from a valid JWT token in the Bearer header.
func EncodedJwtToken(c fiber.Ctx, claimKey string) (string, error) {
	tokenString, err := extractTokenFromHeader(c)
	if err != nil {
		return "", err
	}
	return extractClaimFromToken(tokenString, claimKey, false)
}

// EncodedJwtTokenExpired extracts a claim from JWT token even if expired.
// This is useful for token refresh scenarios where you need user info from
// an expired token.
// WARNING: Only use this for reading claims, NOT for authentication decisions.
func EncodedJwtTokenExpired(c fiber.Ctx, claimKey string) (string, error) {
	tokenString, err := extractTokenFromHeader(c)
	if err != nil {
		return "", err
	}
	return extractClaimFromToken(tokenString, claimKey, true)
}
