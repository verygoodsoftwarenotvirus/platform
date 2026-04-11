package sqlite

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/database"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/samber/do/v2"
	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
)

func TestRegisterDatabaseClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue[context.Context](i, t.Context())
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, tracing.NewNoopTracerProvider())
		do.ProvideValue[metrics.Provider](i, nil)
		do.ProvideValue[database.ClientConfig](i, &testClientConfig{
			connectionString: ":memory:",
			maxPingAttempts:  1,
		})

		RegisterDatabaseClient(i)

		client, err := do.Invoke[database.Client](i)
		must.NoError(t, err)
		test.NotNil(t, client)
	})
}
