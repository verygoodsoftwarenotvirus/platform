package config

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/eventstream"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterEventStreamUpgrader(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, tracing.NewNoopTracerProvider())
		do.ProvideValue(i, &Config{Provider: ProviderSSE})

		RegisterEventStreamUpgrader(i)

		upgrader, err := do.Invoke[eventstream.EventStreamUpgrader](i)
		require.NoError(t, err)
		assert.NotNil(t, upgrader)
	})
}

func TestRegisterBidirectionalEventStreamUpgrader(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, tracing.NewNoopTracerProvider())
		do.ProvideValue(i, &Config{Provider: ProviderWebSocket})

		RegisterBidirectionalEventStreamUpgrader(i)

		upgrader, err := do.Invoke[eventstream.BidirectionalEventStreamUpgrader](i)
		require.NoError(t, err)
		assert.NotNil(t, upgrader)
	})
}
