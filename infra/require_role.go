package infra

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

type DenyHandler func(c fiber.Ctx) error

func NewRequireRole(deny DenyHandler, roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		currentRole := GetAuthRole(c)
		for _, role := range roles {
			if strings.EqualFold(currentRole, role) {
				return c.Next()
			}
		}

		return deny(c)
	}
}
