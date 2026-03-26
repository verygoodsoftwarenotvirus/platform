package config

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v4/notifications/mobile"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
)

// ProvidePushSender provides a PushNotificationSender from config.
func ProvidePushSender(
	ctx context.Context,
	cfg Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) (mobile.PushNotificationSender, error) {
	return (&cfg).ProvidePushSender(ctx, logger, tracerProvider)
}
