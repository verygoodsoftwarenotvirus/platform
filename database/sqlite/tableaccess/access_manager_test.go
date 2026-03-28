package tableaccess

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()

		m := NewManager()
		err := m.CreateUser(t.Context(), "user", "pass")
		assert.ErrorIs(t, err, ErrNotSupported)
	})
}

func TestManager_DeleteUser(T *testing.T) {
	T.Parallel()

	T.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()

		m := NewManager()
		err := m.DeleteUser(t.Context(), "user")
		assert.ErrorIs(t, err, ErrNotSupported)
	})
}

func TestManager_CreateDatabase(T *testing.T) {
	T.Parallel()

	T.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()

		m := NewManager()
		err := m.CreateDatabase(t.Context(), "db", "owner")
		assert.ErrorIs(t, err, ErrNotSupported)
	})
}

func TestManager_DeleteDatabase(T *testing.T) {
	T.Parallel()

	T.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()

		m := NewManager()
		err := m.DeleteDatabase(t.Context(), "db")
		assert.ErrorIs(t, err, ErrNotSupported)
	})
}

func TestManager_UserExists(T *testing.T) {
	T.Parallel()

	T.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()

		m := NewManager()
		exists, err := m.UserExists(t.Context(), "user")
		assert.False(t, exists)
		assert.ErrorIs(t, err, ErrNotSupported)
	})
}

func TestManager_DatabaseExists(T *testing.T) {
	T.Parallel()

	T.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()

		m := NewManager()
		exists, err := m.DatabaseExists(t.Context(), "db")
		assert.False(t, exists)
		assert.ErrorIs(t, err, ErrNotSupported)
	})
}

func TestManager_GrantUserAccessToTable(T *testing.T) {
	T.Parallel()

	T.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()

		m := NewManager()
		err := m.GrantUserAccessToTable(t.Context(), "user", "schema", "table", "SELECT")
		assert.ErrorIs(t, err, ErrNotSupported)
	})
}

func TestManager_UserCanAccessDatabase(T *testing.T) {
	T.Parallel()

	T.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()

		m := NewManager()
		canAccess, err := m.UserCanAccessDatabase(t.Context(), "user", "db")
		assert.False(t, canAccess)
		assert.ErrorIs(t, err, ErrNotSupported)
	})
}
