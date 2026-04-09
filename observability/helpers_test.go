package observability

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func TestObserveValues(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		_, span := tracing.StartSpan(ctx)

		result := ObserveValues(map[string]any{"key": "value", "other": 123}, span, logger)
		assert.NotNil(t, result)
	})

	T.Run("with nil span", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		result := ObserveValues(map[string]any{"key": "value"}, nil, logger)
		assert.NotNil(t, result)
	})

	T.Run("with nil logger", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		_, span := tracing.StartSpan(ctx)

		result := ObserveValues(map[string]any{"key": "value"}, span, nil)
		assert.Nil(t, result)
	})

	T.Run("with nil span and nil logger", func(t *testing.T) {
		t.Parallel()

		result := ObserveValues(map[string]any{"key": "value"}, nil, nil)
		assert.Nil(t, result)
	})

	T.Run("with empty values", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		_, span := tracing.StartSpan(ctx)

		result := ObserveValues(map[string]any{}, span, logger)
		assert.NotNil(t, result)
	})
}
