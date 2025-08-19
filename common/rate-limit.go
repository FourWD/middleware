package common

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/spf13/viper"
)

func isRateLimitPath(c *fiber.Ctx) bool {
	publicPaths := viper.GetStringSlice("rate_limit_path")
	return StringExistsInList(c.Path(), publicPaths)
}

func RateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        viper.GetInt("rate_limit_per_second"),
		Expiration: 1 * time.Second,

		Next: func(c *fiber.Ctx) bool {
			return isRateLimitPath(c)
		},

		KeyGenerator: func(c *fiber.Ctx) string {
			auth := c.Get("Authorization")
			if auth == "" {
				return c.IP()
			}
			return auth
		},
		LimitReached: func(c *fiber.Ctx) error {
			response := map[string]interface{}{
				"status":  0,
				"message": "rate limit exceeded for this token",
			}

			return FiberCustom(c, fiber.StatusTooManyRequests, response)
		},
	})
}
