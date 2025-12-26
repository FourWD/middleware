package common

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

func EncodedJwtToken(c *fiber.Ctx, res string) (string, error) {
	authorizeToken := c.Get("Authorization")
	if authorizeToken == "" {
		return "", errors.New("authorize = nil ")

	}
	tokenString := strings.Replace(authorizeToken, "Bearer ", "", 1)
	if tokenString == "" {
		return "", errors.New("token = nil ")

	}
	secretKeyToken := []byte(viper.GetString("jwt_secret_key"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKeyToken, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claimToken, ok := claims[res].(string); ok {
			return claimToken, nil
		}
	}

	return "", err
}

func EncodedJwtTokenExpired(c *fiber.Ctx, res string) (string, error) {
	authorizeToken := c.Get("Authorization")
	if authorizeToken == "" {
		return "", errors.New("authorize = nil ")

	}
	tokenString := strings.Replace(authorizeToken, "Bearer ", "", 1)
	if tokenString == "" {
		return "", errors.New("token = nil ")

	}
	secretKeyToken := []byte(viper.GetString("jwt_secret_key"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKeyToken, nil
	})
	// if err != nil {
	// 	return err.Error(), err
	// }

	if claims, ok := token.Claims.(jwt.MapClaims); ok || token.Valid {
		if claimToken, ok := claims[res].(string); ok {
			return claimToken, nil
		}
	}

	return "", err
}
