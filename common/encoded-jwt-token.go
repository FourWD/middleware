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
		return err.Error(), err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claimToken, ok := claims[res].(string); ok {
			return claimToken, nil
		}
	}

	return "", err
}

// func GenJwtToken(data map[string]interface{}) (string, error) {
// 	secretKeyToken := []byte(viper.GetString("jwt_secret_key"))
// 	// Add additional claims
// 	data["expire_date"] = time.Now().Add(time.Minute * 15).Unix()

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(data))
// 	tokenString, err := token.SignedString(secretKeyToken)
// 	if err != nil {
// 		return "", err
// 	}

// 	return tokenString, nil
// }

// func AuthorizationRequired() fiber.Handler {
// 	secretKeyToken := []byte(viper.GetString("jwt_secret_key"))
// 	return jwtware.New(jwtware.Config{
// 		SigningMethod: "HS256",
// 		SigningKey:    []byte(secretKeyToken),
// 		SuccessHandler: func(c *fiber.Ctx) error {
// 			return c.Next()
// 		},
// 		ErrorHandler: func(c *fiber.Ctx, e error) error {
// 			return fiber.ErrUnauthorized
// 		},
// 	})
// }

/*
func main() {
	data := map[string]interface{}{
		"user_id":     123,
		"user_role":   "admin",
		"first_name":  "John",
		"last_name":   "Doe",
	}

	token, err := GenJwtToken(data)
	if err != nil {
		panic("Failed to generate JWT token")
	}

	// Use the generated token as needed
	println(token)
}*/
