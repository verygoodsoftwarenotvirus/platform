package logging

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestAllLevels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		levels := AllLevels()
		assert.NotEmpty(t, levels)
		assert.Contains(t, levels, InfoLevel)
		assert.Contains(t, levels, DebugLevel)
		assert.Contains(t, levels, ErrorLevel)
		assert.Contains(t, levels, WarnLevel)
	})
}

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

func TestNewNamedLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewNamedLogger(NewNoopLogger(), "test"))
	})

	T.Run("with nil logger", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewNamedLogger(nil, "test"))
	})
}

func TestNoopLogger(T *testing.T) {
	T.Parallel()

	T.Run("NewNoopLogger", func(t *testing.T) {
		t.Parallel()

		l := NewNoopLogger()
		assert.NotNil(t, l)
	})

	T.Run("Info", func(t *testing.T) {
		t.Parallel()

		NewNoopLogger().Info("test")
	})

	T.Run("Debug", func(t *testing.T) {
		t.Parallel()

		NewNoopLogger().Debug("test")
	})

	T.Run("Error", func(t *testing.T) {
		t.Parallel()

		NewNoopLogger().Error("test", errors.New("blah"))
	})

	T.Run("SetRequestIDFunc", func(t *testing.T) {
		t.Parallel()

		NewNoopLogger().SetRequestIDFunc(func(*http.Request) string { return "" })
	})

	T.Run("WithName", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewNoopLogger().WithName("test"))
	})

	T.Run("Clone", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewNoopLogger().Clone())
	})

	T.Run("WithValues", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewNoopLogger().WithValues(map[string]any{"key": "value"}))
	})

	T.Run("WithValue", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewNoopLogger().WithValue("key", "value"))
	})

	T.Run("WithRequest", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com", http.NoBody)
		require.NoError(t, err)

		assert.NotNil(t, NewNoopLogger().WithRequest(req))
	})

	T.Run("WithResponse", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewNoopLogger().WithResponse(&http.Response{}))
	})

	T.Run("WithError", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewNoopLogger().WithError(errors.New("blah")))
	})

	T.Run("WithSpan", func(t *testing.T) {
		t.Parallel()

		span := noop.NewTracerProvider().Tracer("test").Start
		ctx := t.Context()
		_, s := noop.NewTracerProvider().Tracer("test").Start(ctx, "test")

		_ = span
		assert.NotNil(t, NewNoopLogger().WithSpan(s))
	})
}

func TestExtractSpanInfo(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		_, span := noop.NewTracerProvider().Tracer("test").Start(ctx, "test")

		info := ExtractSpanInfo(span)
		assert.NotEmpty(t, info.SpanID)
		assert.NotEmpty(t, info.TraceID)
	})
}

func TestExtractRequestInfo(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/path?foo=bar", http.NoBody)
		require.NoError(t, err)

		info := ExtractRequestInfo(req, func(r *http.Request) string { return "req-123" })
		assert.Equal(t, http.MethodGet, info.Method)
		assert.Equal(t, "/path", info.Path)
		assert.Equal(t, "foo=bar", info.Query)
		assert.Equal(t, "req-123", info.RequestID)
	})

	T.Run("with nil request", func(t *testing.T) {
		t.Parallel()

		info := ExtractRequestInfo(nil, nil)
		assert.Empty(t, info.Method)
		assert.Empty(t, info.Path)
		assert.Empty(t, info.Query)
		assert.Empty(t, info.RequestID)
	})

	T.Run("with nil URL", func(t *testing.T) {
		t.Parallel()

		req := &http.Request{Method: http.MethodPost}

		info := ExtractRequestInfo(req, nil)
		assert.Equal(t, http.MethodPost, info.Method)
		assert.Empty(t, info.Path)
		assert.Empty(t, info.Query)
	})

	T.Run("with nil requestIDFunc", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/path", http.NoBody)
		require.NoError(t, err)

		info := ExtractRequestInfo(req, nil)
		assert.Equal(t, http.MethodGet, info.Method)
		assert.Empty(t, info.RequestID)
	})
}

var _ Logger = (*noopLogger)(nil)
var _ trace.Span = trace.Span(nil)
