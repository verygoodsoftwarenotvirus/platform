package noop

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocker(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, NewLocker())
	})
}

func TestLocker_Acquire(T *testing.T) {
	T.Parallel()

	T.Run("returns a usable handle", func(t *testing.T) {
		t.Parallel()
		l := NewLocker()
		lock, err := l.Acquire(t.Context(), "k", time.Second)
		require.NoError(t, err)
		require.NotNil(t, lock)
		assert.Equal(t, "k", lock.Key())
		assert.Equal(t, time.Second, lock.TTL())
	})

	T.Run("contended acquires both succeed", func(t *testing.T) {
		t.Parallel()
		l := NewLocker()
		_, err := l.Acquire(t.Context(), "shared", time.Second)
		require.NoError(t, err)
		_, err = l.Acquire(t.Context(), "shared", time.Second)
		require.NoError(t, err)
	})
}

func TestLocker_Ping(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, NewLocker().Ping(t.Context()))
	})
}

func TestLocker_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, NewLocker().Close())
	})
}

func TestLock_ReleaseAndRefresh(T *testing.T) {
	T.Parallel()

	T.Run("release is a no-op", func(t *testing.T) {
		t.Parallel()
		l, err := NewLocker().Acquire(t.Context(), "k", time.Second)
		require.NoError(t, err)
		require.NoError(t, l.Release(t.Context()))
		require.NoError(t, l.Release(t.Context()))
	})

	T.Run("refresh updates ttl", func(t *testing.T) {
		t.Parallel()
		l, err := NewLocker().Acquire(t.Context(), "k", time.Second)
		require.NoError(t, err)
		require.NoError(t, l.Refresh(t.Context(), 5*time.Second))
		assert.Equal(t, 5*time.Second, l.TTL())
	})
}
