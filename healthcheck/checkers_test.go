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

func TestNewCacheChecker(T *testing.T) {
	T.Parallel()

	T.Run("ready", func(t *testing.T) {
		t.Parallel()

		client := &mockCacheClient{err: nil}
		checker := NewCacheChecker("redis", client)
		ctx := context.Background()

		assert.Equal(t, "redis", checker.Name())
		err := checker.Check(ctx)
		require.NoError(t, err)
	})

	T.Run("not ready", func(t *testing.T) {
		t.Parallel()

		client := &mockCacheClient{err: assert.AnError}
		checker := NewCacheChecker("redis", client)
		ctx := context.Background()

		err := checker.Check(ctx)
		require.Error(t, err)
	})

	T.Run("nil client", func(t *testing.T) {
		t.Parallel()

		checker := NewCacheChecker("redis", nil)
		ctx := context.Background()

		err := checker.Check(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})
}

type mockCacheClient struct {
	err error
}

func (m *mockCacheClient) Ping(ctx context.Context) error {
	return m.err
}

func TestNewMessageQueueChecker(T *testing.T) {
	T.Parallel()

	T.Run("ready", func(t *testing.T) {
		t.Parallel()

		client := &mockMQClient{err: nil}
		checker := NewMessageQueueChecker("redis", client)
		ctx := context.Background()

		assert.Equal(t, "redis", checker.Name())
		err := checker.Check(ctx)
		require.NoError(t, err)
	})

	T.Run("not ready", func(t *testing.T) {
		t.Parallel()

		client := &mockMQClient{err: assert.AnError}
		checker := NewMessageQueueChecker("redis", client)
		ctx := context.Background()

		err := checker.Check(ctx)
		require.Error(t, err)
	})

	T.Run("nil client", func(t *testing.T) {
		t.Parallel()

		checker := NewMessageQueueChecker("redis", nil)
		ctx := context.Background()

		err := checker.Check(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})
}

type mockMQClient struct {
	err error
}

func (m *mockMQClient) Ping(ctx context.Context) error {
	return m.err
}
