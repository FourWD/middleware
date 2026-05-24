package common

import (
	"github.com/FourWD/middleware/infra"
	"github.com/gofiber/fiber/v3"
)

// Deprecated: use infra.GetSessionUserID directly. Wrapper kept so existing
// callers keep compiling; will be removed once downstream projects migrate.
func GetSessionUserID(c fiber.Ctx) string {
	return infra.GetSessionUserID(c)
}
