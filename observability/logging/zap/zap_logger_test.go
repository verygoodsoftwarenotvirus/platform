package zap

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestNewLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewZapLogger(logging.DebugLevel))
	})
}

func Test_zapLogger_WithName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZapLogger(logging.DebugLevel)

		assert.NotNil(t, l.WithName(t.Name()))
	})
}

func Test_zapLogger_SetRequestIDFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZapLogger(logging.DebugLevel)

		l.SetRequestIDFunc(func(*http.Request) string {
			return ""
		})
	})
}

func Test_zapLogger_Info(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZapLogger(logging.DebugLevel)

		l.Info(t.Name())
	})
}

func Test_zapLogger_Debug(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZapLogger(logging.DebugLevel)

		l.Debug(t.Name())
	})
}

func Test_zapLogger_Error(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZapLogger(logging.DebugLevel)

		l.Error(t.Name(), errors.New("blah"))
	})
}

func Test_zapLogger_Clone(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZapLogger(logging.DebugLevel)

		assert.NotNil(t, l.Clone())
	})
}

func Test_zapLogger_WithValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZapLogger(logging.DebugLevel)

		assert.NotNil(t, l.WithValue("name", t.Name()))
	})
}

func Test_zapLogger_WithValues(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZapLogger(logging.DebugLevel)

		assert.NotNil(t, l.WithValues(map[string]any{"name": t.Name()}))
	})
}

func Test_zapLogger_WithError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZapLogger(logging.DebugLevel)

		assert.NotNil(t, l.WithError(errors.New("blah")))
	})
}

func Test_zapLogger_WithSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l := NewZapLogger(logging.DebugLevel)

		span := trace.SpanFromContext(ctx)

		assert.NotNil(t, l.WithSpan(span))
	})
}

func Test_zapLogger_WithRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l, ok := NewZapLogger(logging.DebugLevel).(*zapLogger)
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

func Test_zapLogger_WithResponse(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		l := NewZapLogger(logging.DebugLevel)

		assert.NotNil(t, l.WithResponse(&http.Response{}))
	})
}
