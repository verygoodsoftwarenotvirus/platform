package messagequeue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoopPublisher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		p := NewNoopPublisher()
		require.NotNil(t, p)

		assert.NoError(t, p.Publish(context.Background(), "data"))
		p.PublishAsync(context.Background(), "data")
		p.Stop()
	})
}

func TestNewNoopPublisherProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		pp := NewNoopPublisherProvider()
		require.NotNil(t, pp)

		publisher, err := pp.ProvidePublisher(context.Background(), "topic")
		require.NoError(t, err)
		assert.NotNil(t, publisher)

		pp.Close()
	})
}

func TestErrEmptyTopicName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, ErrEmptyTopicName)
		assert.Error(t, ErrEmptyTopicName)
	})
}
