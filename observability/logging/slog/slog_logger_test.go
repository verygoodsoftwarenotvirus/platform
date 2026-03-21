package slog

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/trace"
)

func TestNewLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewSlogLogger(logging.DebugLevel))
	})
}

func Test_zerologLogger_WithName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewSlogLogger(logging.DebugLevel)

		assert.NotNil(t, l.WithName(t.Name()))
	})
}

func Test_zerologLogger_SetRequestIDFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewSlogLogger(logging.DebugLevel)

		l.SetRequestIDFunc(func(*http.Request) string {
			return ""
		})
	})
}

func Test_zerologLogger_Info(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewSlogLogger(logging.DebugLevel)

		l.Info(t.Name())
	})
}

func Test_zerologLogger_Debug(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewSlogLogger(logging.DebugLevel)

		l.Debug(t.Name())
	})
}

func Test_zerologLogger_Error(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewSlogLogger(logging.DebugLevel)

		l.Error(t.Name(), errors.New("blah"))
	})
}

func Test_zerologLogger_Clone(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewSlogLogger(logging.DebugLevel)

		assert.NotNil(t, l.Clone())
	})
}

func Test_zerologLogger_WithValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewSlogLogger(logging.DebugLevel)

		assert.NotNil(t, l.WithValue("name", t.Name()))
	})
}

func Test_zerologLogger_WithValues(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewSlogLogger(logging.DebugLevel)

		assert.NotNil(t, l.WithValues(map[string]any{"name": t.Name()}))
	})
}

func Test_zerologLogger_WithError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewSlogLogger(logging.DebugLevel)

		assert.NotNil(t, l.WithError(errors.New("blah")))
	})
}

func Test_zerologLogger_WithSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l := NewSlogLogger(logging.DebugLevel)

		span := trace.SpanFromContext(ctx)

		assert.NotNil(t, l.WithSpan(span))
	})
}

func Test_zerologLogger_WithRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l, ok := NewSlogLogger(logging.DebugLevel).(*slogLogger)
		require.True(t, ok)

		l.requestIDFunc = func(*http.Request) string {
			return t.Name()
		}

		u, err := url.ParseRequestURI("https://whatever.whocares.gov?things=stuff")
		require.NoError(t, err)

		assert.NotNil(t, l.WithRequest(&http.Request{
			URL: u,
		}))
	})
}

func Test_zerologLogger_WithResponse(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewSlogLogger(logging.DebugLevel)

		assert.NotNil(t, l.WithResponse(&http.Response{}))
	})
}
