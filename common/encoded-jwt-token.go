package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
)

func EncodedJwtToken(req string, res string) (string, error) {
	tokenString := strings.Replace(req, "Bearer ", "", 1)
	secretKeyToken := []byte(os.Getenv("secretKey"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKeyToken, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		claimToken := claims[res].(string)
		return claimToken, nil
	}

	return "", err
}
