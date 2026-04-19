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
			logger.Error(err, M("load timezone failed"),
				WithField("timezone", tz),
				WithComponent("app"),
				WithOperation("setup_timezone"),
				WithLogKind("startup"))
		}
		return
	}
	time.Local = loc

	if logger != nil {
		logger.Info(M("timezone set"),
			WithField("timezone", tz),
			WithComponent("app"),
			WithOperation("setup_timezone"),
			WithLogKind("startup"))
	}
}
