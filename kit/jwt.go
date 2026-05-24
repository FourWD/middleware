package kit

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// DecodeJWT parses an HS-signed JWT and returns its claims as a generic map.
// secret must match the signing key used to issue the token.
func DecodeJWT(token string, secret string) (map[string]any, error) {
	claims := make(map[string]any)

	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return claims, err
	}

	if mapClaims, ok := parsed.Claims.(jwt.MapClaims); ok && parsed.Valid {
		for k, v := range mapClaims {
			claims[k] = v
		}
		return claims, nil
	}

	return claims, err
}
