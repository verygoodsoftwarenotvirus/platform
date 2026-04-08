package config

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/otelgrpc"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterMetricsProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, t.Context())
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, &Config{
			ServiceName:       t.Name(),
			CollectorEndpoint: "localhost:4317",
			Otel: &otelgrpc.Config{
				CollectorEndpoint:  "localhost:4317",
				CollectionInterval: 30 * time.Second,
				Insecure:           true,
			},
		})

		RegisterMetricsProvider(i)

		provider, err := do.Invoke[metrics.Provider](i)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})
}
