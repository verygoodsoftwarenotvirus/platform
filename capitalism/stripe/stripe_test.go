package stripe

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/random"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/webhook"
)

func TestNewStripePaymentManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		pm := ProvideStripePaymentManager(logger, tracing.NewNoopTracerProvider(), &Config{})

		assert.NotNil(t, pm)
	})
}

func Test_stripePaymentManager_HandleSubscriptionEventWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := ProvideStripePaymentManager(nil, nil, &Config{}).(*stripePaymentManager)
		pm.webhookSecret = "example_webhook_secret"

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

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://whatever.whocares.gov", bytes.NewReader(eventPayload))
		require.NoError(t, err)
		require.NotNil(t, req)
		req.Header.Set(stripeSignatureHeaderKey, signedPayload.Header)

		err = pm.HandleEventWebhook(req)
		assert.NoError(t, err)
	})
}
