package common

import (
	"github.com/gofiber/fiber/v3"
)

func Logout(c fiber.Ctx) error {
	jwtToken := c.Get("Authorization")
	return BlacklistJwtToken(jwtToken)
}
