package config

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/mobilenotifications"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPushSender registers a mobilenotifications.PushNotificationSender with the injector.
func RegisterPushSender(i do.Injector) {
	do.Provide[mobilenotifications.PushNotificationSender](i, func(i do.Injector) (mobilenotifications.PushNotificationSender, error) {
		return ProvidePushSender(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[Config](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
		)
	})
}
