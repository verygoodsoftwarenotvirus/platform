package config

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v2/notifications"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPushSender registers a notifications.PushNotificationSender with the injector.
func RegisterPushSender(i do.Injector) {
	do.Provide[notifications.PushNotificationSender](i, func(i do.Injector) (notifications.PushNotificationSender, error) {
		return ProvidePushSender(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[Config](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
		)
	})
}
