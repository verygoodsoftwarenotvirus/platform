package fake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type example struct {
	Name string
	Age  int
}

func TestBuildFakeForTest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual := BuildFakeForTest[*example](t)
		assert.NotNil(t, actual)
	})
}

func TestMustBuildFake(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotPanics(t, func() {
			actual := MustBuildFake[example]()
			assert.NotEmpty(t, actual.Name)
			assert.NotEmpty(t, actual.Age)
		})
	})

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.Panics(t, func() {
			MustBuildFake[any]()
		})
	})
}

func TestBuildFake(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual, err := BuildFake[string]()
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})

	T.Run("with error", func(t *testing.T) {
		t.Parallel()

		actual, err := BuildFake[any]()
		assert.Error(t, err)
		assert.Empty(t, actual)
	})
}
