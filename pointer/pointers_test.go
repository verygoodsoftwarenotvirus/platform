package pointer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTo(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := "things"
		actual := new(expected)

		require.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})
}

func TestDereference(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		rawExpected := "things"
		expected := &rawExpected
		actual := Dereference(expected)

		require.NotNil(t, actual)
		assert.Equal(t, rawExpected, actual)
	})
}
