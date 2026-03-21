package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/random"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcpubsub "github.com/testcontainers/testcontainers-go/modules/gcloud/pubsub"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func buildPubSubBackedConsumer(t *testing.T, ctx context.Context, topic string, handlerFunc func(context.Context, []byte) error) (publisher messagequeue.Consumer, shutdownFunc func(context.Context) error) {
	t.Helper()

	projectID, err := random.GenerateHexEncodedString(ctx, 8)
	require.NoError(t, err)

	pubSubContainer, err := tcpubsub.Run(
		ctx,
		"google/cloud-sdk:latest",
		tcpubsub.WithProjectID(projectID),
	)
	if err != nil {
		panic(err)
	}

	conn, err := grpc.NewClient(pubSubContainer.URI(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	require.NotNil(t, conn)

	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
	require.NoError(t, err)
	require.NotNil(t, client)

	pubSubTopic, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{Name: topic})
	require.NoError(t, err)
	require.NotNil(t, pubSubTopic)

	subscription, err := client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name:  subscriptionNameForTopic(pubSubTopic.GetName()),
		Topic: pubSubTopic.GetName(),
	})
	require.NoError(t, err)
	require.NotNil(t, subscription)

	logger := logging.NewNoopLogger()
	provider := ProvidePubSubConsumerProvider(logger, client)
	require.NotNil(t, provider)

	publisher, err = provider.ProvideConsumer(ctx, pubSubTopic.GetName(), handlerFunc)
	assert.NotNil(t, publisher)
	assert.NoError(t, err)

	return publisher, func(ctx context.Context) error { return pubSubContainer.Terminate(ctx) }
}

func Test_pubSubConsumer_Consume(T *testing.T) {
	// TODO: get this working
	T.SkipNow()
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		publisher, publisherShutdownFunc := buildPubSubBackedPublisher(t, ctx)
		defer func() {
			require.NoError(t, publisherShutdownFunc(ctx))
		}()

		handlerFuncCalled := false
		consumer, consumerShutdownFunc := buildPubSubBackedConsumer(t, ctx, publisher.(*pubSubPublisher).topic, func(_ context.Context, payload []byte) error {
			handlerFuncCalled = true
			return nil
		})
		defer func() {
			require.NoError(t, consumerShutdownFunc(ctx))
		}()

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		errors := make(chan error, 1)
		go consumer.Consume(nil, errors)
		assert.NoError(t, publisher.Publish(ctx, inputData))

		<-time.After(5 * time.Second)

		assert.Empty(t, errors)
		assert.True(t, handlerFuncCalled)
	})
}
