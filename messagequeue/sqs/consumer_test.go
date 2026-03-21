package sqs

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/testutils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockMessageReceiver struct {
	mock.Mock
}

func (m *mockMessageReceiver) ReceiveMessage(ctx context.Context, input *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	retVals := m.Called(ctx, input, optFns)
	out := retVals.Get(0)
	if out == nil {
		return nil, retVals.Error(1)
	}
	return out.(*sqs.ReceiveMessageOutput), retVals.Error(1)
}

func (m *mockMessageReceiver) DeleteMessage(ctx context.Context, input *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	retVals := m.Called(ctx, input, optFns)
	out := retVals.Get(0)
	if out == nil {
		return nil, retVals.Error(1)
	}
	return out.(*sqs.DeleteMessageOutput), retVals.Error(1)
}

func Test_sqsConsumer_Consume(T *testing.T) {
	T.Parallel()

	queueURL := "https://sqs.us-east-1.amazonaws.com/123456789/test-queue"

	T.Run("successful message handling and deletion", func(t *testing.T) {
		t.Parallel()

		mmr := &mockMessageReceiver{}
		mmr.On(
			"ReceiveMessage",
			testutils.ContextMatcher,
			mock.MatchedBy(func(in *sqs.ReceiveMessageInput) bool {
				return aws.ToString(in.QueueUrl) == queueURL &&
					in.MaxNumberOfMessages == maxNumberOfMessages &&
					in.WaitTimeSeconds == longPollWaitSeconds
			}),
			mock.Anything,
		).Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{
				{
					Body:          aws.String("test-payload"),
					ReceiptHandle: aws.String("receipt-handle-123"),
				},
			},
		}, nil).Once()

		mmr.On(
			"ReceiveMessage",
			testutils.ContextMatcher,
			mock.Anything,
			mock.Anything,
		).Return(&sqs.ReceiveMessageOutput{Messages: []types.Message{}}, nil)

		deleteCalled := make(chan struct{})
		mmr.On(
			"DeleteMessage",
			testutils.ContextMatcher,
			mock.MatchedBy(func(in *sqs.DeleteMessageInput) bool {
				return aws.ToString(in.QueueUrl) == queueURL &&
					aws.ToString(in.ReceiptHandle) == "receipt-handle-123"
			}),
			mock.Anything,
		).Run(func(args mock.Arguments) { deleteCalled <- struct{}{} }).Return(&sqs.DeleteMessageOutput{}, nil).Once()

		handlerDone := make(chan []byte, 1)
		handler := func(ctx context.Context, body []byte) error {
			handlerDone <- body
			return nil
		}

		consumer := provideSQSConsumer(logging.NewNoopLogger(), mmr, queueURL, handler)
		stopChan := make(chan bool, 1)
		errs := make(chan error, 4)

		go consumer.Consume(stopChan, errs)

		receivedBody := <-handlerDone
		<-deleteCalled // wait for DeleteMessage before stopping
		stopChan <- true

		assert.Equal(t, []byte("test-payload"), receivedBody)
		mock.AssertExpectationsForObjects(t, mmr)
	})

	T.Run("handler error does not delete message", func(t *testing.T) {
		t.Parallel()

		anticipatedErr := errors.New("handler failed")
		mmr := &mockMessageReceiver{}
		mmr.On(
			"ReceiveMessage",
			testutils.ContextMatcher,
			mock.Anything,
			mock.Anything,
		).Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{
				{
					Body:          aws.String("fail-payload"),
					ReceiptHandle: aws.String("receipt-handle-456"),
				},
			},
		}, nil).Once()

		mmr.On(
			"ReceiveMessage",
			testutils.ContextMatcher,
			mock.Anything,
			mock.Anything,
		).Return(&sqs.ReceiveMessageOutput{Messages: []types.Message{}}, nil)

		handler := func(ctx context.Context, body []byte) error {
			return anticipatedErr
		}

		consumer := provideSQSConsumer(logging.NewNoopLogger(), mmr, queueURL, handler)
		stopChan := make(chan bool, 1)
		errs := make(chan error, 4)

		go consumer.Consume(stopChan, errs)

		receivedErr := <-errs
		assert.Error(t, receivedErr)
		assert.Equal(t, anticipatedErr, receivedErr)

		stopChan <- true

		mmr.AssertNotCalled(t, "DeleteMessage")
		mock.AssertExpectationsForObjects(t, mmr)
	})
}

func TestProvideSQSConsumerProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := Config{}

		actual := ProvideSQSConsumerProvider(ctx, logger, cfg)
		assert.NotNil(t, actual)
	})
}

func Test_consumerProvider_ProvideConsumer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := Config{}

		provider := ProvideSQSConsumerProvider(ctx, logger, cfg)
		require.NotNil(t, provider)

		actual, err := provider.ProvideConsumer(ctx, "https://sqs.us-east-1.amazonaws.com/123/test", nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("with cache hit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := Config{}
		topic := "https://sqs.us-east-1.amazonaws.com/123/cached-queue"

		provider := ProvideSQSConsumerProvider(ctx, logger, cfg)
		require.NotNil(t, provider)

		actual, err := provider.ProvideConsumer(ctx, topic, nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		actual2, err := provider.ProvideConsumer(ctx, topic, nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual2)
		assert.Same(t, actual, actual2)
	})

	T.Run("with empty topic returns error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := Config{}

		provider := ProvideSQSConsumerProvider(ctx, logger, cfg)
		require.NotNil(t, provider)

		actual, err := provider.ProvideConsumer(ctx, "", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, messagequeue.ErrEmptyTopicName)
	})
}
