package http

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v5/routing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterHTTPServer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, Config{Port: 8080, StartupDeadline: time.Second})
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue[routing.Router](i, nil)
		do.ProvideValue(i, tracing.NewNoopTracerProvider())

		RegisterHTTPServer(i, "test_service")

		srv, err := do.Invoke[Server](i)
		require.NoError(t, err)
		assert.NotNil(t, srv)
	})
}
