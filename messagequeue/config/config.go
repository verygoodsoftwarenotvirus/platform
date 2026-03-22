package msgconfig

import (
	"context"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/v2/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v2/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v2/messagequeue/pubsub"
	"github.com/verygoodsoftwarenotvirus/platform/v2/messagequeue/redis"
	"github.com/verygoodsoftwarenotvirus/platform/v2/messagequeue/sqs"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	ps "cloud.google.com/go/pubsub/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ProviderRedis is used to refer to redis.
	ProviderRedis provider = "redis"
	// ProviderSQS is used to refer to sqs.
	ProviderSQS provider = "sqs"
	// ProviderPubSub is used to refer to GCP Pub/Sub.
	ProviderPubSub provider = "pubsub"
)

var (
	ErrNilConfig = errors.New("nil config provided")
)

type (
	// provider is used to indicate what messaging provider we'll use.
	provider string

	// MessageQueueConfig is used to indicate how the messaging provider should be configured.
	MessageQueueConfig struct {
		_ struct{} `json:"-"`

		Provider provider      `env:"PROVIDER"      json:"provider,omitempty"`
		SQS      sqs.Config    `envPrefix:"SQS_"    json:"sqs"`
		PubSub   pubsub.Config `envPrefix:"PUBSUB_" json:"pubSub"`
		Redis    redis.Config  `envPrefix:"REDIS_"  json:"redis"`
	}

	// Config is used to indicate how the messaging provider should be configured.
	Config struct {
		_ struct{} `json:"-"`

		Consumer  MessageQueueConfig `envPrefix:"CONSUMER_"  json:"consumers"`
		Publisher MessageQueueConfig `envPrefix:"PUBLISHER_" json:"publishers"`
	}

	// QueuesConfig contains the various queue names.
	QueuesConfig struct {
		_ struct{} `json:"-"`

		DataChangesTopicName              string `env:"DATA_CHANGES_TOPIC_NAME"               json:"dataChangesTopicName"`
		OutboundEmailsTopicName           string `env:"OUTBOUND_EMAILS_TOPIC_NAME"            json:"outboundEmailsTopicName"`
		SearchIndexRequestsTopicName      string `env:"SEARCH_INDEX_REQUESTS_TOPIC_NAME"      json:"searchIndexRequestsTopicName"`
		MobileNotificationsTopicName      string `env:"MOBILE_NOTIFICATIONS_TOPIC_NAME"       json:"mobileNotificationsTopicName"`
		UserDataAggregationTopicName      string `env:"USER_DATA_AGGREGATION_TOPIC_NAME"      json:"userDataAggregationTopicName"`
		WebhookExecutionRequestsTopicName string `env:"WEBHOOK_EXECUTION_REQUESTS_TOPIC_NAME" json:"webhookExecutionRequestsTopicName"`
	}
)

var _ validation.ValidatableWithContext = (*QueuesConfig)(nil)

// ValidateWithContext validates a QueuesConfig struct.
func (c *QueuesConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.DataChangesTopicName, validation.Required),
		validation.Field(&c.OutboundEmailsTopicName, validation.Required),
		validation.Field(&c.SearchIndexRequestsTopicName, validation.Required),
		validation.Field(&c.MobileNotificationsTopicName, validation.Required),
		validation.Field(&c.UserDataAggregationTopicName, validation.Required),
		validation.Field(&c.WebhookExecutionRequestsTopicName, validation.Required),
	)
}

func cleanString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ProvideConsumerProvider provides a ConsumerProvider.
func ProvideConsumerProvider(ctx context.Context, logger logging.Logger, tracerProvider tracing.TracerProvider, c *Config) (messagequeue.ConsumerProvider, error) {
	if c == nil {
		return nil, ErrNilConfig
	}

	switch cleanString(string(c.Consumer.Provider)) {
	case string(ProviderRedis):
		return redis.ProvideRedisConsumerProvider(logger, tracerProvider, c.Consumer.Redis), nil
	case string(ProviderSQS):
		return sqs.ProvideSQSConsumerProvider(ctx, logger, tracerProvider, c.Consumer.SQS), nil
	case string(ProviderPubSub):
		client, err := ps.NewClientWithConfig(ctx, c.Consumer.PubSub.ProjectID, &ps.ClientConfig{
			EnableOpenTelemetryTracing: true,
		})
		if err != nil {
			return nil, errors.Wrap(err, "establishing PubSub client")
		}

		return pubsub.ProvidePubSubConsumerProvider(logger, tracerProvider, client), nil
	default:
		logger.Info("Using noop consumer provider")
		return &messagequeue.NoopConsumerProvider{}, nil
	}
}

// ProvidePublisherProvider provides a PublisherProvider.
func ProvidePublisherProvider(ctx context.Context, logger logging.Logger, tracerProvider tracing.TracerProvider, c *Config) (messagequeue.PublisherProvider, error) {
	if c == nil {
		return nil, ErrNilConfig
	}

	switch cleanString(string(c.Publisher.Provider)) {
	case string(ProviderRedis):
		return redis.ProvideRedisPublisherProvider(logger, tracerProvider, c.Publisher.Redis), nil
	case string(ProviderSQS):
		return sqs.ProvideSQSPublisherProvider(ctx, logger, tracerProvider), nil
	case string(ProviderPubSub):
		client, err := ps.NewClientWithConfig(ctx, c.Publisher.PubSub.ProjectID, &ps.ClientConfig{
			EnableOpenTelemetryTracing: true,
		})
		if err != nil {
			return nil, errors.Wrap(err, "establishing PubSub client")
		}

		return pubsub.ProvidePubSubPublisherProvider(logger, tracerProvider, client, c.Publisher.PubSub.ProjectID), nil
	default:
		logger.Info("Using noop publisher provider")
		return &messagequeue.NoopPublisherProvider{}, nil
	}
}
