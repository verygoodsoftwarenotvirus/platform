package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrUserAlreadyExists(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, ErrUserAlreadyExists)
		assert.Contains(t, ErrUserAlreadyExists.Error(), "user already exists")
	})
}

func TestErrDatabaseNotReady(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, ErrDatabaseNotReady)
		assert.Contains(t, ErrDatabaseNotReady.Error(), "database is not ready yet")
	})
}
