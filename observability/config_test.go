package observability

import (
	"testing"

	tracingcfg "github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing/config"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing/oteltrace"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Tracing: tracingcfg.Config{
				ServiceName:               t.Name(),
				SpanCollectionProbability: 1,
				Provider:                  tracingcfg.ProviderOtel,
				Otel: &oteltrace.Config{
					CollectorEndpoint: "0.0.0.0",
				},
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}
