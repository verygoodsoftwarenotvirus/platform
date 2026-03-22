package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExponentialBackoffPolicy_Execute(T *testing.T) {
	T.Parallel()

	T.Run("success on first attempt", func(t *testing.T) {
		t.Parallel()

		policy := NewExponentialBackoffPolicy(Config{MaxAttempts: 3})
		ctx := context.Background()
		attempts := 0

		err := policy.Execute(ctx, func(ctx context.Context) error {
			attempts++
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	T.Run("success after retries", func(t *testing.T) {
		t.Parallel()

		policy := NewExponentialBackoffPolicy(Config{
			MaxAttempts:  5,
			InitialDelay: 1,
			MaxDelay:     10,
			UseJitter:    false,
		})
		ctx := context.Background()
		attempts := 0

		err := policy.Execute(ctx, func(ctx context.Context) error {
			attempts++
			if attempts < 3 {
				return errors.New("transient")
			}
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	T.Run("returns last error after max attempts", func(t *testing.T) {
		t.Parallel()

		policy := NewExponentialBackoffPolicy(Config{
			MaxAttempts:  3,
			InitialDelay: 1,
			MaxDelay:     10,
			UseJitter:    false,
		})
		ctx := context.Background()
		attempts := 0
		expectedErr := errors.New("final failure")

		err := policy.Execute(ctx, func(ctx context.Context) error {
			attempts++
			if attempts < 3 {
				return errors.New("transient")
			}
			return expectedErr
		})

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 3, attempts)
	})

	T.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()

		policy := NewExponentialBackoffPolicy(Config{
			MaxAttempts:  10,
			InitialDelay: time.Hour,
			UseJitter:    false,
		})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := policy.Execute(ctx, func(ctx context.Context) error {
			return errors.New("fail")
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
