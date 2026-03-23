package pointer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTo(T *testing.T) {
	T.Parallel()

	T.Run("with string", func(t *testing.T) {
		t.Parallel()

		expected := "things"
		actual := To(expected)

		require.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})

	T.Run("with int", func(t *testing.T) {
		t.Parallel()

		expected := 42
		actual := To(expected)

		require.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})

	T.Run("with zero value", func(t *testing.T) {
		t.Parallel()

		actual := To(0)

		require.NotNil(t, actual)
		assert.Equal(t, 0, *actual)
	})

	T.Run("with struct", func(t *testing.T) {
		t.Parallel()

		type example struct{ Name string }
		expected := example{Name: "test"}
		actual := To(expected)

		require.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})
}

func TestToSlice(T *testing.T) {
	T.Parallel()

	T.Run("with string slice", func(t *testing.T) {
		t.Parallel()

		input := []string{"a", "b", "c"}
		actual := ToSlice(input)

		require.Len(t, actual, 3)
		assert.Equal(t, "a", *actual[0])
		assert.Equal(t, "b", *actual[1])
		assert.Equal(t, "c", *actual[2])
	})

	T.Run("with int slice", func(t *testing.T) {
		t.Parallel()

		input := []int{1, 2, 3}
		actual := ToSlice(input)

		require.Len(t, actual, 3)
		assert.Equal(t, 1, *actual[0])
		assert.Equal(t, 2, *actual[1])
		assert.Equal(t, 3, *actual[2])
	})

	T.Run("with nil slice", func(t *testing.T) {
		t.Parallel()

		actual := ToSlice[string](nil)

		assert.NotNil(t, actual)
		assert.Empty(t, actual)
	})

	T.Run("with empty slice", func(t *testing.T) {
		t.Parallel()

		actual := ToSlice([]string{})

		assert.NotNil(t, actual)
		assert.Empty(t, actual)
	})
}

func TestDereference(T *testing.T) {
	T.Parallel()

	T.Run("with string pointer", func(t *testing.T) {
		t.Parallel()

		rawExpected := "things"
		actual := Dereference(&rawExpected)

		assert.Equal(t, rawExpected, actual)
	})

	T.Run("with int pointer", func(t *testing.T) {
		t.Parallel()

		expected := 42
		actual := Dereference(&expected)

		assert.Equal(t, 42, actual)
	})

	T.Run("with nil string pointer", func(t *testing.T) {
		t.Parallel()

		actual := Dereference[string](nil)

		assert.Equal(t, "", actual)
	})

	T.Run("with nil int pointer", func(t *testing.T) {
		t.Parallel()

		actual := Dereference[int](nil)

		assert.Equal(t, 0, actual)
	})

	T.Run("with nil bool pointer", func(t *testing.T) {
		t.Parallel()

		actual := Dereference[bool](nil)

		assert.False(t, actual)
	})
}

func TestDereferenceSlice(T *testing.T) {
	T.Parallel()

	T.Run("with string pointer slice", func(t *testing.T) {
		t.Parallel()

		a, b, c := "a", "b", "c"
		input := []*string{&a, &b, &c}
		actual := DereferenceSlice(input)

		require.Len(t, actual, 3)
		assert.Equal(t, "a", actual[0])
		assert.Equal(t, "b", actual[1])
		assert.Equal(t, "c", actual[2])
	})

	T.Run("with int pointer slice", func(t *testing.T) {
		t.Parallel()

		a, b := 1, 2
		input := []*int{&a, &b}
		actual := DereferenceSlice(input)

		require.Len(t, actual, 2)
		assert.Equal(t, 1, actual[0])
		assert.Equal(t, 2, actual[1])
	})

	T.Run("with nil slice", func(t *testing.T) {
		t.Parallel()

		actual := DereferenceSlice[string](nil)

		assert.NotNil(t, actual)
		assert.Empty(t, actual)
	})

	T.Run("with empty slice", func(t *testing.T) {
		t.Parallel()

		actual := DereferenceSlice([]*string{})

		assert.NotNil(t, actual)
		assert.Empty(t, actual)
	})
}
