package infra

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func routePath(c fiber.Ctx) string {
	if fullPath := c.FullPath(); fullPath != "" {
		return fullPath
	}

	return c.Path()
}

func statusCodeClass(status int) string {
	if status < 100 || status > 999 {
		return "unknown"
	}
	return strconv.Itoa(status/100) + "xx"
}
