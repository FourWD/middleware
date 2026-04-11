package infra

import "testing"

func TestValidateHeartbeatConfig_InvalidJitter(t *testing.T) {
	cfg := HeartbeatConfig{
		Enabled:                    true,
		Cron:                       "*/1 * * * *",
		TimeoutSeconds:             5,
		RetryMaxAttempts:           3,
		RetryBaseDelayMS:           200,
		RetryMaxDelayMS:            2000,
		RetryJitter:                1.5,
		CircuitFailureThreshold:    3,
		CircuitOpenTimeoutSeconds:  20,
		CircuitHalfOpenMaxRequests: 1,
		CircuitHalfOpenSuccesses:   1,
	}

	if err := validateHeartbeatConfig(cfg); err == nil {
		t.Fatal("expected error for invalid BACKGROUND_RETRY_JITTER")
	}
}

func TestNewHeartbeatScheduler_Disabled(t *testing.T) {
	scheduler, err := NewHeartbeatScheduler(HeartbeatConfig{Enabled: false}, NewLoggerWith("test", "test"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if scheduler != nil {
		t.Fatal("expected nil scheduler when background jobs disabled")
	}
}
