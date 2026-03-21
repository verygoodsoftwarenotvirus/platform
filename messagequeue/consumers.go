package messagequeue

import (
	"context"
)

type (
	// Consumer produces events onto a queue.
	Consumer interface {
		Consume(stopChan chan bool, errors chan error)
	}

	ConsumerFunc func(context.Context, []byte) error

	// ConsumerProvider is a function that provides a Consumer for a given topic.
	ConsumerProvider interface {
		ProvideConsumer(ctx context.Context, topic string, handlerFunc ConsumerFunc) (Consumer, error)
	}
)
