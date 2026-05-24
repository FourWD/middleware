package infra

import (
	"context"
	"io"
	"strings"
	"testing"
)

// silentLogger is a Logger that drops all output. Used by benchmarks so
// the timing measures the logger pipeline, not the test output writer.
func silentLogger() *Logger {
	return NewSlogAPILogger(io.Discard, SlogAPILoggerConfig{LogLevel: "debug"})
}

// BenchmarkLifecycleEvent measures the hot-path cost of emitting a typical
// info-level lifecycle log entry (one component + one operation field).
func BenchmarkLifecycleEvent(b *testing.B) {
	logger := silentLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LifecycleEvent("BENCH_LIFECYCLE", nil,
			WithComponent(ComponentApp),
			WithOperation("benchmark"))
	}
}

// BenchmarkEventCtx measures the hot-path cost of a request-scoped business
// event with a small data payload (the typical handler-emitted log shape).
func BenchmarkEventCtx(b *testing.B) {
	logger := silentLogger()
	ctx := InjectCorrelationID(context.Background(), "req-bench-id")
	data := map[string]any{"user_id": "u-1", "action": "create"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.EventCtx(ctx, "BENCH_EVENT", data,
			WithComponent(ComponentHandler),
			WithOperation("benchmark"))
	}
}

// BenchmarkErrorCtxWithAppError exercises the error path including
// stack-trace capture and error_code extraction.
func BenchmarkErrorCtxWithAppError(b *testing.B) {
	logger := silentLogger()
	err := NewAppError(404, "NOT_FOUND", "resource missing")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.ErrorCtx(context.Background(), err, "BENCH_ERROR", nil,
			WithComponent(ComponentHandler),
			WithOperation("benchmark"))
	}
}

// BenchmarkScrubJSONBody_Small measures the scrubber on a tiny payload —
// the common case for typical API requests.
func BenchmarkScrubJSONBody_Small(b *testing.B) {
	body := []byte(`{"username":"alice","password":"hunter2","email":"a@b.com"}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scrubJSONBody(body)
	}
}

// BenchmarkScrubJSONBody_Nested measures the scrubber on a payload with
// nested objects and arrays — the worst case for recursive walk.
func BenchmarkScrubJSONBody_Nested(b *testing.B) {
	body := []byte(`{"user":{"id":"u1","creds":{"password":"x","token":"y"}},"items":[{"secret":"a"},{"secret":"b"}]}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scrubJSONBody(body)
	}
}

// BenchmarkScrubFormBody measures the urlencoded scrubber on a login form.
func BenchmarkScrubFormBody(b *testing.B) {
	body := []byte("username=alice&password=hunter2&keep_signed_in=true&csrf=tok")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scrubFormBody(body)
	}
}

// BenchmarkSanitizeURL_NoSensitive measures the URL scrubber on the common
// case where nothing needs stripping (early-return path).
func BenchmarkSanitizeURL_NoSensitive(b *testing.B) {
	url := "/api/v1/users/42?page=3&size=20"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizeURL(url)
	}
}

// BenchmarkSanitizeURL_StripsToken measures the path where the query has
// to be re-encoded.
func BenchmarkSanitizeURL_StripsToken(b *testing.B) {
	url := "/cb?access_token=abc.def.ghi&state=ok&id=42"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizeURL(url)
	}
}

// BenchmarkCapDataPayload_SmallSkip exercises the fast-path skip in
// capDataPayload when the map is below the entry-count threshold.
func BenchmarkCapDataPayload_SmallSkip(b *testing.B) {
	data := map[string]any{"a": 1, "b": "two", "c": true}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = capDataPayload(data)
	}
}

// BenchmarkCapDataPayload_Large exercises the marshal-and-truncate path.
func BenchmarkCapDataPayload_Large(b *testing.B) {
	data := map[string]any{}
	for i := 0; i < 50; i++ {
		data["field_"+strings.Repeat("x", i+1)] = strings.Repeat("y", 100)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = capDataPayload(data)
	}
}
