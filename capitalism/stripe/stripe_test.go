package stripe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	mockencoding "github.com/verygoodsoftwarenotvirus/platform/v5/encoding/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v5/random"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/webhook"
)

type errReader struct{}

func (*errReader) Read([]byte) (int, error) { return 0, fmt.Errorf("read error") }
func (*errReader) Close() error             { return nil }

func TestNewStripePaymentManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		pm := ProvideStripePaymentManager(logger, tracing.NewNoopTracerProvider(), &Config{})

		assert.NotNil(t, pm)
	})

	T.Run("nil config", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		pm := ProvideStripePaymentManager(logger, tracing.NewNoopTracerProvider(), nil)

		assert.NotNil(t, pm)
	})
}

func Test_stripePaymentManager_HandleSubscriptionEventWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := ProvideStripePaymentManager(nil, nil, &Config{}).(*stripePaymentManager)

		paymentIntent := &stripe.PaymentIntent{
			APIResource:      stripe.APIResource{},
			Amount:           0,
			AmountCapturable: 0,
			AmountDetails:    nil,
			AmountReceived:   0,
			Customer:         nil,
			ID:               "",
			Invoice:          nil,
			Metadata:         nil,
			PaymentMethod:    nil,
			ReceiptEmail:     "",
			Status:           "",
		}

		rawMessage, err := json.Marshal(paymentIntent)
		require.NoError(t, err)
		require.NotNil(t, rawMessage)

		exampleInput := &stripe.Event{
			APIResource: stripe.APIResource{},
			Account:     "",
			APIVersion:  "2023-08-16",
			Created:     0,
			Data: &stripe.EventData{
				Object:             nil,
				PreviousAttributes: nil,
				Raw:                json.RawMessage(rawMessage),
			},
			ID:              "",
			Livemode:        false,
			Object:          "",
			PendingWebhooks: 0,
			Request:         nil,
			Type:            stripe.EventTypePaymentIntentSucceeded,
		}
		jsonBytes := pm.encoderDecoder.MustEncode(ctx, exampleInput)

		secret, err := random.GenerateHexEncodedString(ctx, 32)
		require.NoError(t, err)
		require.NotEmpty(t, secret)
		pm.webhookSecret = secret

		now := time.Now()
		signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
			Payload:   jsonBytes,
			Secret:    secret,
			Timestamp: now,
		})

		event, err := webhook.ConstructEvent(signedPayload.Payload, signedPayload.Header, signedPayload.Secret)
		require.NoError(t, err)
		eventPayload := pm.encoderDecoder.MustEncode(ctx, event)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On("DecodeBytes", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		pm.encoderDecoder = encoderDecoder

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://whatever.whocares.gov", bytes.NewReader(eventPayload))
		require.NoError(t, err)
		require.NotNil(t, req)
		req.Header.Set(stripeSignatureHeaderKey, signedPayload.Header)

		err = pm.HandleEventWebhook(req)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error reading body", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := ProvideStripePaymentManager(nil, nil, &Config{}).(*stripePaymentManager)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://whatever.whocares.gov", http.NoBody)
		require.NoError(t, err)
		require.NotNil(t, req)
		req.Body = &errReader{}

		err = pm.HandleEventWebhook(req)
		assert.Error(t, err)
	})

	T.Run("with invalid signature", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := ProvideStripePaymentManager(nil, nil, &Config{}).(*stripePaymentManager)
		pm.webhookSecret = "some_secret"

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://whatever.whocares.gov", bytes.NewReader([]byte(`{}`)))
		require.NoError(t, err)
		require.NotNil(t, req)
		req.Header.Set(stripeSignatureHeaderKey, "invalid_signature")

		err = pm.HandleEventWebhook(req)
		assert.Error(t, err)
	})

	T.Run("with decode error for payment intent", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := ProvideStripePaymentManager(nil, nil, &Config{}).(*stripePaymentManager)

		paymentIntent := &stripe.PaymentIntent{}

		rawMessage, err := json.Marshal(paymentIntent)
		require.NoError(t, err)

		exampleInput := &stripe.Event{
			APIVersion: "2023-08-16",
			Data: &stripe.EventData{
				Raw: json.RawMessage(rawMessage),
			},
			Type: stripe.EventTypePaymentIntentSucceeded,
		}
		jsonBytes := pm.encoderDecoder.MustEncode(ctx, exampleInput)

		secret, err := random.GenerateHexEncodedString(ctx, 32)
		require.NoError(t, err)
		require.NotEmpty(t, secret)
		pm.webhookSecret = secret

		signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
			Payload:   jsonBytes,
			Secret:    secret,
			Timestamp: time.Now(),
		})

		event, err := webhook.ConstructEvent(signedPayload.Payload, signedPayload.Header, signedPayload.Secret)
		require.NoError(t, err)
		eventPayload := pm.encoderDecoder.MustEncode(ctx, event)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On("DecodeBytes", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("decode error"))
		pm.encoderDecoder = encoderDecoder

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://whatever.whocares.gov", bytes.NewReader(eventPayload))
		require.NoError(t, err)
		require.NotNil(t, req)
		req.Header.Set(stripeSignatureHeaderKey, signedPayload.Header)

		err = pm.HandleEventWebhook(req)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with unhandled event type", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := ProvideStripePaymentManager(nil, nil, &Config{}).(*stripePaymentManager)

		exampleInput := &stripe.Event{
			APIVersion: "2023-08-16",
			Data: &stripe.EventData{
				Raw: json.RawMessage(`{}`),
			},
			Type: stripe.EventTypeAccountUpdated,
		}
		jsonBytes := pm.encoderDecoder.MustEncode(ctx, exampleInput)

		secret, err := random.GenerateHexEncodedString(ctx, 32)
		require.NoError(t, err)
		require.NotEmpty(t, secret)
		pm.webhookSecret = secret

		signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
			Payload:   jsonBytes,
			Secret:    secret,
			Timestamp: time.Now(),
		})

		event, err := webhook.ConstructEvent(signedPayload.Payload, signedPayload.Header, signedPayload.Secret)
		require.NoError(t, err)
		eventPayload := pm.encoderDecoder.MustEncode(ctx, event)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://whatever.whocares.gov", bytes.NewReader(eventPayload))
		require.NoError(t, err)
		require.NotNil(t, req)
		req.Header.Set(stripeSignatureHeaderKey, signedPayload.Header)

		err = pm.HandleEventWebhook(req)
		assert.NoError(t, err)
	})
}
