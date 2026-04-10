package encoding

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterServerEncoderDecoder(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, Config{ContentType: "application/json"})
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, tracing.NewNoopTracerProvider())

		RegisterServerEncoderDecoder(i)

		ct, err := do.Invoke[ContentType](i)
		require.NoError(t, err)
		assert.NotNil(t, ct)

		sed, err := do.Invoke[ServerEncoderDecoder](i)
		require.NoError(t, err)
		assert.NotNil(t, sed)
	})
}
