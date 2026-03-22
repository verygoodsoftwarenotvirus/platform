package healthcheck

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabaseChecker(T *testing.T) {
	T.Parallel()

	T.Run("ready", func(t *testing.T) {
		t.Parallel()

		client := &mockDBClient{ready: true}
		checker := NewDatabaseChecker("postgres", client)
		ctx := context.Background()

		assert.Equal(t, "postgres", checker.Name())
		err := checker.Check(ctx)
		require.NoError(t, err)
	})

	T.Run("not ready", func(t *testing.T) {
		t.Parallel()

		client := &mockDBClient{ready: false}
		checker := NewDatabaseChecker("postgres", client)
		ctx := context.Background()

		err := checker.Check(ctx)
		require.Error(t, err)
	})

	T.Run("nil client", func(t *testing.T) {
		t.Parallel()

		checker := NewDatabaseChecker("postgres", nil)
		ctx := context.Background()

		err := checker.Check(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})
}

type mockDBClient struct {
	ready bool
}

func (m *mockDBClient) IsReady(ctx context.Context) bool {
	return m.ready
}
