package messagequeue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoopPublisherProvider_ProvidePublisher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		p := &NoopPublisherProvider{}
		publisher, err := p.ProvidePublisher(context.Background(), "topic")
		require.NoError(t, err)
		assert.NotNil(t, publisher)
	})
}

func TestNoopPublisherProvider_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		p := &NoopPublisherProvider{}
		p.Close()
	})
}

func TestNoopPublisher_Publish(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		p := &NoopPublisher{}
		err := p.Publish(context.Background(), "data")
		assert.NoError(t, err)
	})
}

func TestNoopPublisher_PublishAsync(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		p := &NoopPublisher{}
		p.PublishAsync(context.Background(), "data")
	})
}

func TestNoopPublisher_Stop(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		p := &NoopPublisher{}
		p.Stop()
	})
}

func TestNoopConsumerProvider_ProvideConsumer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		p := &NoopConsumerProvider{}
		consumer, err := p.ProvideConsumer(context.Background(), "topic", func(_ context.Context, _ []byte) error { return nil })
		require.NoError(t, err)
		assert.NotNil(t, consumer)
	})
}

func TestNoopConsumer_Consume(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c := &NoopConsumer{}
		c.Consume(context.Background(), make(chan bool), make(chan error))
	})
}
