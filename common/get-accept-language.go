package common

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetAcceptLanguage(c *fiber.Ctx) string {
	language := c.Get("accept-language", "TH")
	return strings.ToUpper(language)
}
