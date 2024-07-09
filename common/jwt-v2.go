package common

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func AuthenticationMiddlewareV2(c *fiber.Ctx) error {
	if isPublicPathV2(c) {
		log.Println("public path")
		return c.Next()
	}
	return checkAuth(c)
}

func isPublicPathV2(c *fiber.Ctx) bool {
	publicPaths := viper.GetStringSlice("public_path")
	log.Println("full_path:", c.Path())
	return StringExistsInList(c.Path(), publicPaths)
}
