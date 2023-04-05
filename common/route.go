package common

import (
	"os"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
)

func AuthorizationRequired() fiber.Handler {
	jwtSecret := []byte(os.Getenv("JWT_SECRET_KEY"))
	return jwtware.New(jwtware.Config{
		SigningMethod: "HS256",
		SigningKey:    []byte(jwtSecret),
		SuccessHandler: func(c *fiber.Ctx) error {
			return c.Next()
		},
		ErrorHandler: func(c *fiber.Ctx, e error) error {
			return fiber.ErrUnauthorized
		},
	})
}
