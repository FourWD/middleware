package infra

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-co-op/gocron/v2"
)

const heartbeatComponent = "background_job"

// HeartbeatScheduler runs a periodic heartbeat job with retry and circuit
// breaker. Configure via env vars (see LoadHeartbeatConfig).
type HeartbeatScheduler struct {
	scheduler gocron.Scheduler
	breaker   *CircuitBreaker
	runCount  *atomic.Uint64

	mu            sync.RWMutex
	lastRunAt     time.Time
	lastSuccessAt *time.Time
	lastError     string
}

// HeartbeatConfig holds heartbeat scheduler configuration.
type HeartbeatConfig struct {
	Enabled                    bool
	Cron                       string
	TimeoutSeconds             int
	RetryMaxAttempts           int
	RetryBaseDelayMS           int
	RetryMaxDelayMS            int
	RetryJitter                float64
	CircuitFailureThreshold    int
	CircuitOpenTimeoutSeconds  int
	CircuitHalfOpenMaxRequests int
	CircuitHalfOpenSuccesses   int
	SimulateFailEvery          int
}

// LoadHeartbeatConfig reads heartbeat configuration from environment variables.
func LoadHeartbeatConfig() HeartbeatConfig {
	return HeartbeatConfig{
		Enabled:                    GetEnvBool("BACKGROUND_JOBS_ENABLED", false),
		Cron:                       GetEnv("BACKGROUND_HEARTBEAT_CRON", "*/1 * * * *"),
		TimeoutSeconds:             GetEnvInt("BACKGROUND_HEARTBEAT_TIMEOUT_SECONDS", 5),
		RetryMaxAttempts:           GetEnvInt("BACKGROUND_RETRY_MAX_ATTEMPTS", 3),
		RetryBaseDelayMS:           GetEnvInt("BACKGROUND_RETRY_BASE_DELAY_MS", 200),
		RetryMaxDelayMS:            GetEnvInt("BACKGROUND_RETRY_MAX_DELAY_MS", 2000),
		RetryJitter:                GetEnvFloat("BACKGROUND_RETRY_JITTER", 0.2),
		CircuitFailureThreshold:    GetEnvInt("BACKGROUND_CB_FAILURE_THRESHOLD", 3),
		CircuitOpenTimeoutSeconds:  GetEnvInt("BACKGROUND_CB_OPEN_TIMEOUT_SECONDS", 20),
		CircuitHalfOpenMaxRequests: GetEnvInt("BACKGROUND_CB_HALF_OPEN_MAX_REQUESTS", 1),
		CircuitHalfOpenSuccesses:   GetEnvInt("BACKGROUND_CB_HALF_OPEN_SUCCESSES", 1),
		SimulateFailEvery:          GetEnvInt("BACKGROUND_HEARTBEAT_SIMULATE_FAIL_EVERY", 0),
	}
}

// NewHeartbeatScheduler creates a heartbeat scheduler. Returns nil when
// cfg.Enabled is false.
func NewHeartbeatScheduler(cfg HeartbeatConfig, appLogger *Logger) (*HeartbeatScheduler, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if err := validateHeartbeatConfig(cfg); err != nil {
		return nil, err
	}

	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.Local))
	if err != nil {
		return nil, fmt.Errorf("create scheduler: %w", err)
	}

	var runCount atomic.Uint64
	h := &HeartbeatScheduler{
		scheduler: scheduler,
		breaker:   buildHeartbeatBreaker(cfg),
		runCount:  &runCount,
	}

	retryConfig := buildHeartbeatRetryConfig(cfg, appLogger)

	_, err = scheduler.NewJob(
		gocron.CronJob(cfg.Cron, false),
		gocron.NewTask(func() { h.tick(cfg, appLogger, retryConfig) }),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return nil, fmt.Errorf("register heartbeat job: %w", err)
	}

	return h, nil
}

func buildHeartbeatBreaker(cfg HeartbeatConfig) *CircuitBreaker {
	return NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold:   cfg.CircuitFailureThreshold,
		OpenTimeout:        time.Duration(cfg.CircuitOpenTimeoutSeconds) * time.Second,
		HalfOpenMaxRequest: cfg.CircuitHalfOpenMaxRequests,
		HalfOpenSuccesses:  cfg.CircuitHalfOpenSuccesses,
	})
}

func buildHeartbeatRetryConfig(cfg HeartbeatConfig, appLogger *Logger) RetryConfig {
	return RetryConfig{
		MaxAttempts: cfg.RetryMaxAttempts,
		Backoff: BackoffConfig{
			BaseDelay: time.Duration(cfg.RetryBaseDelayMS) * time.Millisecond,
			MaxDelay:  time.Duration(cfg.RetryMaxDelayMS) * time.Millisecond,
			Jitter:    cfg.RetryJitter,
		},
		OnRetry: func(attempt int, err error, nextDelay time.Duration) {
			incrementHeartbeatRetryMetric()
			appLogger.EventWarn("HEARTBEAT_RETRY", map[string]any{
				"attempt":       attempt,
				"error":         err.Error(),
				"next_delay_ms": nextDelay.Milliseconds(),
			}, "",
				WithComponent(heartbeatComponent),
				WithOperation("heartbeat_retry"),
				WithLogKind(LogKindDiagnostic))
		},
	}
}

// tick runs one heartbeat invocation: panic-safe, bounded by timeout,
// gated by circuit breaker, with retries.
func (h *HeartbeatScheduler) tick(cfg HeartbeatConfig, appLogger *Logger, retryConfig RetryConfig) {
	defer h.recoverFromPanic(appLogger)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	currentRun := h.runCount.Add(1)
	shouldFail := cfg.SimulateFailEvery > 0 && currentRun%uint64(cfg.SimulateFailEvery) == 0

	err := h.breaker.Execute(ctx, func(ctx context.Context) error {
		return DoWithRetry(ctx, retryConfig, func(context.Context) error {
			if shouldFail {
				return errors.New("simulated transient heartbeat failure")
			}
			appLogger.Event("HEARTBEAT_TICK", nil, "",
				WithComponent(heartbeatComponent),
				WithOperation("heartbeat_tick"),
				WithLogKind(LogKindDiagnostic))
			return nil
		})
	})

	if err != nil {
		h.handleTickFailure(appLogger, err)
		return
	}
	incrementHeartbeatRunMetric("success")
	h.recordRunSuccess()
}

func (h *HeartbeatScheduler) recoverFromPanic(appLogger *Logger) {
	r := recover()
	if r == nil {
		return
	}
	err := fmt.Errorf("panic recovered in heartbeat job: %v", r)
	appLogger.EventError(err, "HEARTBEAT_PANIC", nil, "",
		WithComponent(heartbeatComponent),
		WithOperation("heartbeat_panic"),
		WithLogKind(LogKindError))
	h.recordRunFailure(err)
}

func (h *HeartbeatScheduler) handleTickFailure(appLogger *Logger, err error) {
	incrementHeartbeatRunMetric("failure")
	if h.breaker.Snapshot().State == CircuitOpen {
		incrementBackgroundCircuitBreakerOpenMetric()
	}
	appLogger.EventError(err, "HEARTBEAT_FAILURE", map[string]any{
		"circuit_state": h.breaker.Snapshot().State,
	}, "",
		WithComponent(heartbeatComponent),
		WithOperation("heartbeat_run"),
		WithLogKind(LogKindError))
	h.recordRunFailure(err)
}

func validateHeartbeatConfig(cfg HeartbeatConfig) error {
	if strings.TrimSpace(cfg.Cron) == "" {
		return fmt.Errorf("invalid BACKGROUND_HEARTBEAT_CRON")
	}
	if cfg.TimeoutSeconds <= 0 {
		return fmt.Errorf("invalid BACKGROUND_HEARTBEAT_TIMEOUT_SECONDS")
	}
	if cfg.RetryMaxAttempts <= 0 {
		return fmt.Errorf("invalid BACKGROUND_RETRY_MAX_ATTEMPTS")
	}
	if cfg.RetryBaseDelayMS <= 0 {
		return fmt.Errorf("invalid BACKGROUND_RETRY_BASE_DELAY_MS")
	}
	if cfg.RetryMaxDelayMS <= 0 {
		return fmt.Errorf("invalid BACKGROUND_RETRY_MAX_DELAY_MS")
	}
	if cfg.RetryMaxDelayMS < cfg.RetryBaseDelayMS {
		return fmt.Errorf("BACKGROUND_RETRY_MAX_DELAY_MS must be >= BACKGROUND_RETRY_BASE_DELAY_MS")
	}
	if cfg.RetryJitter < 0 || cfg.RetryJitter > 1 {
		return fmt.Errorf("invalid BACKGROUND_RETRY_JITTER")
	}
	if cfg.CircuitFailureThreshold <= 0 {
		return fmt.Errorf("invalid BACKGROUND_CB_FAILURE_THRESHOLD")
	}
	if cfg.CircuitOpenTimeoutSeconds <= 0 {
		return fmt.Errorf("invalid BACKGROUND_CB_OPEN_TIMEOUT_SECONDS")
	}
	if cfg.CircuitHalfOpenMaxRequests <= 0 {
		return fmt.Errorf("invalid BACKGROUND_CB_HALF_OPEN_MAX_REQUESTS")
	}
	if cfg.CircuitHalfOpenSuccesses <= 0 {
		return fmt.Errorf("invalid BACKGROUND_CB_HALF_OPEN_SUCCESSES")
	}
	if cfg.SimulateFailEvery < 0 {
		return fmt.Errorf("invalid BACKGROUND_HEARTBEAT_SIMULATE_FAIL_EVERY")
	}
	return nil
}

func (h *HeartbeatScheduler) recordRunSuccess() {
	now := time.Now()
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastRunAt = now
	h.lastSuccessAt = &now
	h.lastError = ""
}

func (h *HeartbeatScheduler) recordRunFailure(err error) {
	now := time.Now()
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastRunAt = now
	h.lastError = err.Error()
}

func (h *HeartbeatScheduler) Start() {
	if h == nil || h.scheduler == nil {
		return
	}
	h.scheduler.Start()
}

func (h *HeartbeatScheduler) Shutdown(ctx context.Context) error {
	if h == nil || h.scheduler == nil {
		return nil
	}

	done := make(chan error, 1)
	go func() {
		done <- h.scheduler.Shutdown()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

type HeartbeatDebugStatus struct {
	Enabled       bool         `json:"enabled"`
	RunCount      uint64       `json:"run_count"`
	State         CircuitState `json:"state"`
	FailureCount  int          `json:"failure_count"`
	SuccessCount  int          `json:"success_count"`
	OpenedAt      *time.Time   `json:"opened_at,omitempty"`
	LastRunAt     *time.Time   `json:"last_run_at,omitempty"`
	LastSuccessAt *time.Time   `json:"last_success_at,omitempty"`
	LastError     string       `json:"last_error,omitempty"`
}

func (h *HeartbeatScheduler) DebugStatus() HeartbeatDebugStatus {
	if h == nil || h.breaker == nil || h.runCount == nil {
		return HeartbeatDebugStatus{Enabled: false, State: CircuitClosed}
	}

	snapshot := h.breaker.Snapshot()

	h.mu.RLock()
	lastRunAt := h.lastRunAt
	lastSuccessAt := h.lastSuccessAt
	lastError := h.lastError
	h.mu.RUnlock()

	status := HeartbeatDebugStatus{
		Enabled:      true,
		RunCount:     h.runCount.Load(),
		State:        snapshot.State,
		FailureCount: snapshot.FailureCount,
		SuccessCount: snapshot.SuccessCount,
		LastError:    lastError,
	}
	if !snapshot.OpenedAt.IsZero() {
		opened := snapshot.OpenedAt
		status.OpenedAt = &opened
	}
	if !lastRunAt.IsZero() {
		lr := lastRunAt
		status.LastRunAt = &lr
	}
	if lastSuccessAt != nil {
		ls := *lastSuccessAt
		status.LastSuccessAt = &ls
	}
	return status
}
