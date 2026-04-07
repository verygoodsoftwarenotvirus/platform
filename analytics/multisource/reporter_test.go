package multisource

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics"
	analyticsmock "github.com/verygoodsoftwarenotvirus/platform/v5/analytics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewMultiSourceEventReporter(T *testing.T) {
	T.Parallel()

	T.Run("with nil reporters map", func(t *testing.T) {
		t.Parallel()

		r := NewMultiSourceEventReporter(nil, nil, nil)
		require.NotNil(t, r)
		assert.NotNil(t, r.reporters)
	})

	T.Run("with populated reporters map", func(t *testing.T) {
		t.Parallel()

		reporters := map[string]analytics.EventReporter{
			"ios": noop.NewEventReporter(),
		}
		r := NewMultiSourceEventReporter(reporters, nil, nil)
		require.NotNil(t, r)
		assert.Len(t, r.reporters, 1)
	})
}

func TestMultiSourceEventReporter_getReporter(T *testing.T) {
	T.Parallel()

	T.Run("returns reporter for known source", func(t *testing.T) {
		t.Parallel()

		expected := noop.NewEventReporter()
		reporters := map[string]analytics.EventReporter{
			"ios": expected,
		}
		m := NewMultiSourceEventReporter(reporters, nil, nil)

		got := m.getReporter("ios")
		assert.Equal(t, expected, got)
	})

	T.Run("returns noop for unknown source", func(t *testing.T) {
		t.Parallel()

		m := NewMultiSourceEventReporter(nil, nil, nil)

		got := m.getReporter("unknown")
		assert.NotNil(t, got)
	})

	T.Run("returns noop when reporter is nil in map", func(t *testing.T) {
		t.Parallel()

		reporters := map[string]analytics.EventReporter{
			"ios": nil,
		}
		m := NewMultiSourceEventReporter(reporters, nil, nil)

		got := m.getReporter("ios")
		assert.NotNil(t, got)
	})
}

func TestMultiSourceEventReporter_TrackEvent(T *testing.T) {
	T.Parallel()

	T.Run("delegates to correct reporter", func(t *testing.T) {
		t.Parallel()

		mockReporter := &analyticsmock.EventReporter{}
		mockReporter.On("EventOccurred", mock.AnythingOfType("*context.valueCtx"), "signup", "user1", mock.MatchedBy(func(props map[string]any) bool {
			return props[SourcePropertyKey] == "ios" && props["plan"] == "pro"
		})).Return(nil)

		reporters := map[string]analytics.EventReporter{
			"ios": mockReporter,
		}
		m := NewMultiSourceEventReporter(reporters, nil, nil)

		err := m.TrackEvent(context.Background(), "ios", "signup", "user1", map[string]any{"plan": "pro"})
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockReporter)
	})

	T.Run("uses noop for unknown source", func(t *testing.T) {
		t.Parallel()

		m := NewMultiSourceEventReporter(nil, nil, nil)

		err := m.TrackEvent(context.Background(), "unknown", "signup", "user1", nil)
		assert.NoError(t, err)
	})
}

func TestMultiSourceEventReporter_TrackAnonymousEvent(T *testing.T) {
	T.Parallel()

	T.Run("delegates to correct reporter", func(t *testing.T) {
		t.Parallel()

		mockReporter := &analyticsmock.EventReporter{}
		mockReporter.On("EventOccurredAnonymous", mock.AnythingOfType("*context.valueCtx"), "page_view", "anon1", mock.MatchedBy(func(props map[string]any) bool {
			return props[SourcePropertyKey] == "web"
		})).Return(nil)

		reporters := map[string]analytics.EventReporter{
			"web": mockReporter,
		}
		m := NewMultiSourceEventReporter(reporters, nil, nil)

		err := m.TrackAnonymousEvent(context.Background(), "web", "page_view", "anon1", map[string]any{})
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockReporter)
	})
}

func Test_withSourceProperty(T *testing.T) {
	T.Parallel()

	T.Run("adds source to nil properties", func(t *testing.T) {
		t.Parallel()

		result := withSourceProperty("ios", nil)
		assert.Equal(t, "ios", result[SourcePropertyKey])
		assert.Len(t, result, 1)
	})

	T.Run("adds source to existing properties without mutation", func(t *testing.T) {
		t.Parallel()

		original := map[string]any{"key": "value"}
		result := withSourceProperty("web", original)

		assert.Equal(t, "web", result[SourcePropertyKey])
		assert.Equal(t, "value", result["key"])
		assert.Len(t, result, 2)

		// original should not be mutated
		_, exists := original[SourcePropertyKey]
		assert.False(t, exists)
	})
}
