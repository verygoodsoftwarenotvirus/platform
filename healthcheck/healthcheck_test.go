package healthcheck

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockChecker struct {
	checkFn func(ctx context.Context) error
	name    string
}

func (m *mockChecker) Name() string {
	return m.name
}

func (m *mockChecker) Check(ctx context.Context) error {
	if m.checkFn != nil {
		return m.checkFn(ctx)
	}
	return nil
}

func TestRegistry_CheckAll(t *testing.T) {
	t.Parallel()

	t.Run("empty registry returns up", func(t *testing.T) {
		t.Parallel()

		reg := NewRegistry()
		ctx := context.Background()

		result := reg.CheckAll(ctx)

		require.NotNil(t, result)
		assert.Equal(t, StatusUp, result.Status)
		assert.Empty(t, result.Components)
	})

	t.Run("all checkers up", func(t *testing.T) {
		t.Parallel()

		reg := NewRegistry()
		reg.Register(&mockChecker{name: "a"})
		reg.Register(&mockChecker{name: "b"})
		ctx := context.Background()

		result := reg.CheckAll(ctx)

		require.NotNil(t, result)
		assert.Equal(t, StatusUp, result.Status)
		assert.Len(t, result.Components, 2)
		assert.Equal(t, ComponentResult{Status: StatusUp}, result.Components["a"])
		assert.Equal(t, ComponentResult{Status: StatusUp}, result.Components["b"])
	})

	t.Run("one checker down", func(t *testing.T) {
		t.Parallel()

		reg := NewRegistry()
		reg.Register(&mockChecker{name: "up"})
		reg.Register(&mockChecker{
			name: "down",
			checkFn: func(context.Context) error {
				return errors.New("connection refused")
			},
		})
		ctx := context.Background()

		result := reg.CheckAll(ctx)

		require.NotNil(t, result)
		assert.Equal(t, StatusDown, result.Status)
		assert.Len(t, result.Components, 2)
		assert.Equal(t, ComponentResult{Status: StatusUp}, result.Components["up"])
		assert.Equal(t, ComponentResult{Status: StatusDown, Message: "connection refused"}, result.Components["down"])
	})

	t.Run("ignores nil checker", func(t *testing.T) {
		t.Parallel()

		reg := NewRegistry()
		reg.Register(nil)
		reg.Register(&mockChecker{name: "a"})
		ctx := context.Background()

		result := reg.CheckAll(ctx)

		require.NotNil(t, result)
		assert.Equal(t, StatusUp, result.Status)
		assert.Len(t, result.Components, 1)
	})
}
