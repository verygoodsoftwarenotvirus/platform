package msgconfig

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterMessageQueue registers both messagequeue.ConsumerProvider and
// messagequeue.PublisherProvider with the injector.
func RegisterMessageQueue(i do.Injector) {
	do.Provide[messagequeue.ConsumerProvider](i, func(i do.Injector) (messagequeue.ConsumerProvider, error) {
		return ProvideConsumerProvider(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[*Config](i),
		)
	})
	do.Provide[messagequeue.PublisherProvider](i, func(i do.Injector) (messagequeue.PublisherProvider, error) {
		return ProvidePublisherProvider(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[*Config](i),
		)
	})
}
