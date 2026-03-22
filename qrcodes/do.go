package qrcodes

import (
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterBuilder registers a Builder with the injector.
func RegisterBuilder(i do.Injector) {
	do.Provide[Builder](i, func(i do.Injector) (Builder, error) {
		return NewBuilder(
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[logging.Logger](i),
		), nil
	})
}
