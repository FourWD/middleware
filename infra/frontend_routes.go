package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// defaultFrontendSkipPrefixes are the path prefixes the SPA fallback must never
// swallow: the API surface plus every route mounted by registerDefaultRoutes and
// RegisterStandardRoutes.
var defaultFrontendSkipPrefixes = []string{
	"/api/",
	"/_ah/",
	"/debug/",
	"/wake-up",
	"/livez",
	"/healthz",
	"/readyz",
	"/metrics",
}

// FrontendConfig configures static asset serving for a bundled SPA.
type FrontendConfig struct {
	// DistPath is the directory holding the built frontend. Must contain index.html.
	DistPath string

	// SkipPrefixes are path prefixes passed through to the API and infra routes
	// instead of falling back to index.html.
	// Defaults to defaultFrontendSkipPrefixes when empty.
	SkipPrefixes []string

	// IndexName overrides the SPA entrypoint filename. Defaults to "index.html".
	IndexName string
}

// RegisterFrontendRoutes serves the built frontend from cfg.DistPath and falls
// back to index.html for any unmatched GET/HEAD path so client-side routing works
// on deep links and hard refreshes.
//
// MUST be registered last — after every API and infra route. Fiber matches in
// registration order, so a fallback mounted early would shadow routes added after it.
//
// Returns an error at boot when the entrypoint is missing, rather than failing
// on the first request.
func RegisterFrontendRoutes(app *fiber.App, cfg FrontendConfig) error {
	distPath := strings.TrimSpace(cfg.DistPath)
	if distPath == "" {
		return fmt.Errorf("frontend dist path is empty")
	}

	indexName := strings.TrimSpace(cfg.IndexName)
	if indexName == "" {
		indexName = "index.html"
	}

	indexPath := filepath.Join(distPath, indexName)
	if _, err := os.Stat(indexPath); err != nil {
		return fmt.Errorf("frontend entrypoint %s: %w", indexPath, err)
	}

	skips := cfg.SkipPrefixes
	if len(skips) == 0 {
		skips = defaultFrontendSkipPrefixes
	}

	skip := func(c fiber.Ctx) bool {
		path := c.Path()
		for _, prefix := range skips {
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
		return false
	}

	app.Use(static.New(distPath, static.Config{
		IndexNames: []string{indexName},
		Next:       skip,
	}))

	app.Use(func(c fiber.Ctx) error {
		if c.Method() != fiber.MethodGet && c.Method() != fiber.MethodHead {
			return c.Next()
		}
		if skip(c) {
			return c.Next()
		}
		return c.Status(fiber.StatusOK).SendFile(indexPath)
	})

	return nil
}
