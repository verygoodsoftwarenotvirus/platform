package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentinelErrors(T *testing.T) {
	T.Parallel()

	T.Run("ErrNilInputParameter", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, ErrNilInputParameter)
		assert.Contains(t, ErrNilInputParameter.Error(), "nil")
	})

	T.Run("ErrEmptyInputParameter", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, ErrEmptyInputParameter)
		assert.Contains(t, ErrEmptyInputParameter.Error(), "empty")
	})

	T.Run("ErrNilInputProvided", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, ErrNilInputProvided)
		assert.Contains(t, ErrNilInputProvided.Error(), "nil input")
	})

	T.Run("ErrInvalidIDProvided", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, ErrInvalidIDProvided)
		assert.Contains(t, ErrInvalidIDProvided.Error(), "ID")
	})

	T.Run("ErrEmptyInputProvided", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, ErrEmptyInputProvided)
		assert.Contains(t, ErrEmptyInputProvided.Error(), "empty")
	})

	T.Run("sentinels are distinct", func(t *testing.T) {
		t.Parallel()
		assert.False(t, errors.Is(ErrNilInputParameter, ErrEmptyInputParameter))
		assert.False(t, errors.Is(ErrNilInputProvided, ErrInvalidIDProvided))
		assert.False(t, errors.Is(ErrEmptyInputProvided, ErrNilInputProvided))
	})
}

func TestNew(T *testing.T) {
	T.Parallel()

	T.Run("creates error with message", func(t *testing.T) {
		t.Parallel()
		err := New("test error")
		require.Error(t, err)
		assert.Equal(t, "test error", err.Error())
	})
}

func TestNewf(T *testing.T) {
	T.Parallel()

	T.Run("creates formatted error", func(t *testing.T) {
		t.Parallel()
		err := Newf("error %d: %s", 42, "details")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "42")
		assert.Contains(t, err.Error(), "details")
	})
}

func TestErrorf(T *testing.T) {
	T.Parallel()

	T.Run("creates formatted error", func(t *testing.T) {
		t.Parallel()
		err := Errorf("something %s", "failed")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "something failed")
	})
}

func TestWrap(T *testing.T) {
	T.Parallel()

	T.Run("wraps error with message", func(t *testing.T) {
		t.Parallel()
		inner := fmt.Errorf("inner")
		wrapped := Wrap(inner, "outer")
		require.Error(t, wrapped)
		assert.True(t, errors.Is(wrapped, inner))
		assert.Contains(t, wrapped.Error(), "outer")
	})

	T.Run("nil error returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, Wrap(nil, "outer"))
	})
}

func TestWrapf(T *testing.T) {
	T.Parallel()

	T.Run("wraps error with formatted message", func(t *testing.T) {
		t.Parallel()
		inner := fmt.Errorf("inner")
		wrapped := Wrapf(inner, "outer %d", 1)
		require.Error(t, wrapped)
		assert.True(t, errors.Is(wrapped, inner))
		assert.Contains(t, wrapped.Error(), "outer 1")
	})
}
