package qrcodes

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterBuilder(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, Issuer(t.Name()))
		do.ProvideValue(i, tracing.NewNoopTracerProvider())
		do.ProvideValue(i, logging.NewNoopLogger())

		RegisterBuilder(i)

		b, err := do.Invoke[Builder](i)
		require.NoError(t, err)
		assert.NotNil(t, b)
	})
}
