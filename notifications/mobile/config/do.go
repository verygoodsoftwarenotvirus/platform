package config

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v4/notifications/mobile"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPushSender registers a mobile.PushNotificationSender with the injector.
func RegisterPushSender(i do.Injector) {
	do.Provide[mobile.PushNotificationSender](i, func(i do.Injector) (mobile.PushNotificationSender, error) {
		return ProvidePushSender(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[Config](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
		)
	})
}
