package otelgrpc

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestNewLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		assert.NotNil(t, l)
		assert.NoError(t, err)
	})
}

func Test_zerologLogger_WithName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		assert.NotNil(t, l.WithName(t.Name()))
	})
}

func Test_zerologLogger_SetRequestIDFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		l.SetRequestIDFunc(func(*http.Request) string {
			return ""
		})
	})
}

func Test_zerologLogger_Info(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		l.Info(t.Name())
	})
}

func Test_zerologLogger_Debug(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		l.Debug(t.Name())
	})
}

func Test_zerologLogger_Error(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		l.Error(t.Name(), errors.New("blah"))
	})
}

func Test_zerologLogger_Clone(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		assert.NotNil(t, l.Clone())
	})
}

func Test_zerologLogger_WithValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		assert.NotNil(t, l.WithValue("name", t.Name()))
	})
}

func Test_zerologLogger_WithValues(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		assert.NotNil(t, l.WithValues(map[string]any{"name": t.Name()}))
	})
}

func Test_zerologLogger_WithError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		assert.NotNil(t, l.WithError(errors.New("blah")))
	})
}

func Test_zerologLogger_WithSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		span := trace.SpanFromContext(ctx)

		assert.NotNil(t, l.WithSpan(span))
	})
}

func Test_zerologLogger_WithRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		l, ok := logger.(*otelSlogLogger)
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

		ctx := t.Context()
		l, err := NewOtelSlogLogger(ctx, logging.DebugLevel, t.Name(), &Config{})
		require.NoError(t, err)

		assert.NotNil(t, l.WithResponse(&http.Response{}))
	})
}
