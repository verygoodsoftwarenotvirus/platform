package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewMockDatabase(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewMockDatabase())
	})
}

func TestMockDatabase_Migrate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := NewMockDatabase()
		m.On("Migrate", mock.Anything).Return(nil)

		assert.NoError(t, m.Migrate(context.Background()))
	})
}

func TestMockDatabase_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := NewMockDatabase()
		m.On("Close").Return()

		m.Close()
		m.AssertExpectations(t)
	})
}

func TestMockDatabase_DB(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := NewMockDatabase()
		m.On("DB").Return((*sql.DB)(nil))

		assert.Nil(t, m.DB())
	})
}

func TestMockDatabase_ReadDB(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := NewMockDatabase()
		m.On("ReadDB").Return((*sql.DB)(nil))

		assert.Nil(t, m.ReadDB())
	})
}

func TestMockDatabase_WriteDB(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := NewMockDatabase()
		m.On("WriteDB").Return((*sql.DB)(nil))

		assert.Nil(t, m.WriteDB())
	})
}

func TestMockDatabase_IsReady(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := NewMockDatabase()
		m.On("IsReady", mock.Anything).Return(true)

		assert.True(t, m.IsReady(context.Background()))
	})
}

func TestMockDatabase_BeginTx(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := NewMockDatabase()
		m.On("BeginTx", mock.Anything, mock.Anything).Return((*sql.Tx)(nil), nil)

		tx, err := m.BeginTx(context.Background(), nil)
		assert.NoError(t, err)
		assert.Nil(t, tx)
	})
}

func TestMockResultIterator_Scan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockResultIterator{}
		m.On("Scan").Return(nil)

		assert.NoError(t, m.Scan())
	})
}

func TestMockResultIterator_Next(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockResultIterator{}
		m.On("Next").Return(true)

		assert.True(t, m.Next())
	})
}

func TestMockResultIterator_Err(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockResultIterator{}
		m.On("Err").Return(nil)

		assert.NoError(t, m.Err())
	})
}

func TestMockResultIterator_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockResultIterator{}
		m.On("Close").Return(nil)

		assert.NoError(t, m.Close())
	})
}

func TestMockSQLResult_LastInsertId(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockSQLResult{}
		m.On("LastInsertId").Return(int64(1), nil)

		id, err := m.LastInsertId()
		assert.Equal(t, int64(1), id)
		assert.NoError(t, err)
	})
}

func TestMockSQLResult_RowsAffected(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockSQLResult{}
		m.On("RowsAffected").Return(int64(5), nil)

		count, err := m.RowsAffected()
		assert.Equal(t, int64(5), count)
		assert.NoError(t, err)
	})
}

func TestMockQueryExecutor_ExecContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mockResult := &MockSQLResult{}
		m := &MockQueryExecutor{}
		m.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)

		result, err := m.ExecContext(context.Background(), "SELECT 1")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestMockQueryExecutor_PrepareContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockQueryExecutor{}
		m.On("PrepareContext", mock.Anything, mock.Anything).Return((*sql.Stmt)(nil), nil)

		stmt, err := m.PrepareContext(context.Background(), "SELECT 1")
		assert.NoError(t, err)
		assert.Nil(t, stmt)
	})
}

func TestMockQueryExecutor_QueryContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockQueryExecutor{}
		m.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Return((*sql.Rows)(nil), nil)

		rows, err := m.QueryContext(context.Background(), "SELECT 1")
		assert.NoError(t, err)
		assert.Nil(t, rows)
	})
}

func TestMockQueryExecutor_QueryRowContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockQueryExecutor{}
		m.On("QueryRowContext", mock.Anything, mock.Anything, mock.Anything).Return((*sql.Row)(nil))

		row := m.QueryRowContext(context.Background(), "SELECT 1")
		assert.Nil(t, row)
	})
}

func TestMockClient_ReadDB(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockClient{}

		assert.Nil(t, m.ReadDB())
	})
}

func TestMockClient_WriteDB(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockClient{}

		assert.Nil(t, m.WriteDB())
	})
}

func TestMockClient_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockClient{}
		m.On("Close").Return(nil)

		assert.NoError(t, m.Close())
	})
}

func TestMockClient_CurrentTime(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		m := &MockClient{}
		m.On("CurrentTime").Return(now)

		assert.Equal(t, now, m.CurrentTime())
	})
}

func TestMockClient_RollbackTransaction(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &MockClient{}
		m.On("RollbackTransaction", mock.Anything, mock.Anything).Return()

		m.RollbackTransaction(context.Background(), nil)
		m.AssertExpectations(t)
	})
}
