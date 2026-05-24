package infra

import (
	"context"
	"testing"
	"time"
)

func testLogger() *Logger {
	return NewLoggerWith(CommonConfig{AppID: "test", AppEnv: "test", LogLevel: "info"})
}

func testContext() context.Context {
	return context.Background()
}

// validHeartbeatConfig is a baseline config that passes validation.
// Tests override fields they care about.
func validHeartbeatConfig() HeartbeatConfig {
	return HeartbeatConfig{
		Enabled:                    true,
		Cron:                       "*/1 * * * *",
		TimeoutSeconds:             5,
		RetryMaxAttempts:           3,
		RetryBaseDelayMS:           200,
		RetryMaxDelayMS:            2000,
		RetryJitter:                0.2,
		CircuitFailureThreshold:    3,
		CircuitOpenTimeoutSeconds:  20,
		CircuitHalfOpenMaxRequests: 1,
		CircuitHalfOpenSuccesses:   1,
	}
}

// --- validateHeartbeatConfig --- table driven ------------------------------

func TestValidateHeartbeatConfig(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*HeartbeatConfig)
		wantErr bool
	}{
		{"valid baseline", func(c *HeartbeatConfig) {}, false},
		{"empty cron", func(c *HeartbeatConfig) { c.Cron = "" }, true},
		{"whitespace cron", func(c *HeartbeatConfig) { c.Cron = "   " }, true},
		{"zero timeout", func(c *HeartbeatConfig) { c.TimeoutSeconds = 0 }, true},
		{"negative timeout", func(c *HeartbeatConfig) { c.TimeoutSeconds = -1 }, true},
		{"zero retry attempts", func(c *HeartbeatConfig) { c.RetryMaxAttempts = 0 }, true},
		{"zero base delay", func(c *HeartbeatConfig) { c.RetryBaseDelayMS = 0 }, true},
		{"zero max delay", func(c *HeartbeatConfig) { c.RetryMaxDelayMS = 0 }, true},
		{"max < base", func(c *HeartbeatConfig) { c.RetryMaxDelayMS = 100; c.RetryBaseDelayMS = 200 }, true},
		{"jitter negative", func(c *HeartbeatConfig) { c.RetryJitter = -0.1 }, true},
		{"jitter > 1", func(c *HeartbeatConfig) { c.RetryJitter = 1.5 }, true},
		{"jitter at boundary (0)", func(c *HeartbeatConfig) { c.RetryJitter = 0 }, false},
		{"jitter at boundary (1)", func(c *HeartbeatConfig) { c.RetryJitter = 1 }, false},
		{"zero cb failure threshold", func(c *HeartbeatConfig) { c.CircuitFailureThreshold = 0 }, true},
		{"zero cb open timeout", func(c *HeartbeatConfig) { c.CircuitOpenTimeoutSeconds = 0 }, true},
		{"zero cb half-open max", func(c *HeartbeatConfig) { c.CircuitHalfOpenMaxRequests = 0 }, true},
		{"zero cb half-open successes", func(c *HeartbeatConfig) { c.CircuitHalfOpenSuccesses = 0 }, true},
		{"negative simulate-fail", func(c *HeartbeatConfig) { c.SimulateFailEvery = -1 }, true},
		{"simulate-fail zero allowed", func(c *HeartbeatConfig) { c.SimulateFailEvery = 0 }, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validHeartbeatConfig()
			tc.mutate(&cfg)
			err := validateHeartbeatConfig(cfg)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// --- NewHeartbeatScheduler --------------------------------------------------

func TestNewHeartbeatScheduler_Disabled(t *testing.T) {
	scheduler, err := NewHeartbeatScheduler(HeartbeatConfig{Enabled: false}, testLogger())
	if err != nil {
		t.Fatalf("got %v", err)
	}
	if scheduler != nil {
		t.Fatal("expected nil scheduler when disabled")
	}
}

func TestNewHeartbeatScheduler_InvalidConfigRejected(t *testing.T) {
	cfg := validHeartbeatConfig()
	cfg.Cron = ""
	if _, err := NewHeartbeatScheduler(cfg, testLogger()); err == nil {
		t.Fatal("expected error for invalid cron")
	}
}

func TestNewHeartbeatScheduler_ValidConfigSucceeds(t *testing.T) {
	scheduler, err := NewHeartbeatScheduler(validHeartbeatConfig(), testLogger())
	if err != nil {
		t.Fatalf("got %v", err)
	}
	if scheduler == nil {
		t.Fatal("expected non-nil scheduler")
	}
	// Shutdown immediately to clean up the gocron goroutine.
	defer scheduler.Shutdown(testContext())
}

// --- buildHeartbeatBreaker --------------------------------------------------

func TestBuildHeartbeatBreaker_MapsConfig(t *testing.T) {
	cfg := HeartbeatConfig{
		CircuitFailureThreshold:    7,
		CircuitOpenTimeoutSeconds:  30,
		CircuitHalfOpenMaxRequests: 2,
		CircuitHalfOpenSuccesses:   3,
	}
	br := buildHeartbeatBreaker(cfg)
	if br == nil {
		t.Fatal("expected non-nil breaker")
	}
	// Snapshot starts in closed state with zero counts.
	snap := br.Snapshot()
	if snap.State != CircuitClosed {
		t.Fatalf("initial state: got %v want CircuitClosed", snap.State)
	}
	if snap.FailureCount != 0 || snap.SuccessCount != 0 {
		t.Fatalf("initial counts non-zero: %+v", snap)
	}
}

// --- buildHeartbeatRetryConfig ---------------------------------------------

func TestBuildHeartbeatRetryConfig_MapsConfig(t *testing.T) {
	cfg := HeartbeatConfig{
		RetryMaxAttempts: 5,
		RetryBaseDelayMS: 100,
		RetryMaxDelayMS:  3000,
		RetryJitter:      0.5,
	}
	rc := buildHeartbeatRetryConfig(cfg, testLogger())
	if rc.MaxAttempts != 5 {
		t.Errorf("MaxAttempts: got %d want 5", rc.MaxAttempts)
	}
	if rc.Backoff.BaseDelay != 100*time.Millisecond {
		t.Errorf("BaseDelay: got %v want 100ms", rc.Backoff.BaseDelay)
	}
	if rc.Backoff.MaxDelay != 3000*time.Millisecond {
		t.Errorf("MaxDelay: got %v want 3s", rc.Backoff.MaxDelay)
	}
	if rc.Backoff.Jitter != 0.5 {
		t.Errorf("Jitter: got %v want 0.5", rc.Backoff.Jitter)
	}
	if rc.OnRetry == nil {
		t.Error("OnRetry should be non-nil")
	}
}

// --- DebugStatus -----------------------------------------------------------

func TestDebugStatus_NilSchedulerReturnsDisabled(t *testing.T) {
	var h *HeartbeatScheduler
	status := h.DebugStatus()
	if status.Enabled {
		t.Fatal("expected Enabled=false for nil scheduler")
	}
	if status.State != CircuitClosed {
		t.Fatalf("expected CircuitClosed, got %v", status.State)
	}
}

func TestDebugStatus_FreshSchedulerEnabled(t *testing.T) {
	scheduler, err := NewHeartbeatScheduler(validHeartbeatConfig(), testLogger())
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	defer scheduler.Shutdown(testContext())

	status := scheduler.DebugStatus()
	if !status.Enabled {
		t.Error("expected Enabled=true")
	}
	if status.RunCount != 0 {
		t.Errorf("RunCount: got %d want 0", status.RunCount)
	}
	if status.State != CircuitClosed {
		t.Errorf("State: got %v want CircuitClosed", status.State)
	}
	if status.LastError != "" {
		t.Errorf("LastError should be empty, got %q", status.LastError)
	}
}

// --- record* (state mutation) ---------------------------------------------

func TestRecordRunSuccess_SetsTimestampAndClearsError(t *testing.T) {
	h := &HeartbeatScheduler{}
	h.lastError = "previous failure"
	h.recordRunSuccess()
	if h.lastRunAt.IsZero() {
		t.Error("lastRunAt should be set")
	}
	if h.lastSuccessAt == nil {
		t.Error("lastSuccessAt should be set")
	}
	if h.lastError != "" {
		t.Errorf("lastError not cleared: %q", h.lastError)
	}
}

func TestRecordRunFailure_SetsErrorAndTimestamp(t *testing.T) {
	h := &HeartbeatScheduler{}
	h.recordRunFailure(errString("boom"))
	if h.lastRunAt.IsZero() {
		t.Error("lastRunAt should be set")
	}
	if h.lastError != "boom" {
		t.Errorf("lastError: got %q want %q", h.lastError, "boom")
	}
	if h.lastSuccessAt != nil {
		t.Error("lastSuccessAt should remain nil on failure-only path")
	}
}
