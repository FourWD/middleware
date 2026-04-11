package infra

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// BackoffConfig controls exponential backoff behaviour for retry and circuit breaker.
type BackoffConfig struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
	Jitter     float64
}

func defaultBackoffConfig() BackoffConfig {
	return BackoffConfig{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Multiplier: 2,
		Jitter:     0.2,
	}
}

func (c BackoffConfig) normalized() BackoffConfig {
	cfg := c
	def := defaultBackoffConfig()

	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = def.BaseDelay
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = def.MaxDelay
	}
	if cfg.MaxDelay < cfg.BaseDelay {
		cfg.MaxDelay = cfg.BaseDelay
	}
	if cfg.Multiplier < 1 {
		cfg.Multiplier = def.Multiplier
	}
	if cfg.Jitter < 0 {
		cfg.Jitter = 0
	}

	return cfg
}

func ExponentialBackoff(attempt int, cfg BackoffConfig, rng *rand.Rand) time.Duration {
	if attempt < 1 {
		attempt = 1
	}

	c := cfg.normalized()
	backoff := float64(c.BaseDelay) * math.Pow(c.Multiplier, float64(attempt-1))
	if backoff > float64(c.MaxDelay) {
		backoff = float64(c.MaxDelay)
	}

	if c.Jitter == 0 {
		return time.Duration(backoff)
	}

	r := rng
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	min := 1 - c.Jitter
	max := 1 + c.Jitter
	factor := min + r.Float64()*(max-min)
	return time.Duration(backoff * factor)
}

// RetryConfig controls retry behaviour.
type RetryConfig struct {
	MaxAttempts int
	Backoff     BackoffConfig
	ShouldRetry func(error) bool
	OnRetry     func(attempt int, err error, nextDelay time.Duration)
}

func defaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Backoff:     defaultBackoffConfig(),
		ShouldRetry: func(err error) bool { return err != nil },
	}
}

func (c RetryConfig) normalized() RetryConfig {
	cfg := c
	def := defaultRetryConfig()

	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = def.MaxAttempts
	}
	if cfg.ShouldRetry == nil {
		cfg.ShouldRetry = def.ShouldRetry
	}
	cfg.Backoff = cfg.Backoff.normalized()

	return cfg
}

// DoWithRetry executes fn with retry logic according to cfg.
func DoWithRetry(ctx context.Context, cfg RetryConfig, fn func(context.Context) error) error {
	c := cfg.normalized()
	var lastErr error

	for attempt := 1; attempt <= c.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}
		lastErr = err

		if attempt == c.MaxAttempts || !c.ShouldRetry(err) {
			break
		}

		delay := ExponentialBackoff(attempt, c.Backoff, nil)
		if c.OnRetry != nil {
			c.OnRetry(attempt, err, delay)
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return lastErr
}
