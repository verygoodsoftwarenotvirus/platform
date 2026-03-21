package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/stretchr/testify/mock"
)

// NewMockDatabase builds a mock database.
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{}
}

// MockDatabase is our mock database structure. Note, when using this in tests, you must directly access the type name of all the implicit fields.
// So `mockDB.On(reflection.GetMethodName(mockDB.GetUserByUsername)...)` is destined to fail, whereas `mockDB.UserDataManagerMock.On(reflection.GetMethodName(UserDataManagerMock.GetUserByUsername)...)` would do what you want it to do.
type MockDatabase struct {
	mock.Mock
}

// Migrate satisfies the DataManager interface.
func (m *MockDatabase) Migrate(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// Close satisfies the DataManager interface.
func (m *MockDatabase) Close() {
	m.Called()
}

// DB satisfies the DataManager interface.
func (m *MockDatabase) DB() *sql.DB {
	return m.Called().Get(0).(*sql.DB)
}

// ReadDB satisfies the DataManager interface.
func (m *MockDatabase) ReadDB() *sql.DB {
	return m.Called().Get(0).(*sql.DB)
}

// WriteDB satisfies the DataManager interface.
func (m *MockDatabase) WriteDB() *sql.DB {
	return m.Called().Get(0).(*sql.DB)
}

// IsReady satisfies the DataManager interface.
func (m *MockDatabase) IsReady(ctx context.Context) (ready bool) {
	return m.Called(ctx).Bool(0)
}

// BeginTx satisfies the DataManager interface.
func (m *MockDatabase) BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(*sql.Tx), args.Error(1)
}

var _ ResultIterator = (*MockResultIterator)(nil)

// MockResultIterator is our mock sql.Rows structure.
type MockResultIterator struct {
	mock.Mock
}

// Scan satisfies the ResultIterator interface.
func (m *MockResultIterator) Scan(dest ...any) error {
	return m.Called(dest...).Error(0)
}

// Next satisfies the ResultIterator interface.
func (m *MockResultIterator) Next() bool {
	return m.Called().Bool(0)
}

// Err satisfies the ResultIterator interface.
func (m *MockResultIterator) Err() error {
	return m.Called().Error(0)
}

// Close satisfies the ResultIterator interface.
func (m *MockResultIterator) Close() error {
	return m.Called().Error(0)
}

// MockSQLResult mocks a sql.Result.
type MockSQLResult struct {
	mock.Mock
}

// LastInsertId implements our interface.
func (m *MockSQLResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// RowsAffected implements our interface.
func (m *MockSQLResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

var _ SQLQueryExecutor = (*MockQueryExecutor)(nil)

// MockQueryExecutor mocks a sql.Tx|DB.
type MockQueryExecutor struct {
	mock.Mock
}

// ExecContext is a mock function.
func (m *MockQueryExecutor) ExecContext(ctx context.Context, query string, queryArgs ...any) (sql.Result, error) {
	args := m.Called(ctx, query, queryArgs)
	return args.Get(0).(sql.Result), args.Error(1)
}

// PrepareContext is a mock function.
func (m *MockQueryExecutor) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*sql.Stmt), args.Error(1)
}

// QueryContext is a mock function.
func (m *MockQueryExecutor) QueryContext(ctx context.Context, query string, queryArgs ...any) (*sql.Rows, error) {
	args := m.Called(ctx, query, queryArgs)
	return args.Get(0).(*sql.Rows), args.Error(1)
}

// QueryRowContext is a mock function.
func (m *MockQueryExecutor) QueryRowContext(ctx context.Context, query string, queryArgs ...any) *sql.Row {
	args := m.Called(ctx, query, queryArgs)
	return args.Get(0).(*sql.Row)
}

type MockClient struct {
	mock.Mock
}

func (m *MockClient) ReadDB() *sql.DB {
	return nil
}

func (m *MockClient) WriteDB() *sql.DB {
	return nil
}

func (m *MockClient) Close() error {
	return m.Called().Error(0)
}

func (m *MockClient) CurrentTime() time.Time {
	return m.Called().Get(0).(time.Time)
}

func (m *MockClient) RollbackTransaction(ctx context.Context, tx SQLQueryExecutorAndTransactionManager) {
	m.Called(ctx, tx)
}
