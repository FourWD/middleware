package infra

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

// unmarshalLogLine parses the last JSON object written to the capture buffer.
// Each log call emits one JSON object per line; tests can call this after
// exercising one log statement to inspect attributes.
func unmarshalLogLine(t *testing.T, raw string) map[string]any {
	t.Helper()
	line := strings.TrimSpace(raw)
	if idx := strings.LastIndex(line, "\n"); idx >= 0 {
		line = strings.TrimSpace(line[idx+1:])
	}
	var entry map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &entry))
	return entry
}

// --- scrubJSONBody ----------------------------------------------------------

func TestScrubJSONBody_TopLevelPassword(t *testing.T) {
	in := []byte(`{"username":"alice","password":"hunter2"}`)
	out := scrubJSONBody(in)
	var got map[string]any
	require.NoError(t, json.Unmarshal(out, &got))
	require.Equal(t, "alice", got["username"])
	require.Equal(t, redactedFieldValue, got["password"])
}

func TestScrubJSONBody_NestedSensitive(t *testing.T) {
	in := []byte(`{"user":{"id":"u1","token":"secret-jwt"},"meta":{"trace":"abc"}}`)
	out := scrubJSONBody(in)
	var got map[string]any
	require.NoError(t, json.Unmarshal(out, &got))
	user := got["user"].(map[string]any)
	require.Equal(t, "u1", user["id"])
	require.Equal(t, redactedFieldValue, user["token"])
}

func TestScrubJSONBody_ArrayOfObjects(t *testing.T) {
	in := []byte(`{"items":[{"id":1,"secret":"a"},{"id":2,"secret":"b"}]}`)
	out := scrubJSONBody(in)
	var got map[string]any
	require.NoError(t, json.Unmarshal(out, &got))
	items := got["items"].([]any)
	require.Equal(t, redactedFieldValue, items[0].(map[string]any)["secret"])
	require.Equal(t, redactedFieldValue, items[1].(map[string]any)["secret"])
}

func TestScrubJSONBody_NotJSONPassesThrough(t *testing.T) {
	in := []byte(`<xml><password>hunter2</password></xml>`)
	out := scrubJSONBody(in)
	require.Equal(t, in, out)
}

// --- scrubFormBody ----------------------------------------------------------

func TestScrubFormBody_Redacts(t *testing.T) {
	in := []byte("username=alice&password=hunter2&otp=123456")
	out := scrubFormBody(in)
	// url.Values.Encode percent-encodes the redacted marker (`<` and `>`),
	// so parse the output back to compare values as decoded strings.
	got, err := parseForm(out)
	require.NoError(t, err)
	require.Equal(t, "alice", got["username"])
	require.Equal(t, redactedFieldValue, got["password"])
	require.Equal(t, redactedFieldValue, got["otp"])
}

// parseForm returns the first value for each key from a urlencoded body.
func parseForm(body []byte) (map[string]string, error) {
	parsed, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(parsed))
	for k, v := range parsed {
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	return out, nil
}

func TestScrubFormBody_NoSensitivePassThrough(t *testing.T) {
	in := []byte("a=1&b=2")
	out := scrubFormBody(in)
	require.Equal(t, in, out)
}

// --- sanitizeURL ------------------------------------------------------------

func TestSanitizeURL_StripsTokenQuery(t *testing.T) {
	out := sanitizeURL("/api/v1/cb?token=jwt&state=ok&id=42")
	require.NotContains(t, out, "token=jwt")
	require.Contains(t, out, "state=ok")
	require.Contains(t, out, "id=42")
}

func TestSanitizeURL_NoQueryUnchanged(t *testing.T) {
	out := sanitizeURL("/api/v1/users/42")
	require.Equal(t, "/api/v1/users/42", out)
}

func TestSanitizeURL_MultipleSensitive(t *testing.T) {
	out := sanitizeURL("/cb?access_token=a&refresh_token=b&keep=ok")
	require.NotContains(t, out, "access_token=a")
	require.NotContains(t, out, "refresh_token=b")
	require.Contains(t, out, "keep=ok")
}

func TestSanitizeURL_StripsFragment(t *testing.T) {
	out := sanitizeURL("/cb?keep=ok#token=jwt-in-fragment")
	require.NotContains(t, out, "token=jwt-in-fragment")
	require.NotContains(t, out, "#")
	require.Contains(t, out, "keep=ok")
}

// --- capDataPayload ---------------------------------------------------------

func TestCapDataPayload_SmallPassesThrough(t *testing.T) {
	data := map[string]any{"a": 1, "b": "two"}
	out := capDataPayload(data)
	require.Equal(t, data, out)
	_, truncated := out["_truncated"]
	require.False(t, truncated)
}

func TestCapDataPayload_OversizedReplacedWithMarker(t *testing.T) {
	// Build a map with enough keys to trip the entry-count threshold,
	// then stuff one of them with a value large enough to blow the
	// byte cap.
	data := map[string]any{}
	for i := 0; i < capDataPayloadSkipThreshold+1; i++ {
		data[strings.Repeat("k", i+1)] = i
	}
	data["big"] = strings.Repeat("x", maxDataPayloadBytes+1)

	out := capDataPayload(data)
	require.Equal(t, true, out["_truncated"])
	require.GreaterOrEqual(t, out["_size_bytes"].(int), maxDataPayloadBytes)
	keys := out["_keys"].([]string)
	require.Contains(t, keys, "big")
}

// --- LifecycleEvent log_kind override --------------------------------------

func TestLifecycleEvent_DefaultsToLifecycleKind(t *testing.T) {
	buf := CaptureLogs(t)
	AppLog.LifecycleEvent("TEST_DEFAULT", nil, WithComponent("test"), WithOperation("op"))

	entry := unmarshalLogLine(t, buf.String())
	require.Equal(t, LogKindLifecycle, entry["log_kind"])
}

func TestLifecycleEvent_CallerOverridesKind(t *testing.T) {
	buf := CaptureLogs(t)
	AppLog.LifecycleEvent("TEST_OVERRIDE", nil,
		WithComponent("test"),
		WithOperation("op"),
		WithLogKind(LogKindSecurity))

	entry := unmarshalLogLine(t, buf.String())
	require.Equal(t, LogKindSecurity, entry["log_kind"],
		"caller-provided WithLogKind must override the helper's default — bug #1 regression guard")
}

// --- CaptureLogs ------------------------------------------------------------

func TestCaptureLogs_RestoresAppLog(t *testing.T) {
	original := AppLog
	t.Run("scoped", func(t *testing.T) {
		_ = CaptureLogs(t)
		require.NotSame(t, original, AppLog, "AppLog should be swapped inside the test")
	})
	require.Same(t, original, AppLog, "AppLog should be restored after the sub-test")
}

// --- error_code extraction from *AppError ----------------------------------

func TestErrorCtx_ExtractsAppErrorCode(t *testing.T) {
	buf := CaptureLogs(t)
	appErr := NewAppError(404, "USER_NOT_FOUND", "user does not exist")
	AppLog.ErrorCtx(context.Background(), appErr, "TEST_APPERR", nil,
		WithComponent("test"), WithOperation("op"))

	entry := unmarshalLogLine(t, buf.String())
	require.Equal(t, "USER_NOT_FOUND", entry["error_code"])
	require.EqualValues(t, 404, entry["error_status"])
	require.Equal(t, "user does not exist", entry["error"])
}

func TestErrorCtx_NoAppErrorNoCodeField(t *testing.T) {
	buf := CaptureLogs(t)
	plainErr := errString("boom")
	AppLog.ErrorCtx(context.Background(), plainErr, "TEST_PLAIN", nil,
		WithComponent("test"), WithOperation("op"))

	entry := unmarshalLogLine(t, buf.String())
	_, hasCode := entry["error_code"]
	require.False(t, hasCode, "plain errors must not synthesise error_code")
	require.Equal(t, "boom", entry["error"])
}

// errString is a minimal local error type used by TestErrorCtx_NoAppErrorNoCodeField.
type errString string

func (e errString) Error() string { return string(e) }

// codedDomainError implements the ErrorWithCode interface as a service
// repo would: a domain-specific error type that exposes a code without
// depending on infra.AppError.
type codedDomainError struct {
	msg  string
	code string
}

func (e *codedDomainError) Error() string { return e.msg }
func (e *codedDomainError) Code() string  { return e.code }

func TestErrorCtx_ExtractsCustomCodedError(t *testing.T) {
	buf := CaptureLogs(t)
	err := &codedDomainError{msg: "vehicle locked", code: "VEHICLE_LOCKED"}
	AppLog.ErrorCtx(context.Background(), err, "TEST_CODED", nil,
		WithComponent("test"), WithOperation("op"))

	entry := unmarshalLogLine(t, buf.String())
	require.Equal(t, "VEHICLE_LOCKED", entry["error_code"])
	_, hasStatus := entry["error_status"]
	require.False(t, hasStatus, "ErrorWithCode interface has no Status — must not synthesise one")
}

// --- _keys redaction in truncation marker (item #7) ------------------------

func TestCapDataPayload_RedactsSensitiveKeyNames(t *testing.T) {
	data := map[string]any{}
	for i := 0; i < capDataPayloadSkipThreshold+1; i++ {
		data["k"+strings.Repeat("x", i+1)] = i
	}
	data["password"] = "irrelevant"
	data["big"] = strings.Repeat("x", maxDataPayloadBytes+1)

	out := capDataPayload(data)
	require.Equal(t, true, out["_truncated"])
	keys := out["_keys"].([]string)
	require.NotContains(t, keys, "password", "sensitive field names must not appear in _keys")
	require.Contains(t, keys, "<redacted>", "redaction marker must appear in place of sensitive keys")
	require.Contains(t, keys, "big")
}

// --- WithLogger / LoggerFromContext ----------------------------------------

func TestLoggerFromContext_PrefersScoped(t *testing.T) {
	scoped := NewSlogAPILogger(nil, SlogAPILoggerConfig{LogLevel: "debug"})
	ctx := WithLogger(context.Background(), scoped)
	require.Same(t, scoped, LoggerFromContext(ctx))
}

func TestLoggerFromContext_FallsBackToAppLog(t *testing.T) {
	require.Same(t, AppLog, LoggerFromContext(context.Background()))
}

// --- Helper functions (log_helpers.go) -------------------------------------
//
// These tests double as the regression guard for Bug #2: if a helper drops
// its WithCallerSkip, the `source` field starts pointing at log_helpers.go
// instead of the test file and the assertion below fails.

func TestLogDBError_TaxonomyAndSource(t *testing.T) {
	buf := CaptureLogs(t)
	LogDBError(context.Background(), errString("boom"), "create", "users", "u1")

	entry := unmarshalLogLine(t, buf.String())
	require.Equal(t, "DB_CREATE_FAILURE", entry["message"])
	require.Equal(t, ComponentDB, entry["component"])
	require.Equal(t, "create", entry["operation"])
	require.Equal(t, LogKindError, entry["log_kind"])

	data := entry["data"].(map[string]any)
	require.Equal(t, "users", data["table"])
	require.Equal(t, "u1", data["record_id"])

	source := entry["source"].(string)
	require.Contains(t, source, "log_test.go",
		"source must point to caller — caller-skip regression guard (Bug #2)")
	require.NotContains(t, source, "log_helpers.go",
		"source must NOT point to helper file")
}

func TestLogHTTPClientError_TaxonomyAndSource(t *testing.T) {
	buf := CaptureLogs(t)
	LogHTTPClientError(context.Background(), errString("timeout"), "status", "POST", "https://api.example.com/x", 503)

	entry := unmarshalLogLine(t, buf.String())
	require.Equal(t, "HTTP_STATUS_FAILURE", entry["message"])
	require.Equal(t, ComponentHTTPClient, entry["component"])
	require.Equal(t, "post", entry["operation"])
	require.Equal(t, LogKindError, entry["log_kind"])
	// "status" is a topLevelLogField, so it surfaces alongside component
	// rather than nesting under data.
	require.EqualValues(t, 503, entry["status"])

	data := entry["data"].(map[string]any)
	require.Equal(t, "https://api.example.com/x", data["url"])
	require.Equal(t, "status", data["stage"])

	require.Contains(t, entry["source"].(string), "log_test.go")
}

func TestLogBusinessEvent_TaxonomyAndSource(t *testing.T) {
	buf := CaptureLogs(t)
	LogBusinessEvent(context.Background(), "USER_CREATED", "handler", "create_user",
		map[string]any{"user_id": "u1"})

	entry := unmarshalLogLine(t, buf.String())
	require.Equal(t, "USER_CREATED", entry["message"])
	require.Equal(t, "handler", entry["component"])
	require.Equal(t, "create_user", entry["operation"])
	require.Equal(t, LogKindBusiness, entry["log_kind"])

	data := entry["data"].(map[string]any)
	require.Equal(t, "u1", data["user_id"])

	require.Contains(t, entry["source"].(string), "log_test.go")
}

func TestLogLifecycle_TaxonomyAndSource(t *testing.T) {
	buf := CaptureLogs(t)
	LogLifecycle("WORKER_STARTED", ComponentApp, "worker_start",
		map[string]any{"worker": "consumer-1"})

	entry := unmarshalLogLine(t, buf.String())
	require.Equal(t, "WORKER_STARTED", entry["message"])
	require.Equal(t, ComponentApp, entry["component"])
	require.Equal(t, "worker_start", entry["operation"])
	require.Equal(t, LogKindLifecycle, entry["log_kind"])

	data := entry["data"].(map[string]any)
	require.Equal(t, "consumer-1", data["worker"])

	require.Contains(t, entry["source"].(string), "log_test.go")
}

// --- request_log sampling --------------------------------------------------

// countingRequestLogger counts Log calls without emitting; lets sampling
// tests assert how many log entries the middleware produced for a given
// SamplingRate configuration.
type countingRequestLogger struct {
	count int
}

func (c *countingRequestLogger) Log(ctx context.Context, entry Entry, options ...RequestLoggerOption) {
	c.count++
}

func runRequest(t *testing.T, app *fiber.App, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		_, _ = app.Test(req)
	}
}

func TestRequestLog_SamplingDropsSuccess(t *testing.T) {
	rl := &countingRequestLogger{}
	app := fiber.New()
	app.Use(NewRequestLog(RequestLogConfig{
		RequestLogger:       rl,
		SamplingRate:        0.0001,
		AlwaysLogStatusFrom: fiber.StatusBadRequest,
	}))
	app.Get("/", func(c fiber.Ctx) error { return c.SendStatus(200) })

	runRequest(t, app, 50)
	// With 0.0001 sampling, the expected count is ~0 out of 50 — allow up
	// to a couple of false positives for the random source.
	require.LessOrEqual(t, rl.count, 2,
		"SamplingRate≈0 must drop nearly all 2xx responses; got %d", rl.count)
}

func TestRequestLog_AlwaysLogErrorStatus(t *testing.T) {
	rl := &countingRequestLogger{}
	app := fiber.New()
	app.Use(NewRequestLog(RequestLogConfig{
		RequestLogger:       rl,
		SamplingRate:        0.0001, // would drop 2xx
		AlwaysLogStatusFrom: fiber.StatusBadRequest,
	}))
	app.Get("/", func(c fiber.Ctx) error { return c.SendStatus(500) })

	runRequest(t, app, 20)
	require.Equal(t, 20, rl.count,
		"5xx must bypass sampling and always be logged")
}

func TestRequestLog_DefaultLogsEverything(t *testing.T) {
	rl := &countingRequestLogger{}
	app := fiber.New()
	app.Use(NewRequestLog(RequestLogConfig{
		RequestLogger: rl,
		// SamplingRate left at 0 — fast-path "log everything"
	}))
	app.Get("/", func(c fiber.Ctx) error { return c.SendStatus(200) })

	runRequest(t, app, 25)
	require.Equal(t, 25, rl.count,
		"SamplingRate=0 (default) must log every request")
}

// --- sanitizeHeaders -------------------------------------------------------

func TestSanitizeHeaders_RedactsSensitive(t *testing.T) {
	headers := map[string][]string{
		"Authorization": {"Bearer secret-jwt"},
		"Cookie":        {"session=abc"},
		"X-Api-Key":     {"abcdef"},
		"X-Request-Id":  {"req-123"},
		"Content-Type":  {"application/json"},
	}
	result := sanitizeHeaders(headers)

	require.Equal(t, []string{redactedFieldValue}, result["Authorization"])
	require.Equal(t, []string{redactedFieldValue}, result["Cookie"])
	require.Equal(t, []string{redactedFieldValue}, result["X-Api-Key"])
	// Non-sensitive headers must pass through unchanged.
	require.Equal(t, []string{"req-123"}, result["X-Request-Id"])
	require.Equal(t, []string{"application/json"}, result["Content-Type"])
}

func TestSanitizeHeaders_CaseInsensitive(t *testing.T) {
	headers := map[string][]string{
		"AUTHORIZATION": {"Bearer x"},
		"authorization": {"Bearer y"},
	}
	result := sanitizeHeaders(headers)
	require.Equal(t, []string{redactedFieldValue}, result["AUTHORIZATION"])
	require.Equal(t, []string{redactedFieldValue}, result["authorization"])
}

func TestSanitizeHeaders_EmptyInputReturnsNil(t *testing.T) {
	require.Nil(t, sanitizeHeaders(nil))
	require.Nil(t, sanitizeHeaders(map[string][]string{}))
}

// --- bodyString ------------------------------------------------------------

func TestBodyString_OmitFlagReplacesBody(t *testing.T) {
	out := bodyString([]byte(`{"k":"v"}`), "application/json", true, 4096)
	require.Equal(t, omittedBodyValue, out)
}

func TestBodyString_EmptyBodyReturnsEmpty(t *testing.T) {
	out := bodyString(nil, "application/json", false, 4096)
	require.Equal(t, "", out)
}

func TestBodyString_NonLoggableContentTypeOmits(t *testing.T) {
	out := bodyString([]byte("binary..."), "application/octet-stream", false, 4096)
	require.Equal(t, omittedBodyValue, out)
}

func TestBodyString_TruncatesBeyondMax(t *testing.T) {
	body := []byte(`{"data":"` + strings.Repeat("x", 100) + `"}`)
	out := bodyString(body, "application/json", false, 20)
	require.True(t, strings.HasSuffix(out, "...<truncated>"),
		"oversize body must be truncated with the marker suffix")
	require.LessOrEqual(t, len(out), 20+len("...<truncated>"))
}

func TestBodyString_ScrubsJSONBeforeReturn(t *testing.T) {
	body := []byte(`{"username":"alice","password":"hunter2"}`)
	out := bodyString(body, "application/json", false, 4096)
	require.NotContains(t, out, "hunter2",
		"password value must be scrubbed before becoming a log field")
	// json.Marshal HTML-escapes "<" / ">" — parse the JSON back and
	// compare values rather than substring-matching the encoded form.
	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	require.Equal(t, redactedFieldValue, got["password"])
	require.Equal(t, "alice", got["username"])
}

// --- End-to-end integration: full request → middleware → log line ---------
//
// Asserts the complete pipeline produces a log entry with the canonical
// shape: required taxonomy fields, request_id, scrubbed body, correct
// severity by status. If anything in the middleware composition or log
// pipeline breaks, this test catches it where unit tests would not.

func TestRequestLog_EndToEnd_LoginRequest(t *testing.T) {
	buf := CaptureLogs(t)

	app := fiber.New()
	app.Use(NewRequestID())
	app.Use(NewRequestLog(RequestLogConfig{
		RequestLogger:   NewSlogRequestLogger(AppLog),
		OmitRequestBody: false, // capture body so we can assert it was scrubbed
	}))
	app.Post("/api/login", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	reqBody := `{"username":"alice","password":"hunter2","email":"alice@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login?next=/home&token=secret", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer leak-me-if-you-can")

	_, err := app.Test(req)
	require.NoError(t, err)

	entry := unmarshalLogLine(t, buf.String())

	// Top-level taxonomy + request shape.
	require.Equal(t, ComponentHTTP, entry["component"])
	require.Equal(t, "request", entry["operation"])
	require.Equal(t, LogKindRequest, entry["log_kind"])
	require.Equal(t, "POST", entry["method"])
	require.Equal(t, "/api/login", entry["route_name"])
	require.EqualValues(t, fiber.StatusOK, entry["status"])
	require.NotEmpty(t, entry["request_id"], "RequestID middleware must populate request_id")

	// data sub-object carries the body / headers / URI fields.
	data := entry["data"].(map[string]any)

	// raw_uri keeps safe query param, strips sensitive ones.
	rawURI := data["raw_uri"].(string)
	require.Contains(t, rawURI, "next=")
	require.NotContains(t, rawURI, "token=secret",
		"sensitive query keys must be stripped from raw_uri")

	// Request body — password redacted, harmless fields preserved.
	var parsedBody map[string]any
	require.NoError(t, json.Unmarshal([]byte(data["req_body"].(string)), &parsedBody))
	require.Equal(t, "alice", parsedBody["username"])
	require.Equal(t, redactedFieldValue, parsedBody["password"],
		"password must be redacted in req_body")
	require.Equal(t, "alice@example.com", parsedBody["email"])

	// Request headers — Authorization sanitised.
	var reqHeaders map[string][]string
	require.NoError(t, json.Unmarshal([]byte(data["req_header"].(string)), &reqHeaders))
	if auth, ok := reqHeaders["Authorization"]; ok {
		require.Equal(t, []string{redactedFieldValue}, auth,
			"Authorization header must be sanitised in req_header")
	}
}

func TestRequestLog_EndToEnd_ErrorStatusUsesErrorLevel(t *testing.T) {
	buf := CaptureLogs(t)

	app := fiber.New()
	app.Use(NewRequestLog(RequestLogConfig{
		RequestLogger: NewSlogRequestLogger(AppLog),
	}))
	app.Get("/boom", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	_, err := app.Test(req)
	require.NoError(t, err)

	entry := unmarshalLogLine(t, buf.String())
	require.Equal(t, "error", entry["log_level"],
		"5xx responses must emit at log_level=error")
	require.EqualValues(t, fiber.StatusInternalServerError, entry["status"])
}

