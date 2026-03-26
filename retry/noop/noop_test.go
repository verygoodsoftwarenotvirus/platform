package noop

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicy_Execute(T *testing.T) {
	T.Parallel()

	T.Run("executes exactly once on success", func(t *testing.T) {
		t.Parallel()

		p := NewPolicy()
		ctx := context.Background()
		attempts := 0

		err := p.Execute(ctx, func(ctx context.Context) error {
			attempts++
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	T.Run("executes exactly once on failure", func(t *testing.T) {
		t.Parallel()

		p := NewPolicy()
		ctx := context.Background()
		attempts := 0
		expectedErr := errors.New("fail")

		err := p.Execute(ctx, func(ctx context.Context) error {
			attempts++
			return expectedErr
		})

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, attempts)
	})
}
