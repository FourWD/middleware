package infra

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type captureRequestLogger struct {
	entries []Entry
}

func (l *captureRequestLogger) Log(_ context.Context, entry Entry, _ ...RequestLoggerOption) {
	l.entries = append(l.entries, entry)
}

// The request log must record the status the client receives, including
// when the handler returns an error and the status is only written later by
// the app ErrorHandler.
func TestRequestLogStatusMatchesClient(t *testing.T) {
	testCases := []struct {
		name string
		path string
		want int
	}{
		{name: "ok", path: "/ok", want: http.StatusOK},
		{name: "unmatched route", path: "/.env", want: http.StatusNotFound},
		{name: "fiber error", path: "/fiber-error", want: http.StatusBadGateway},
		{name: "app error", path: "/app-error", want: http.StatusConflict},
		{name: "plain error", path: "/plain-error", want: http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			logger := &captureRequestLogger{}
			app := fiber.New(fiber.Config{ErrorHandler: AppErrorHandler()})
			app.Use(NewRequestLog(RequestLogConfig{RequestLogger: logger}))
			app.Get("/ok", func(c fiber.Ctx) error { return c.SendString("ok") })
			app.Get("/fiber-error", func(c fiber.Ctx) error { return fiber.NewError(fiber.StatusBadGateway, "upstream") })
			app.Get("/app-error", func(c fiber.Ctx) error { return NewAppError(fiber.StatusConflict, "conflict", "conflict") })
			app.Get("/plain-error", func(c fiber.Ctx) error { return errors.New("boom") })

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, tc.path, nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.want {
				t.Fatalf("client got %d, want %d", resp.StatusCode, tc.want)
			}
			if len(logger.entries) != 1 {
				t.Fatalf("logged %d entries, want 1", len(logger.entries))
			}
			if got := logger.entries[0].Response.Status; got != tc.want {
				t.Fatalf("request log recorded %d, client got %d", got, tc.want)
			}
		})
	}
}
