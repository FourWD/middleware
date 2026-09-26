package infra

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half_open"
)

type CircuitBreakerConfig struct {
	FailureThreshold   int
	OpenTimeout        time.Duration
	HalfOpenMaxRequest int
	HalfOpenSuccesses  int
}

type CircuitSnapshot struct {
	State        CircuitState
	FailureCount int
	SuccessCount int
	OpenedAt     time.Time
}

type CircuitBreaker struct {
	mu       sync.Mutex
	cfg      CircuitBreakerConfig
	state    CircuitState
	failures int
	success  int
	openedAt time.Time
	inFlight int
	// generation changes on every state transition so results of calls
	// admitted under an earlier state are not counted against the current one.
	generation uint64
}

var errCircuitPanic = errors.New("circuit breaker: fn panicked")

func defaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:   5,
		OpenTimeout:        30 * time.Second,
		HalfOpenMaxRequest: 1,
		HalfOpenSuccesses:  1,
	}
}

func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	c := cfg
	def := defaultCircuitBreakerConfig()

	if c.FailureThreshold < 1 {
		c.FailureThreshold = def.FailureThreshold
	}
	if c.OpenTimeout <= 0 {
		c.OpenTimeout = def.OpenTimeout
	}
	if c.HalfOpenMaxRequest < 1 {
		c.HalfOpenMaxRequest = def.HalfOpenMaxRequest
	}
	if c.HalfOpenSuccesses < 1 {
		c.HalfOpenSuccesses = def.HalfOpenSuccesses
	}

	return &CircuitBreaker{
		cfg:   c,
		state: CircuitClosed,
	}
}

func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	gen, ok := cb.allow()
	if !ok {
		return ErrCircuitOpen
	}

	// Deferred so a panicking fn still releases its half-open slot and counts
	// as a failure; the panic itself propagates unchanged.
	err := errCircuitPanic
	defer func() { cb.afterExecution(gen, err) }()

	err = fn(ctx)
	return err
}

func (cb *CircuitBreaker) Snapshot() CircuitSnapshot {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return CircuitSnapshot{
		State:        cb.state,
		FailureCount: cb.failures,
		SuccessCount: cb.success,
		OpenedAt:     cb.openedAt,
	}
}

func (cb *CircuitBreaker) allow() (uint64, bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if cb.state == CircuitOpen && now.Sub(cb.openedAt) >= cb.cfg.OpenTimeout {
		cb.transitionLocked(CircuitHalfOpen)
	}

	switch cb.state {
	case CircuitClosed:
		return cb.generation, true
	case CircuitOpen:
		return 0, false
	case CircuitHalfOpen:
		if cb.inFlight >= cb.cfg.HalfOpenMaxRequest {
			return 0, false
		}
		cb.inFlight++
		return cb.generation, true
	default:
		return 0, false
	}
}

func (cb *CircuitBreaker) afterExecution(gen uint64, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Every transition resets inFlight, so a stale result has no slot to
	// release and must not influence the new state.
	if gen != cb.generation {
		return
	}

	if cb.state == CircuitHalfOpen && cb.inFlight > 0 {
		cb.inFlight--
	}

	switch cb.state {
	case CircuitClosed:
		if err == nil {
			cb.failures = 0
			return
		}
		cb.failures++
		if cb.failures >= cb.cfg.FailureThreshold {
			cb.openLocked()
		}
	case CircuitHalfOpen:
		if err != nil {
			cb.openLocked()
			return
		}
		cb.success++
		if cb.success >= cb.cfg.HalfOpenSuccesses {
			cb.closeLocked()
		}
	}
}

func (cb *CircuitBreaker) openLocked() {
	cb.transitionLocked(CircuitOpen)
	cb.openedAt = time.Now()
}

func (cb *CircuitBreaker) closeLocked() {
	cb.transitionLocked(CircuitClosed)
	cb.openedAt = time.Time{}
}

func (cb *CircuitBreaker) transitionLocked(state CircuitState) {
	cb.state = state
	cb.failures = 0
	cb.success = 0
	cb.inFlight = 0
	cb.generation++
}
