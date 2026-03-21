package circuitbreaking

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/metrics"

	"github.com/samber/do/v2"
)

// RegisterCircuitBreaker registers a CircuitBreaker with the injector.
func RegisterCircuitBreaker(i do.Injector) {
	do.Provide[CircuitBreaker](i, func(i do.Injector) (CircuitBreaker, error) {
		return ProvideCircuitBreaker(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*Config](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}
