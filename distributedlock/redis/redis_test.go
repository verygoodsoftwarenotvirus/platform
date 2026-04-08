package redis

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/distributedlock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/identifiers"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rediscontainers "github.com/testcontainers/testcontainers-go/modules/redis"
)

const redisImage = "docker.io/redis:7-bullseye"

var runningContainerTests = strings.ToLower(os.Getenv("RUN_CONTAINER_TESTS")) == "true"

func buildContainerBackedRedisConfig(t *testing.T) (cfg *Config, shutdown func(context.Context) error) {
	t.Helper()

	ctx := t.Context()
	container, err := rediscontainers.Run(ctx,
		redisImage,
		rediscontainers.WithLogLevel(rediscontainers.LogLevelNotice),
	)
	require.NoError(t, err)
	require.NotNil(t, container)

	addr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	cfg = &Config{
		Addresses: []string{strings.TrimPrefix(addr, "redis://")},
		KeyPrefix: "lock:",
	}
	return cfg, func(ctx context.Context) error { return container.Terminate(ctx) }
}

func newTestLocker(t *testing.T, cfg *Config) distributedlock.Locker {
	t.Helper()
	l, err := NewRedisLocker(cfg, nil, nil, nil, cbnoop.NewCircuitBreaker())
	require.NoError(t, err)
	require.NotNil(t, l)
	return l
}

// directRedisClient builds a raw go-redis client against the same address. Tests
// use it to forge ownership tokens and verify the wrong-owner branch.
func directRedisClient(t *testing.T, cfg *Config) *redis.Client {
	t.Helper()
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addresses[0],
		Username: cfg.Username,
		Password: cfg.Password,
	})
}

// --------- unit tests (no container) ---------

func TestNewRedisLocker(T *testing.T) {
	T.Parallel()

	T.Run("nil config", func(t *testing.T) {
		t.Parallel()
		_, err := NewRedisLocker(nil, nil, nil, nil, cbnoop.NewCircuitBreaker())
		require.ErrorIs(t, err, distributedlock.ErrNilConfig)
	})
}

// --------- container-backed integration tests ---------

func TestRedisLocker_Container(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	cfg, shutdown := buildContainerBackedRedisConfig(T)
	T.Cleanup(func() { _ = shutdown(context.Background()) })

	T.Run("Acquire happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, cfg)
		key := "happy_" + identifiers.New()

		lock, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		require.NotNil(t, lock)
		assert.Equal(t, key, lock.Key())
		assert.Equal(t, time.Minute, lock.TTL())

		require.NoError(t, lock.Release(ctx))
	})

	T.Run("Acquire contended returns ErrLockNotAcquired", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, cfg)
		key := "contended_" + identifiers.New()

		first, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		t.Cleanup(func() { _ = first.Release(ctx) })

		_, err = l.Acquire(ctx, key, time.Minute)
		require.ErrorIs(t, err, distributedlock.ErrLockNotAcquired)
	})

	T.Run("Acquire rejects empty key", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t, cfg)
		_, err := l.Acquire(t.Context(), "", time.Minute)
		require.ErrorIs(t, err, distributedlock.ErrEmptyKey)
	})

	T.Run("Acquire rejects zero TTL", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t, cfg)
		_, err := l.Acquire(t.Context(), "k", 0)
		require.ErrorIs(t, err, distributedlock.ErrInvalidTTL)
	})

	T.Run("Release after expiration returns ErrLockNotHeld", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, cfg)
		key := "expired_" + identifiers.New()

		lock, err := l.Acquire(ctx, key, 100*time.Millisecond)
		require.NoError(t, err)
		time.Sleep(250 * time.Millisecond)

		require.ErrorIs(t, lock.Release(ctx), distributedlock.ErrLockNotHeld)
	})

	T.Run("Release wrong owner returns ErrLockNotHeld", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, cfg)
		key := "stolen_" + identifiers.New()

		lock, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)

		// Forge a different owner by overwriting the value out-of-band.
		direct := directRedisClient(t, cfg)
		t.Cleanup(func() { _ = direct.Close() })
		require.NoError(t, direct.Set(ctx, "lock:"+key, "someone-else", time.Minute).Err())

		require.ErrorIs(t, lock.Release(ctx), distributedlock.ErrLockNotHeld)
	})

	T.Run("Refresh extends TTL", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, cfg)
		key := "refresh_" + identifiers.New()

		lock, err := l.Acquire(ctx, key, 200*time.Millisecond)
		require.NoError(t, err)
		require.NoError(t, lock.Refresh(ctx, 5*time.Second))
		t.Cleanup(func() { _ = lock.Release(ctx) })

		// Sleep past the original TTL; lock should still be held.
		time.Sleep(300 * time.Millisecond)

		_, err = l.Acquire(ctx, key, time.Minute)
		require.ErrorIs(t, err, distributedlock.ErrLockNotAcquired)
		assert.Equal(t, 5*time.Second, lock.TTL())
	})

	T.Run("Refresh rejects invalid TTL", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, cfg)
		key := "refreshinv_" + identifiers.New()

		lock, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		t.Cleanup(func() { _ = lock.Release(ctx) })

		require.ErrorIs(t, lock.Refresh(ctx, 0), distributedlock.ErrInvalidTTL)
	})

	T.Run("Double release returns ErrLockNotHeld on second call", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, cfg)
		key := "double_" + identifiers.New()

		lock, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		require.NoError(t, lock.Release(ctx))
		require.ErrorIs(t, lock.Release(ctx), distributedlock.ErrLockNotHeld)
	})

	T.Run("Released lock can be reacquired", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, cfg)
		key := "reacquire_" + identifiers.New()

		first, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		require.NoError(t, first.Release(ctx))

		second, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		require.NoError(t, second.Release(ctx))
	})

	T.Run("Ping success", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t, cfg)
		require.NoError(t, l.Ping(t.Context()))
	})
}
