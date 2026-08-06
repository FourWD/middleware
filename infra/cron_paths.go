package infra

import (
	"sync"

	"github.com/gofiber/fiber/v3"
)

const appEngineCronHeader = "X-Appengine-Cron"

var cronPaths = sync.OnceValue(func() *publicPathMatcher {
	return newPathMatcher(CronPathsFromEnv())
})

func CronPathsFromEnv() []string {
	return SplitCSV(GetEnv("CRON_PATHS", ""))
}

func isAppEngineCronRequest(c fiber.Ctx, paths *publicPathMatcher) bool {
	if c.Get(appEngineCronHeader) != "true" {
		return false
	}
	return paths.matches(c.Path())
}
