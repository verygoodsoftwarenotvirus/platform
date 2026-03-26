package sqs

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/testutils"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockMessagePublisher struct {
	mock.Mock
}

func (m *mockMessagePublisher) SendMessage(ctx context.Context, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	retVals := m.Called(ctx, input, optFns)
	return retVals.Get(0).(*sqs.SendMessageOutput), retVals.Error(1)
}

func Test_sqsPublisher_Publish(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)

		a, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*sqsPublisher)
		require.True(t, ok)

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		mmp := &mockMessagePublisher{}
		mmp.On(
			"SendMessage",
			testutils.ContextMatcher,
			mock.MatchedBy(func(*sqs.SendMessageInput) bool { return true }),
			mock.Anything,
		).Return(&sqs.SendMessageOutput{}, nil)

		actual.publisher = mmp

		err = actual.Publish(ctx, inputData)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmp)
	})

	T.Run("with error encoding value", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)

		a, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*sqsPublisher)
		require.True(t, ok)

		inputData := &struct {
			Name json.Number `json:"name"`
		}{
			Name: json.Number(t.Name()),
		}

		err = actual.Publish(ctx, inputData)
		assert.Error(t, err)
	})
}

func TestProvideSQSPublisherProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		actual := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		assert.NotNil(t, actual)
	})
}

func Test_publisherProvider_ProvidePublisher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with cache hit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		actual, err = provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}
