package config

import (
	"github.com/verygoodsoftwarenotvirus/platform/v3/capitalism"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPaymentManager registers a capitalism.PaymentManager with the injector.
func RegisterPaymentManager(i do.Injector) {
	do.Provide[capitalism.PaymentManager](i, func(i do.Injector) (capitalism.PaymentManager, error) {
		return ProvideCapitalismImplementation(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[*Config](i),
		)
	})
}
