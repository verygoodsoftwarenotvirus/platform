package algolia

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/stretchr/testify/assert"
)

type example struct{}

func TestProvideIndexManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()

		im, err := ProvideIndexManager[example](logger, tracerProvider, &Config{}, "test", circuitbreaking.NewNoopCircuitBreaker())
		assert.NoError(t, err)
		assert.NotNil(t, im)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()

		im, err := ProvideIndexManager[example](logger, tracerProvider, nil, "test", circuitbreaking.NewNoopCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, im)
	})
}
