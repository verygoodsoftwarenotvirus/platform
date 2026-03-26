package stripe

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/platform/v3/capitalism"
	"github.com/verygoodsoftwarenotvirus/platform/v3/encoding"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/webhook"
)

const (
	stripeSignatureHeaderKey = "Stripe-Signature"
	implementationName       = "stripe_payment_manager"
)

var _ capitalism.PaymentManager = (*stripePaymentManager)(nil)

type (
	// WebhookSecret is a string alias for dependency injection.
	WebhookSecret string
	// APIKey is a string alias for dependency injection.
	APIKey string

	stripePaymentManager struct {
		logger         logging.Logger
		tracer         tracing.Tracer
		encoderDecoder encoding.ServerEncoderDecoder
		webhookSecret  string
	}
)

// ProvideStripePaymentManager builds a Stripe-backed PaymentManager.
func ProvideStripePaymentManager(logger logging.Logger, tracerProvider tracing.TracerProvider, cfg *Config) capitalism.PaymentManager {
	if cfg == nil {
		return &capitalism.NoopPaymentManager{}
	}

	return &stripePaymentManager{
		webhookSecret:  cfg.WebhookSecret,
		encoderDecoder: encoding.ProvideServerEncoderDecoder(logger, tracerProvider, encoding.ContentTypeJSON),
		logger:         logging.EnsureLogger(logger),
		tracer:         tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(implementationName)),
	}
}

func (s *stripePaymentManager) HandleEventWebhook(req *http.Request) error {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req).WithSpan(span)

	payload, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	signatureHeader := req.Header.Get(stripeSignatureHeaderKey)
	event, err := webhook.ConstructEvent(payload, signatureHeader, s.webhookSecret)
	if err != nil {
		return err
	}

	switch event.Type {
	case stripe.EventTypePaymentIntentSucceeded:
		var paymentIntent stripe.PaymentIntent
		if marshallErr := json.Unmarshal(event.Data.Raw, &paymentIntent); marshallErr != nil {
			return marshallErr
		}
	default:
		logger.WithValue("event_type", event.Type).Info("Unhandled event type")
	}

	return nil
}
