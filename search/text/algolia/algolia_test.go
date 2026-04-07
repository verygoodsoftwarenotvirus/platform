package algolia

import (
	"testing"

	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
)

type example struct{}

func TestProvideIndexManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()

		im, err := ProvideIndexManager[example](logger, tracerProvider, &Config{}, "test", cbnoop.NewCircuitBreaker())
		assert.NoError(t, err)
		assert.NotNil(t, im)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()

		im, err := ProvideIndexManager[example](logger, tracerProvider, nil, "test", cbnoop.NewCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, im)
	})
}
