package infra

import (
	"strings"
	"time"
)

// setupTimezone applies the configured timezone to time.Local.
// Called automatically from NewApp.
func setupTimezone(tz string, logger *Logger) {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		if logger != nil {
			logger.LifecycleError(err, "TIMEZONE_LOAD_FAILURE", map[string]any{
				"timezone": tz,
			}, WithComponent(ComponentApp), WithOperation("setup_timezone"))
		}
		return
	}
	time.Local = loc

	if logger != nil {
		logger.LifecycleEvent("TIMEZONE_SET", map[string]any{
			"timezone": tz,
		}, WithComponent(ComponentApp), WithOperation("setup_timezone"))
	}
}
