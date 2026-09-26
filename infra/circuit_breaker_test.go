package infra

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_OpenAndReject(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		OpenTimeout:      50 * time.Millisecond,
	})

	err := cb.Execute(context.Background(), func(context.Context) error {
		return errors.New("fail")
	})
	require.Error(t, err)

	err = cb.Execute(context.Background(), func(context.Context) error {
		return errors.New("fail")
	})
	require.Error(t, err)

	err = cb.Execute(context.Background(), func(context.Context) error {
		return nil
	})
	require.ErrorIs(t, err, ErrCircuitOpen)
	assert.Equal(t, CircuitOpen, cb.Snapshot().State)
}

func TestCircuitBreaker_HalfOpenAndClose(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold:   1,
		OpenTimeout:        30 * time.Millisecond,
		HalfOpenMaxRequest: 1,
		HalfOpenSuccesses:  1,
	})

	err := cb.Execute(context.Background(), func(context.Context) error {
		return errors.New("fail")
	})
	require.Error(t, err)
	require.Equal(t, CircuitOpen, cb.Snapshot().State)

	time.Sleep(35 * time.Millisecond)

	err = cb.Execute(context.Background(), func(context.Context) error {
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, CircuitClosed, cb.Snapshot().State)
}

func TestCircuitBreaker_PanicInHalfOpenDoesNotWedge(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold:   1,
		OpenTimeout:        20 * time.Millisecond,
		HalfOpenMaxRequest: 1,
		HalfOpenSuccesses:  1,
	})

	_ = cb.Execute(context.Background(), func(context.Context) error { return errors.New("fail") })
	require.Equal(t, CircuitOpen, cb.Snapshot().State)
	time.Sleep(25 * time.Millisecond)

	require.PanicsWithValue(t, "boom", func() {
		_ = cb.Execute(context.Background(), func(context.Context) error { panic("boom") })
	})
	require.Equal(t, CircuitOpen, cb.Snapshot().State, "a panicking probe counts as a failure")

	time.Sleep(25 * time.Millisecond)
	err := cb.Execute(context.Background(), func(context.Context) error { return nil })
	require.NoError(t, err)
	assert.Equal(t, CircuitClosed, cb.Snapshot().State)
}

func TestCircuitBreaker_StaleResultIgnoredInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold:   1,
		OpenTimeout:        20 * time.Millisecond,
		HalfOpenMaxRequest: 1,
		HalfOpenSuccesses:  1,
	})

	staleRelease := make(chan struct{})
	staleDone := make(chan error, 1)
	staleStarted := make(chan struct{})
	go func() {
		staleDone <- cb.Execute(context.Background(), func(context.Context) error {
			close(staleStarted)
			<-staleRelease
			return nil
		})
	}()
	<-staleStarted

	_ = cb.Execute(context.Background(), func(context.Context) error { return errors.New("fail") })
	require.Equal(t, CircuitOpen, cb.Snapshot().State)
	time.Sleep(25 * time.Millisecond)

	probeRelease := make(chan struct{})
	probeDone := make(chan error, 1)
	probeStarted := make(chan struct{})
	go func() {
		probeDone <- cb.Execute(context.Background(), func(context.Context) error {
			close(probeStarted)
			<-probeRelease
			return errors.New("probe fail")
		})
	}()
	<-probeStarted
	require.Equal(t, CircuitHalfOpen, cb.Snapshot().State)

	close(staleRelease)
	require.NoError(t, <-staleDone)
	snap := cb.Snapshot()
	assert.Equal(t, CircuitHalfOpen, snap.State, "stale success must not close the breaker")
	assert.Equal(t, 0, snap.SuccessCount)
	require.ErrorIs(t, cb.Execute(context.Background(), func(context.Context) error { return nil }), ErrCircuitOpen,
		"the half-open slot is still held by the real probe")

	close(probeRelease)
	require.Error(t, <-probeDone)
	assert.Equal(t, CircuitOpen, cb.Snapshot().State)
}
