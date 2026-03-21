package msgconfig

import (
	"github.com/google/wire"
)

var (
	// MessageQueueProviders are what we provide to dependency injection.
	MessageQueueProviders = wire.NewSet(
		ProvideConsumerProvider,
		ProvidePublisherProvider,
	)
)
