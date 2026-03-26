package config

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/mobilenotifications"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

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
) (mobilenotifications.PushNotificationSender, error) {
	return (&cfg).ProvidePushSender(ctx, logger, tracerProvider)
}
