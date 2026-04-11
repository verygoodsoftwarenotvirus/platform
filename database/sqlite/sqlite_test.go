package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/database"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testClientConfig is a test implementation of database.ClientConfig.
type testClientConfig struct {
	readConnectionString  string
	writeConnectionString string
	connectionString      string
	maxPingAttempts       uint64
	pingWaitPeriod        time.Duration
}

var _ database.ClientConfig = (*testClientConfig)(nil)

func (c *testClientConfig) GetReadConnectionString() string {
	if c.readConnectionString != "" {
		return c.readConnectionString
	}
	return c.connectionString
}

func (c *testClientConfig) GetWriteConnectionString() string {
	if c.writeConnectionString != "" {
		return c.writeConnectionString
	}
	return c.connectionString
}

func (c *testClientConfig) GetMaxPingAttempts() uint64 {
	return c.maxPingAttempts
}

func (c *testClientConfig) GetPingWaitPeriod() time.Duration {
	return c.pingWaitPeriod
}

func (c *testClientConfig) GetMaxIdleConns() int {
	return 5
}

func (c *testClientConfig) GetMaxOpenConns() int {
	return 7
}

func (c *testClientConfig) GetConnMaxLifetime() time.Duration {
	return 30 * time.Minute
}

func buildTestClient(t *testing.T) (*Client, sqlmock.Sqlmock) {
	t.Helper()

	fakeDB, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	c := &Client{
		readDB:  fakeDB,
		writeDB: fakeDB,
		config: &testClientConfig{
			maxPingAttempts: 1,
			pingWaitPeriod:  time.Second,
		},
		logger:   logging.NewNoopLogger(),
		timeFunc: defaultTimeFunc,
		tracer:   tracing.NewTracerForTest("test"),
	}

	return c, sqlMock
}

// end helper funcs

func TestQuerier_IsReady(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c, db := buildTestClient(t)
		c.config = &testClientConfig{pingWaitPeriod: time.Second, maxPingAttempts: 1}

		// same DB for read/write, so only one ping
		db.ExpectPing().WillDelayFor(0)

		assert.True(t, c.IsReady(ctx))
	})

	T.Run("with read DB ping error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c, db := buildTestClient(t)
		c.config = &testClientConfig{pingWaitPeriod: time.Millisecond, maxPingAttempts: 1}

		db.ExpectPing().WillReturnError(errors.New("blah"))

		assert.False(t, c.IsReady(ctx))
	})

	T.Run("with write DB ping error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		readDB, readMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)

		writeDB, writeMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)

		c := &Client{
			readDB:  readDB,
			writeDB: writeDB,
			config:  &testClientConfig{pingWaitPeriod: time.Millisecond, maxPingAttempts: 1},
			logger:  logging.NewNoopLogger(),
			tracer:  tracing.NewTracerForTest("test"),
		}

		readMock.ExpectPing().WillDelayFor(0)
		writeMock.ExpectPing().WillReturnError(errors.New("blah"))

		assert.False(t, c.IsReady(ctx))
	})

	T.Run("exhausting all available queries", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		c, db := buildTestClient(t)
		c.config = &testClientConfig{pingWaitPeriod: time.Millisecond, maxPingAttempts: 1}

		db.ExpectPing().WillReturnError(errors.New("blah"))

		assert.False(t, c.IsReady(ctx))
	})
}

func TestProvideDatabaseClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleConfig := &testClientConfig{
			connectionString: ":memory:",
			maxPingAttempts:  1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), exampleConfig, nil)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with no connection strings", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleConfig := &testClientConfig{}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), exampleConfig, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with only read connection string", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleConfig := &testClientConfig{
			readConnectionString: ":memory:",
			maxPingAttempts:      1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), exampleConfig, nil)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with only write connection string", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleConfig := &testClientConfig{
			writeConnectionString: ":memory:",
			maxPingAttempts:       1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), exampleConfig, nil)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with metrics provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleConfig := &testClientConfig{
			connectionString: ":memory:",
			maxPingAttempts:  1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), exampleConfig, metrics.NewNoopMetricsProvider())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with metrics provider and single connection", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleConfig := &testClientConfig{
			readConnectionString: ":memory:",
			maxPingAttempts:      1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), exampleConfig, metrics.NewNoopMetricsProvider())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

func TestDefaultTimeFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotZero(t, defaultTimeFunc())
	})
}

func TestQuerier_currentTime(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, _ := buildTestClient(t)

		assert.NotEmpty(t, c.CurrentTime())
	})

	T.Run("handles nil", func(t *testing.T) {
		t.Parallel()

		var c *Client

		assert.NotEmpty(t, c.CurrentTime())
	})
}

func TestQuerier_rollbackTransaction(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c, db := buildTestClient(t)

		db.ExpectBegin()
		db.ExpectRollback().WillReturnError(errors.New("blah"))

		tx, err := c.writeDB.BeginTx(ctx, nil)
		require.NoError(t, err)

		c.RollbackTransaction(ctx, tx)
	})

	T.Run("with successful rollback", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c, db := buildTestClient(t)

		db.ExpectBegin()
		db.ExpectRollback()

		tx, err := c.writeDB.BeginTx(ctx, nil)
		require.NoError(t, err)

		c.RollbackTransaction(ctx, tx)
	})
}

func TestClient_ReadDB(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, _ := buildTestClient(t)

		assert.NotNil(t, c.ReadDB())
	})
}

func TestClient_WriteDB(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, _ := buildTestClient(t)

		assert.NotNil(t, c.WriteDB())
	})
}

func TestClient_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, db := buildTestClient(t)

		db.ExpectClose()

		assert.NoError(t, c.Close())
	})

	T.Run("with separate read and write DBs", func(t *testing.T) {
		t.Parallel()

		readDB, readMock, err := sqlmock.New()
		require.NoError(t, err)

		writeDB, writeMock, err := sqlmock.New()
		require.NoError(t, err)

		c := &Client{
			readDB:  readDB,
			writeDB: writeDB,
			logger:  logging.NewNoopLogger(),
			tracer:  tracing.NewTracerForTest("test"),
		}

		readMock.ExpectClose()
		writeMock.ExpectClose()

		assert.NoError(t, c.Close())
	})

	T.Run("with read close error", func(t *testing.T) {
		t.Parallel()

		c, db := buildTestClient(t)

		db.ExpectClose().WillReturnError(errors.New("blah"))

		assert.Error(t, c.Close())
	})

	T.Run("with write close error", func(t *testing.T) {
		t.Parallel()

		readDB, readMock, err := sqlmock.New()
		require.NoError(t, err)

		writeDB, writeMock, err := sqlmock.New()
		require.NoError(t, err)

		c := &Client{
			readDB:  readDB,
			writeDB: writeDB,
			logger:  logging.NewNoopLogger(),
			tracer:  tracing.NewTracerForTest("test"),
		}

		readMock.ExpectClose()
		writeMock.ExpectClose().WillReturnError(errors.New("blah"))

		assert.Error(t, c.Close())
	})
}
