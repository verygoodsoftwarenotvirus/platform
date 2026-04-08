package postgres

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/database"
	"github.com/verygoodsoftwarenotvirus/platform/v5/distributedlock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/identifiers"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const postgresImage = "postgres:16-alpine"

var runningContainerTests = strings.ToLower(os.Getenv("RUN_CONTAINER_TESTS")) == "true"

// testDBClient is a minimal database.Client backed by a single *sql.DB. It exists
// only to avoid pulling in database/postgres for tests in this leaf package.
type testDBClient struct {
	db *sql.DB
}

func (c *testDBClient) WriteDB() *sql.DB { return c.db }
func (c *testDBClient) ReadDB() *sql.DB  { return c.db }
func (c *testDBClient) Close() error     { return c.db.Close() }
func (c *testDBClient) CurrentTime() time.Time {
	return time.Now()
}
func (c *testDBClient) RollbackTransaction(_ context.Context, tx database.SQLQueryExecutorAndTransactionManager) {
	_ = tx.Rollback()
}

func buildContainerBackedPostgres(t *testing.T) (client *testDBClient, shutdown func(context.Context) error) {
	t.Helper()

	ctx := t.Context()
	container, err := postgrescontainer.Run(
		ctx,
		postgresImage,
		postgrescontainer.WithDatabase("locktest"),
		postgrescontainer.WithUsername("locktest"),
		postgrescontainer.WithPassword("locktest"),
		testcontainers.WithWaitStrategyAndDeadline(2*time.Minute, wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
	)
	require.NoError(t, err)
	require.NotNil(t, container)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	// Allow plenty of conns so the parallel subtests don't starve.
	db.SetMaxOpenConns(64)
	require.NoError(t, db.PingContext(ctx))

	return &testDBClient{db: db}, func(ctx context.Context) error {
		_ = db.Close()
		return container.Terminate(ctx)
	}
}

func newTestLocker(t *testing.T, client database.Client) distributedlock.Locker {
	t.Helper()
	l, err := NewPostgresLocker(&Config{}, client, nil, nil, nil, cbnoop.NewCircuitBreaker())
	require.NoError(t, err)
	require.NotNil(t, l)
	return l
}

// --------- unit tests (no container) ---------

func TestNewPostgresLocker(T *testing.T) {
	T.Parallel()

	T.Run("nil config", func(t *testing.T) {
		t.Parallel()
		_, err := NewPostgresLocker(nil, &testDBClient{}, nil, nil, nil, cbnoop.NewCircuitBreaker())
		require.ErrorIs(t, err, distributedlock.ErrNilConfig)
	})

	T.Run("nil database", func(t *testing.T) {
		t.Parallel()
		_, err := NewPostgresLocker(&Config{}, nil, nil, nil, nil, cbnoop.NewCircuitBreaker())
		require.ErrorIs(t, err, distributedlock.ErrNilDatabaseClient)
	})
}

func TestHashLockID(T *testing.T) {
	T.Parallel()

	T.Run("stable across calls", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, hashLockID(0, "k"), hashLockID(0, "k"))
	})

	T.Run("namespace changes the result", func(t *testing.T) {
		t.Parallel()
		assert.NotEqual(t, hashLockID(0, "k"), hashLockID(1, "k"))
	})

	T.Run("different keys produce different ids", func(t *testing.T) {
		t.Parallel()
		assert.NotEqual(t, hashLockID(0, "a"), hashLockID(0, "b"))
	})
}

// --------- container-backed integration tests ---------

func TestPostgresLocker_Container(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	client, shutdown := buildContainerBackedPostgres(T)
	T.Cleanup(func() { _ = shutdown(context.Background()) })

	T.Run("Acquire happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, client)
		key := "happy_" + identifiers.New()

		lock, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		require.NotNil(t, lock)
		assert.Equal(t, key, lock.Key())
		assert.Equal(t, time.Minute, lock.TTL())
		require.NoError(t, lock.Release(ctx))
	})

	T.Run("Acquire contended on the same locker returns ErrLockNotAcquired", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, client)
		key := "contend_same_" + identifiers.New()

		first, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		t.Cleanup(func() { _ = first.Release(ctx) })

		_, err = l.Acquire(ctx, key, time.Minute)
		require.ErrorIs(t, err, distributedlock.ErrLockNotAcquired)
	})

	T.Run("Acquire contended across separate lockers returns ErrLockNotAcquired", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l1 := newTestLocker(t, client)
		l2 := newTestLocker(t, client)
		key := "contend_cross_" + identifiers.New()

		first, err := l1.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		t.Cleanup(func() { _ = first.Release(ctx) })

		_, err = l2.Acquire(ctx, key, time.Minute)
		require.ErrorIs(t, err, distributedlock.ErrLockNotAcquired)
	})

	T.Run("Acquire rejects empty key", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t, client)
		_, err := l.Acquire(t.Context(), "", time.Minute)
		require.ErrorIs(t, err, distributedlock.ErrEmptyKey)
	})

	T.Run("Acquire rejects zero TTL", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t, client)
		_, err := l.Acquire(t.Context(), "k", 0)
		require.ErrorIs(t, err, distributedlock.ErrInvalidTTL)
	})

	T.Run("Released lock can be reacquired", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, client)
		key := "reacquire_" + identifiers.New()

		first, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		require.NoError(t, first.Release(ctx))

		second, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		require.NoError(t, second.Release(ctx))
	})

	T.Run("Double release returns ErrLockNotHeld on second call", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, client)
		key := "double_" + identifiers.New()

		lock, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		require.NoError(t, lock.Release(ctx))
		require.ErrorIs(t, lock.Release(ctx), distributedlock.ErrLockNotHeld)
	})

	T.Run("Refresh succeeds and updates local TTL", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, client)
		key := "refresh_" + identifiers.New()

		lock, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		t.Cleanup(func() { _ = lock.Release(ctx) })

		require.NoError(t, lock.Refresh(ctx, 5*time.Minute))
		assert.Equal(t, 5*time.Minute, lock.TTL())
	})

	T.Run("Refresh after release returns ErrLockNotHeld", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, client)
		key := "refresh_after_release_" + identifiers.New()

		lock, err := l.Acquire(ctx, key, time.Minute)
		require.NoError(t, err)
		require.NoError(t, lock.Release(ctx))

		require.ErrorIs(t, lock.Refresh(ctx, time.Minute), distributedlock.ErrLockNotHeld)
	})

	T.Run("Close releases all outstanding locks", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		l := newTestLocker(t, client)
		keyA := "close_a_" + identifiers.New()
		keyB := "close_b_" + identifiers.New()

		_, err := l.Acquire(ctx, keyA, time.Minute)
		require.NoError(t, err)
		_, err = l.Acquire(ctx, keyB, time.Minute)
		require.NoError(t, err)
		require.NoError(t, l.Close())

		// Both keys are acquirable again from a fresh locker.
		l2 := newTestLocker(t, client)
		t.Cleanup(func() { _ = l2.Close() })

		_, err = l2.Acquire(ctx, keyA, time.Minute)
		require.NoError(t, err)
		_, err = l2.Acquire(ctx, keyB, time.Minute)
		require.NoError(t, err)
	})

	T.Run("Ping success", func(t *testing.T) {
		t.Parallel()
		l := newTestLocker(t, client)
		require.NoError(t, l.Ping(t.Context()))
	})
}
