package memory

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/distributedlock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLocker(t *testing.T) distributedlock.Locker {
	t.Helper()
	l, err := NewLocker(nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, l)
	return l
}

func TestNewLocker(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		l, err := NewLocker(nil, nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, l)
	})
}

func TestLocker_Acquire(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		lock, err := l.Acquire(t.Context(), "k", time.Second)
		require.NoError(t, err)
		require.NotNil(t, lock)
		assert.Equal(t, "k", lock.Key())
		assert.Equal(t, time.Second, lock.TTL())
	})

	T.Run("contended", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		_, err := l.Acquire(t.Context(), "shared", time.Minute)
		require.NoError(t, err)
		_, err = l.Acquire(t.Context(), "shared", time.Minute)
		require.ErrorIs(t, err, distributedlock.ErrLockNotAcquired)
	})

	T.Run("re-acquire after expiry", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		_, err := l.Acquire(t.Context(), "exp", 50*time.Millisecond)
		require.NoError(t, err)
		time.Sleep(80 * time.Millisecond)
		_, err = l.Acquire(t.Context(), "exp", time.Second)
		require.NoError(t, err)
	})

	T.Run("rejects empty key", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		_, err := l.Acquire(t.Context(), "", time.Second)
		require.ErrorIs(t, err, distributedlock.ErrEmptyKey)
	})

	T.Run("rejects zero TTL", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		_, err := l.Acquire(t.Context(), "k", 0)
		require.ErrorIs(t, err, distributedlock.ErrInvalidTTL)
	})

	T.Run("rejects negative TTL", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		_, err := l.Acquire(t.Context(), "k", -time.Second)
		require.ErrorIs(t, err, distributedlock.ErrInvalidTTL)
	})
}

func TestLocker_Release(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		lock, err := l.Acquire(t.Context(), "k", time.Minute)
		require.NoError(t, err)
		require.NoError(t, lock.Release(t.Context()))
	})

	T.Run("released lock can be reacquired", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		first, err := l.Acquire(t.Context(), "k", time.Minute)
		require.NoError(t, err)
		require.NoError(t, first.Release(t.Context()))
		_, err = l.Acquire(t.Context(), "k", time.Minute)
		require.NoError(t, err)
	})

	T.Run("double release returns ErrLockNotHeld", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		lock, err := l.Acquire(t.Context(), "k", time.Minute)
		require.NoError(t, err)
		require.NoError(t, lock.Release(t.Context()))
		require.ErrorIs(t, lock.Release(t.Context()), distributedlock.ErrLockNotHeld)
	})

	T.Run("release after expiration returns ErrLockNotHeld", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		lock, err := l.Acquire(t.Context(), "k", 50*time.Millisecond)
		require.NoError(t, err)
		time.Sleep(80 * time.Millisecond)
		require.ErrorIs(t, lock.Release(t.Context()), distributedlock.ErrLockNotHeld)
	})
}

func TestLocker_Refresh(T *testing.T) {
	T.Parallel()

	T.Run("extends TTL", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		lock, err := l.Acquire(t.Context(), "k", 50*time.Millisecond)
		require.NoError(t, err)
		require.NoError(t, lock.Refresh(t.Context(), 5*time.Second))
		// Even after the original TTL elapses, the lock is still held.
		time.Sleep(80 * time.Millisecond)
		_, err = l.Acquire(t.Context(), "k", time.Second)
		require.ErrorIs(t, err, distributedlock.ErrLockNotAcquired)
	})

	T.Run("refresh after expiration returns ErrLockNotHeld", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		lock, err := l.Acquire(t.Context(), "k", 50*time.Millisecond)
		require.NoError(t, err)
		time.Sleep(80 * time.Millisecond)
		require.ErrorIs(t, lock.Refresh(t.Context(), time.Second), distributedlock.ErrLockNotHeld)
	})

	T.Run("rejects invalid TTL", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		lock, err := l.Acquire(t.Context(), "k", time.Minute)
		require.NoError(t, err)
		require.ErrorIs(t, lock.Refresh(t.Context(), 0), distributedlock.ErrInvalidTTL)
	})
}

func TestLocker_Ping(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, newTestLocker(t).Ping(t.Context()))
	})
}

func TestLocker_Close(T *testing.T) {
	T.Parallel()

	T.Run("closes and drops outstanding locks", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)
		lock, err := l.Acquire(t.Context(), "k", time.Minute)
		require.NoError(t, err)
		require.NoError(t, l.Close())
		// The previous handle now sees the lock as not-held.
		require.ErrorIs(t, lock.Release(t.Context()), distributedlock.ErrLockNotHeld)
		// And the key is acquirable again.
		_, err = l.Acquire(t.Context(), "k", time.Second)
		require.NoError(t, err)
	})
}

func TestLocker_Concurrency(T *testing.T) {
	T.Parallel()

	T.Run("only one goroutine wins per key", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t)

		const goroutines = 100
		var winners atomic.Int64
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for range goroutines {
			go func() {
				defer wg.Done()
				if _, err := l.Acquire(t.Context(), "racekey", time.Minute); err == nil {
					winners.Add(1)
				}
			}()
		}
		wg.Wait()

		assert.Equal(t, int64(1), winners.Load())
	})
}
