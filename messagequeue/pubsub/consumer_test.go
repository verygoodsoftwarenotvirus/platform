package pubsub

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v2/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v2/random"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcpubsub "github.com/testcontainers/testcontainers-go/modules/gcloud/pubsub"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type pubsubTestInfra struct {
	client    *pubsub.Client
	shutdown  func(context.Context) error
	topicName string
}

func buildPubSubTestInfra(t *testing.T, ctx context.Context) *pubsubTestInfra {
	t.Helper()

	randomID, err := random.GenerateHexEncodedString(ctx, 8)
	require.NoError(t, err)
	projectID := "project-" + randomID
	topicID := "topic-" + randomID

	pubsubContainer, err := tcpubsub.Run(
		ctx,
		"google/cloud-sdk:latest",
		tcpubsub.WithProjectID(projectID),
	)
	require.NoError(t, err)
	require.NotNil(t, pubsubContainer)

	conn, err := grpc.NewClient(pubsubContainer.URI(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	require.NotNil(t, conn)

	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
	require.NoError(t, err)
	require.NotNil(t, client)

	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)
	pubSubTopic, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{Name: topicName})
	require.NoError(t, err)
	require.NotNil(t, pubSubTopic)

	subName := subscriptionNameForTopic(pubSubTopic.GetName())
	subscription, err := client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name:  subName,
		Topic: pubSubTopic.GetName(),
	})
	require.NoError(t, err)
	require.NotNil(t, subscription)

	return &pubsubTestInfra{
		client:    client,
		topicName: pubSubTopic.GetName(),
		shutdown:  func(ctx context.Context) error { return pubsubContainer.Terminate(ctx) },
	}
}

func TestSubscriptionNameForTopic(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		result := subscriptionNameForTopic("projects/my-project/topics/my-topic")
		assert.Equal(t, "projects/my-project/subscriptions/my-topic", result)
	})

	T.Run("no match", func(t *testing.T) {
		t.Parallel()

		result := subscriptionNameForTopic("some-other-string")
		assert.Equal(t, "some-other-string", result)
	})
}

func TestBuildPubSubConsumer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		handler := func(_ context.Context, _ []byte) error { return nil }

		consumer := buildPubSubConsumer(logger, tracing.NewNoopTracerProvider(), nil, "test-topic", handler)
		require.NotNil(t, consumer)
	})
}

func TestProvidePubSubConsumerProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		provider := ProvidePubSubConsumerProvider(logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)
	})
}

func TestPubSubConsumerProvider_ProvideConsumer(T *testing.T) {
	T.Parallel()

	T.Run("returns error for empty topic", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		provider := ProvidePubSubConsumerProvider(logger, tracing.NewNoopTracerProvider(), nil)

		consumer, err := provider.ProvideConsumer(t.Context(), "", func(_ context.Context, _ []byte) error { return nil })
		assert.Nil(t, consumer)
		assert.ErrorIs(t, err, messagequeue.ErrEmptyTopicName)
	})

	T.Run("caches consumers for same topic", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		infra := buildPubSubTestInfra(t, ctx)
		t.Cleanup(func() { require.NoError(t, infra.shutdown(context.Background())) })

		logger := logging.NewNoopLogger()
		provider := ProvidePubSubConsumerProvider(logger, tracing.NewNoopTracerProvider(), infra.client)

		handler := func(_ context.Context, _ []byte) error { return nil }

		c1, err := provider.ProvideConsumer(ctx, infra.topicName, handler)
		require.NoError(t, err)
		require.NotNil(t, c1)

		c2, err := provider.ProvideConsumer(ctx, infra.topicName, handler)
		require.NoError(t, err)
		assert.Equal(t, c1, c2)
	})
}

func TestPubSubConsumer_Consume(T *testing.T) {
	T.Parallel()

	T.Run("receives published message", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		infra := buildPubSubTestInfra(t, ctx)
		t.Cleanup(func() { require.NoError(t, infra.shutdown(context.Background())) })

		var called atomic.Bool
		handler := func(_ context.Context, payload []byte) error {
			called.Store(true)
			return nil
		}

		logger := logging.NewNoopLogger()
		provider := ProvidePubSubConsumerProvider(logger, tracing.NewNoopTracerProvider(), infra.client)
		consumer, err := provider.ProvideConsumer(ctx, infra.topicName, handler)
		require.NoError(t, err)

		stopChan := make(chan bool, 1)
		errChan := make(chan error, 1)
		go consumer.Consume(ctx, stopChan, errChan)

		// Publish a message.
		publisher := infra.client.Publisher(infra.topicName)
		result := publisher.Publish(ctx, &pubsub.Message{Data: []byte(`{"name":"test"}`)})
		<-result.Ready()
		_, err = result.Get(ctx)
		require.NoError(t, err)

		// Wait for handler to be called.
		assert.Eventually(t, called.Load, 10*time.Second, 100*time.Millisecond)

		stopChan <- true

		select {
		case err = <-errChan:
			t.Fatalf("unexpected error: %v", err)
		default:
		}
	})

	T.Run("handler error is sent to error channel", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		infra := buildPubSubTestInfra(t, ctx)
		t.Cleanup(func() { require.NoError(t, infra.shutdown(context.Background())) })

		expectedErr := fmt.Errorf("handler failure")
		handler := func(_ context.Context, _ []byte) error {
			return expectedErr
		}

		logger := logging.NewNoopLogger()
		provider := ProvidePubSubConsumerProvider(logger, tracing.NewNoopTracerProvider(), infra.client)
		consumer, err := provider.ProvideConsumer(ctx, infra.topicName, handler)
		require.NoError(t, err)

		stopChan := make(chan bool, 1)
		errChan := make(chan error, 1)
		go consumer.Consume(ctx, stopChan, errChan)

		// Publish a message.
		publisher := infra.client.Publisher(infra.topicName)
		result := publisher.Publish(ctx, &pubsub.Message{Data: []byte(`{"name":"test"}`)})
		<-result.Ready()
		_, err = result.Get(ctx)
		require.NoError(t, err)

		// Wait for the error to appear.
		select {
		case receivedErr := <-errChan:
			assert.Equal(t, expectedErr, receivedErr)
		case <-time.After(10 * time.Second):
			t.Fatal("timed out waiting for handler error")
		}

		stopChan <- true
	})

	T.Run("stops when stop channel is signaled", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		infra := buildPubSubTestInfra(t, ctx)
		t.Cleanup(func() { require.NoError(t, infra.shutdown(context.Background())) })

		handler := func(_ context.Context, _ []byte) error { return nil }

		logger := logging.NewNoopLogger()
		provider := ProvidePubSubConsumerProvider(logger, tracing.NewNoopTracerProvider(), infra.client)
		consumer, err := provider.ProvideConsumer(ctx, infra.topicName, handler)
		require.NoError(t, err)

		stopChan := make(chan bool, 1)
		errChan := make(chan error, 1)

		done := make(chan struct{})
		go func() {
			consumer.Consume(ctx, stopChan, errChan)
			close(done)
		}()

		stopChan <- true

		select {
		case <-done:
			// Consume returned, success.
		case <-time.After(10 * time.Second):
			t.Fatal("timed out waiting for Consume to return after stop signal")
		}
	})

	T.Run("nil stop channel does not panic", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		infra := buildPubSubTestInfra(t, ctx)
		t.Cleanup(func() { require.NoError(t, infra.shutdown(context.Background())) })

		var called atomic.Bool
		handler := func(_ context.Context, _ []byte) error {
			called.Store(true)
			return nil
		}

		logger := logging.NewNoopLogger()
		provider := ProvidePubSubConsumerProvider(logger, tracing.NewNoopTracerProvider(), infra.client)
		consumer, err := provider.ProvideConsumer(ctx, infra.topicName, handler)
		require.NoError(t, err)

		errChan := make(chan error, 1)

		// Pass nil stopChan — should create its own internally.
		done := make(chan struct{})
		go func() {
			consumer.Consume(ctx, nil, errChan)
			close(done)
		}()

		// Publish a message to verify it still works.
		publisher := infra.client.Publisher(infra.topicName)
		result := publisher.Publish(ctx, &pubsub.Message{Data: []byte(`{"name":"test"}`)})
		<-result.Ready()
		_, err = result.Get(ctx)
		require.NoError(t, err)

		assert.Eventually(t, called.Load, 10*time.Second, 100*time.Millisecond)
	})
}
