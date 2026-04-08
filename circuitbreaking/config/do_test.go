package circuitbreakingcfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // race condition in the core circuit breaker library, I think?
func TestRegisterCircuitBreaker(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		cfg := &Config{}
		cfg.EnsureDefaults()

		i := do.New()
		do.ProvideValue(i, t.Context())
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, metrics.NewNoopMetricsProvider())
		do.ProvideValue(i, cfg)

		RegisterCircuitBreaker(i)

		cb, err := do.Invoke[circuitbreaking.CircuitBreaker](i)
		require.NoError(t, err)
		assert.NotNil(t, cb)
	})
}
