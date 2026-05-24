package infra

import (
	"bytes"
	"testing"
)

// CaptureLogs swaps the package-level AppLog with a logger that writes to
// the returned buffer for the duration of the test. The original AppLog is
// restored via t.Cleanup so test ordering does not matter.
//
// NOT safe with t.Parallel(): the swap mutates the AppLog global, so two
// parallel tests calling CaptureLogs race on the assignment. For
// parallel-safe scoping, attach the logger to a context with
// infra.WithLogger(ctx, customLogger) and read it back with
// LoggerFromContext — helpers in log_helpers.go follow that pattern.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    buf := infra.CaptureLogs(t)
//	    // ... exercise code that calls infra.AppLog.* ...
//	    require.Contains(t, buf.String(), `"label":"DB_CREATE_FAILURE"`)
//	}
func CaptureLogs(t testing.TB) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	old := AppLog
	AppLog = NewSlogAPILogger(&buf, SlogAPILoggerConfig{LogLevel: "debug"})
	t.Cleanup(func() { AppLog = old })
	return &buf
}
