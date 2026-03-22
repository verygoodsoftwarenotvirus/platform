package internalerrors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNilConfigError(T *testing.T) {
	T.Parallel()

	T.Run("returns error with name", func(t *testing.T) {
		t.Parallel()
		err := NilConfigError("redis")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "redis")
		assert.Contains(t, err.Error(), "nil config")
	})

	T.Run("returns different errors for different names", func(t *testing.T) {
		t.Parallel()
		err1 := NilConfigError("redis")
		err2 := NilConfigError("postgres")
		assert.NotEqual(t, err1.Error(), err2.Error())
		assert.Contains(t, err2.Error(), "postgres")
	})
}
