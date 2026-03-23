package analytics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoopEventReporter(T *testing.T) {
	T.Parallel()

	T.Run("returns non-nil reporter", func(t *testing.T) {
		t.Parallel()

		r := NewNoopEventReporter()
		require.NotNil(t, r)
	})
}

func TestNoopEventReporter_Close(T *testing.T) {
	T.Parallel()

	T.Run("does not panic", func(t *testing.T) {
		t.Parallel()

		r := NewNoopEventReporter()
		assert.NotPanics(t, func() {
			r.Close()
		})
	})
}

func TestNoopEventReporter_AddUser(T *testing.T) {
	T.Parallel()

	T.Run("returns nil", func(t *testing.T) {
		t.Parallel()

		r := NewNoopEventReporter()
		err := r.AddUser(context.Background(), "user123", map[string]any{"key": "value"})
		assert.NoError(t, err)
	})
}

func TestNoopEventReporter_EventOccurred(T *testing.T) {
	T.Parallel()

	T.Run("returns nil", func(t *testing.T) {
		t.Parallel()

		r := NewNoopEventReporter()
		err := r.EventOccurred(context.Background(), "event_name", "user123", map[string]any{"key": "value"})
		assert.NoError(t, err)
	})
}

func TestNoopEventReporter_EventOccurredAnonymous(T *testing.T) {
	T.Parallel()

	T.Run("returns nil", func(t *testing.T) {
		t.Parallel()

		r := NewNoopEventReporter()
		err := r.EventOccurredAnonymous(context.Background(), "event_name", "anon123", map[string]any{"key": "value"})
		assert.NoError(t, err)
	})
}
