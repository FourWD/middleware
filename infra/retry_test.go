package infra

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryDo_SuccessAfterRetries(t *testing.T) {
	attempt := 0
	err := DoWithRetry(context.Background(), RetryConfig{
		MaxAttempts: 3,
		Backoff: BackoffConfig{
			BaseDelay: 1 * time.Millisecond,
			MaxDelay:  2 * time.Millisecond,
			Jitter:    0,
		},
	}, func(context.Context) error {
		attempt++
		if attempt < 3 {
			return errors.New("transient")
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 3, attempt)
}

func TestRetryDo_StopOnNonRetryableError(t *testing.T) {
	targetErr := errors.New("do not retry")
	attempt := 0
	err := DoWithRetry(context.Background(), RetryConfig{
		MaxAttempts: 5,
		ShouldRetry: func(err error) bool {
			return err != targetErr
		},
	}, func(context.Context) error {
		attempt++
		return targetErr
	})

	require.ErrorIs(t, err, targetErr)
	assert.Equal(t, 1, attempt)
}

func TestRetryDo_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := DoWithRetry(ctx, RetryConfig{MaxAttempts: 3}, func(context.Context) error {
		return errors.New("should not execute")
	})
	require.ErrorIs(t, err, context.Canceled)
}
