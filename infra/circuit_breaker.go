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
}

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
	if !cb.allow() {
		return ErrCircuitOpen
	}

	err := fn(ctx)
	cb.afterExecution(err)
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

func (cb *CircuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if cb.state == CircuitOpen && now.Sub(cb.openedAt) >= cb.cfg.OpenTimeout {
		cb.state = CircuitHalfOpen
		cb.failures = 0
		cb.success = 0
		cb.inFlight = 0
	}

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		return false
	case CircuitHalfOpen:
		if cb.inFlight >= cb.cfg.HalfOpenMaxRequest {
			return false
		}
		cb.inFlight++
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) afterExecution(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

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
	cb.state = CircuitOpen
	cb.openedAt = time.Now()
	cb.failures = 0
	cb.success = 0
	cb.inFlight = 0
}

func (cb *CircuitBreaker) closeLocked() {
	cb.state = CircuitClosed
	cb.openedAt = time.Time{}
	cb.failures = 0
	cb.success = 0
	cb.inFlight = 0
}
