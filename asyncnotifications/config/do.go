package asyncnotificationscfg

import (
	"github.com/verygoodsoftwarenotvirus/platform/v2/asyncnotifications"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAsyncNotifier registers an asyncnotifications.AsyncNotifier with the injector.
func RegisterAsyncNotifier(i do.Injector) {
	do.Provide[asyncnotifications.AsyncNotifier](i, func(i do.Injector) (asyncnotifications.AsyncNotifier, error) {
		return ProvideAsyncNotifierFromConfig(
			do.MustInvoke[*Config](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
		)
	})
}
