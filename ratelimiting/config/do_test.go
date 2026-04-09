package ratelimitingcfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/ratelimiting"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRateLimiter(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue[metrics.Provider](i, nil)
		do.ProvideValue(i, &Config{})

		RegisterRateLimiter(i)

		limiter, err := do.Invoke[ratelimiting.RateLimiter](i)
		require.NoError(t, err)
		assert.NotNil(t, limiter)
	})
}
