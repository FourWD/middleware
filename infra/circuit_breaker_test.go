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
