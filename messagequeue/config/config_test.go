package msgconfig

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/messagequeue/sqs"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func Test_cleanString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, cleanString(t.Name()))
	})
}

func TestProvideConsumerProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := &Config{
			Consumer: MessageQueueConfig{
				Provider: ProviderRedis,
			},
		}

		p, err := ProvideConsumerProvider(ctx, logger, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	T.Run("with SQS provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := &Config{
			Consumer: MessageQueueConfig{
				Provider: ProviderSQS,
				SQS:      sqs.Config{},
			},
		}

		p, err := ProvideConsumerProvider(ctx, logger, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := &Config{}

		p, err := ProvideConsumerProvider(ctx, logger, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})
}

func TestProvidePublisherProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := &Config{
			Publisher: MessageQueueConfig{
				Provider: ProviderRedis,
			},
		}

		p, err := ProvidePublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), cfg)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := &Config{}

		p, err := ProvidePublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), cfg)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})
}
