package infra

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// --- newPathMatcher ---------------------------------------------------------

func TestNewPathMatcher_ExactMatch(t *testing.T) {
	m := newPathMatcher([]string{"/api/v1/sync-service/sync-md2db/cron"})

	if !m.matches("/api/v1/sync-service/sync-md2db/cron") {
		t.Error("exact path should match")
	}
	if m.matches("/api/v1/sync-service/sync-md2db/cron/extra") {
		t.Error("exact match must not behave as a prefix")
	}
}

func TestNewPathMatcher_RegexPattern(t *testing.T) {
	m := newPathMatcher([]string{`^/api/v1/.*/cron$`})

	if !m.matches("/api/v1/clean-service/clean-model/cron") {
		t.Error("regex should match cron route")
	}
	if m.matches("/api/v1/clean-service/clean-model") {
		t.Error("regex should not match non-cron route")
	}
}

func TestNewPathMatcher_SkipsBlankAndInvalidPatterns(t *testing.T) {
	// "[" is an invalid regex — dropped rather than panicking or failing boot.
	m := newPathMatcher([]string{"", "  ", "[", "/ok"})

	if !m.matches("/ok") {
		t.Error("valid pattern should survive alongside invalid ones")
	}
	if m.matches("") {
		t.Error("blank patterns must not produce a match-all")
	}
}

// --- CRON_PATHS -------------------------------------------------------------

// withCronPaths swaps the package-level matchers for the duration of the test.
// Both are sync.OnceValue, so t.Setenv cannot reach them once initialised;
// publicPaths is pinned to the baseline so an unrelated test's HTTP_PUBLIC_PATHS
// cannot leak in and mask the cron branch.
func withCronPaths(t *testing.T, patterns ...string) {
	t.Helper()
	originalCron, originalPublic := cronPaths, publicPaths
	cronPaths = func() *publicPathMatcher { return newPathMatcher(patterns) }
	publicPaths = func() *publicPathMatcher { return newPathMatcher(baselinePublicPaths) }
	t.Cleanup(func() { cronPaths, publicPaths = originalCron, originalPublic })
}

// runAuthMiddleware drives AuthenticationMiddleware over a bare fiber app and
// returns the status code. headers are applied verbatim.
func runAuthMiddleware(t *testing.T, path string, headers map[string]string) int {
	t.Helper()
	app := fiber.New()
	app.Use(AuthenticationMiddleware)
	app.Get("/*", func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return resp.StatusCode
}

func TestAuthenticationMiddleware_CronPathWithCronHeaderSkipsAuth(t *testing.T) {
	withCronPaths(t, "/api/v1/sync-service/sync-md2db/cron")

	got := runAuthMiddleware(t, "/api/v1/sync-service/sync-md2db/cron", map[string]string{
		appEngineCronHeader: "true",
	})
	if got != http.StatusOK {
		t.Fatalf("got %d want 200", got)
	}
}

func TestAuthenticationMiddleware_CronPathWithoutHeaderRequiresToken(t *testing.T) {
	withCronPaths(t, "/api/v1/sync-service/sync-md2db/cron")

	got := runAuthMiddleware(t, "/api/v1/sync-service/sync-md2db/cron", nil)
	if got != http.StatusUnauthorized {
		t.Fatalf("got %d want 401 — a listed path alone must not skip auth", got)
	}
}

func TestAuthenticationMiddleware_CronHeaderOnUnlistedPathRequiresToken(t *testing.T) {
	withCronPaths(t, "/api/v1/sync-service/sync-md2db/cron")

	got := runAuthMiddleware(t, "/api/v1/users", map[string]string{
		appEngineCronHeader: "true",
	})
	if got != http.StatusUnauthorized {
		t.Fatalf("got %d want 401 — the header must not widen access beyond CRON_PATHS", got)
	}
}

func TestAuthenticationMiddleware_CronHeaderValueMustBeExactlyTrue(t *testing.T) {
	withCronPaths(t, "/cron")

	// "true " is absent on purpose: HTTP strips optional trailing whitespace
	// from header values, so it arrives as "true" and legitimately passes.
	for _, value := range []string{"TRUE", "1", "yes", "truthy", ""} {
		got := runAuthMiddleware(t, "/cron", map[string]string{appEngineCronHeader: value})
		if got != http.StatusUnauthorized {
			t.Errorf("header %q: got %d want 401", value, got)
		}
	}
}

func TestAuthenticationMiddleware_EmptyCronPathsDisablesBypass(t *testing.T) {
	withCronPaths(t) // CRON_PATHS unset

	got := runAuthMiddleware(t, "/cron", map[string]string{appEngineCronHeader: "true"})
	if got != http.StatusUnauthorized {
		t.Fatalf("got %d want 401 — bypass must stay off when CRON_PATHS is empty", got)
	}
}
