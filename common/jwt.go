package common

import (
	"fmt"

	"github.com/FourWD/middleware/infra"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// AuthenticationMiddleware is an alias for infra.AuthenticationMiddleware,
// kept for backwards compatibility. Prefer infra.AuthenticationMiddleware
// directly — NewApp now auto-registers it.
var AuthenticationMiddleware fiber.Handler = infra.AuthenticationMiddleware

// IsJwtValid is an alias for infra.IsJwtValid, kept for backwards compatibility.
func IsJwtValid(token string) bool {
	return infra.IsJwtValid(token)
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
