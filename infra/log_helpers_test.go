package infra

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

// withTestLogger attaches a buffer-backed logger to ctx and returns the
// buffer. Parallel-safe (scoped per-test via ctx, no global mutation).
func withTestLogger(t *testing.T) (context.Context, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	logger := NewSlogAPILogger(&buf, SlogAPILoggerConfig{LogLevel: "debug"})
	return WithLogger(context.Background(), logger), &buf
}

func TestLogFirebaseError_LabelAndFields(t *testing.T) {
	ctx, buf := withTestLogger(t)
	LogFirebaseError(ctx, errors.New("boom"), "subscribe", "topic-x")

	out := buf.String()
	assertContainsAll(t, out,
		`"message":"FIREBASE_SUBSCRIBE_FAILURE"`,
		`"component":"firebase"`,
		`"operation":"subscribe"`,
		`"log_kind":"error"`,
		`"resource":"topic-x"`,
	)
}

func TestLogAuthSecurity_NoError_LogsWarn(t *testing.T) {
	ctx, buf := withTestLogger(t)
	LogAuthSecurity(ctx, nil, "read_session", "no claims attached")

	out := buf.String()
	assertContainsAll(t, out,
		`"message":"AUTH_READ_SESSION_FAILURE"`,
		`"log_kind":"security"`,
		`"reason":"no claims attached"`,
	)
	// Warn level (no err) — slog renders log_level as "warn".
	if !strings.Contains(out, `"log_level":"warn"`) {
		t.Fatalf("expected warn log_level when err is nil, got: %s", out)
	}
}

func TestLogAuthSecurity_WithError_LogsError(t *testing.T) {
	ctx, buf := withTestLogger(t)
	LogAuthSecurity(ctx, errors.New("verify failed"), "verify_jwt", "bad signature")

	out := buf.String()
	assertContainsAll(t, out,
		`"message":"AUTH_VERIFY_JWT_FAILURE"`,
		`"log_kind":"security"`,
		`"reason":"bad signature"`,
		`"log_level":"error"`,
	)
}

func TestLogPubSubError_LabelAndFields(t *testing.T) {
	ctx, buf := withTestLogger(t)
	LogPubSubError(ctx, errors.New("upstream down"), "publish", "orders.created")

	out := buf.String()
	assertContainsAll(t, out,
		`"message":"PUBSUB_PUBLISH_FAILURE"`,
		`"component":"pubsub"`,
		`"topic":"orders.created"`,
	)
}

func TestLogStorageError_LabelAndFields(t *testing.T) {
	ctx, buf := withTestLogger(t)
	LogStorageError(ctx, errors.New("403"), "upload", "my-bucket", "uploads/2026/05/file.pdf")

	out := buf.String()
	assertContainsAll(t, out,
		`"message":"STORAGE_UPLOAD_FAILURE"`,
		`"component":"storage"`,
		`"bucket":"my-bucket"`,
		`"path":"uploads/2026/05/file.pdf"`,
	)
}

func TestLogCritical_StampsCriticalSeverity(t *testing.T) {
	ctx, buf := withTestLogger(t)
	LogCritical(ctx, errors.New("invariant violated"), "USER_BALANCE_NEGATIVE",
		ComponentApp, "settle_invoice")

	out := buf.String()
	assertContainsAll(t, out,
		`"message":"USER_BALANCE_NEGATIVE"`,
		`"component":"app"`,
		`"operation":"settle_invoice"`,
		`"severity":"critical"`,
	)
}

func TestLogHelpers_NilLoggerInCtxAreSilent(t *testing.T) {
	// LoggerFromContext returns AppLog when ctx has none. If AppLog is also
	// nil (unusual but possible during early boot or torn-down tests), the
	// helpers must NOT panic — they should silently return.
	prevAppLog := AppLog
	AppLog = nil
	t.Cleanup(func() { AppLog = prevAppLog })

	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("helper panicked when AppLog was nil: %v", r)
		}
	}()

	LogFirebaseError(ctx, errors.New("x"), "test", "")
	LogAuthSecurity(ctx, nil, "test", "")
	LogPubSubError(ctx, errors.New("x"), "test", "")
	LogStorageError(ctx, errors.New("x"), "test", "", "")
	LogCritical(ctx, errors.New("x"), "TEST", ComponentApp, "test")
}

func assertContainsAll(t *testing.T, output string, needles ...string) {
	t.Helper()
	for _, want := range needles {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\n--- got ---\n%s", want, output)
		}
	}
}
