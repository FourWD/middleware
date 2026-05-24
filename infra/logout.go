package infra

import (
	"github.com/gofiber/fiber/v3"
)

func Logout(c fiber.Ctx) error {
	return BlacklistJwtToken(c.Context(), c.Get("Authorization"))
}
