package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, EnsureLogger(NewNoopLogger()))
	})

	T.Run("with nil", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, EnsureLogger(nil))
	})
}
