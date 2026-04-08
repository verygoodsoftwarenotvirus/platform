package routingcfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/routing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRouteParamManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, &Config{
			Provider: ProviderChi,
		})

		RegisterRouteParamManager(i)

		manager, err := do.Invoke[routing.RouteParamManager](i)
		require.NoError(t, err)
		assert.NotNil(t, manager)
	})
}
