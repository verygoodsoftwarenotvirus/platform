package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/database"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// testClientConfig is a test implementation of database.ClientConfig.
type testClientConfig struct {
	connectionString string
	maxPingAttempts  uint64
	pingWaitPeriod   time.Duration
}

var _ database.ClientConfig = (*testClientConfig)(nil)

func (c *testClientConfig) GetReadConnectionString() string {
	return c.connectionString
}

func (c *testClientConfig) GetWriteConnectionString() string {
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

type sqlmockExpecterWrapper struct {
	sqlmock.Sqlmock
}

func (e *sqlmockExpecterWrapper) AssertExpectations(t mock.TestingT) bool {
	return assert.NoError(t, e.ExpectationsWereMet(), "not all database expectations were met")
}

func buildTestClient(t *testing.T) (*Client, *sqlmockExpecterWrapper) {
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

	return c, &sqlmockExpecterWrapper{Sqlmock: sqlMock}
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
			connectionString: "test:test@tcp(localhost:3306)/test",
			maxPingAttempts:  1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), exampleConfig, nil)
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
}
