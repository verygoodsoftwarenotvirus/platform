package emailcfg

import (
	"context"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/platform/v4/email"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterEmailer registers an email.Emailer with the injector.
func RegisterEmailer(i do.Injector) {
	do.Provide[email.Emailer](i, func(i do.Injector) (email.Emailer, error) {
		return ProvideEmailer(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*Config](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[*http.Client](i),
		)
	})
}
