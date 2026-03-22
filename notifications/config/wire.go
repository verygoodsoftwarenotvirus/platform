package config

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v2/notifications"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/google/wire"
)

var (
	// Providers are what we provide to dependency injection.
	Providers = wire.NewSet(
		ProvidePushSender,
	)
)

// ProvidePushSender provides a PushNotificationSender from config.
func ProvidePushSender(
	ctx context.Context,
	cfg Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) (notifications.PushNotificationSender, error) {
	return (&cfg).ProvidePushSender(ctx, logger, tracerProvider)
}
