package messagequeue

import (
	"context"
)

var _ PublisherProvider = (*NoopPublisherProvider)(nil)

type NoopPublisherProvider struct{}

func (n *NoopPublisherProvider) Close() {}

func (n *NoopPublisherProvider) ProvidePublisher(context.Context, string) (Publisher, error) {
	return &NoopPublisher{}, nil
}

var _ Publisher = (*NoopPublisher)(nil)

type NoopPublisher struct{}

func (n *NoopPublisher) Stop() {}

func (n *NoopPublisher) Publish(context.Context, any) error {
	return nil
}

func (n *NoopPublisher) PublishAsync(context.Context, any) {}

var _ ConsumerProvider = (*NoopConsumerProvider)(nil)

type NoopConsumerProvider struct{}

func (n *NoopConsumerProvider) ProvideConsumer(context.Context, string, ConsumerFunc) (Consumer, error) {
	return &NoopConsumer{}, nil
}

var _ Consumer = (*NoopConsumer)(nil)

type NoopConsumer struct{}

func (n *NoopConsumer) Consume(chan bool, chan error) {}
