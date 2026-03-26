package fcm

import (
	"context"
	"os"

	"github.com/verygoodsoftwarenotvirus/platform/v3/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

const (
	o11yName = "android_notif_sender"
)

// Config holds FCM configuration.
type Config struct {
	// CredentialsPath is the path to the Firebase service account JSON file.
	// If empty, Application Default Credentials (ADC) are used.
	CredentialsPath string
}

// Sender sends push notifications to Android devices via FCM.
type Sender struct {
	client *messaging.Client
	tracer tracing.Tracer
	logger logging.Logger
}

// NewSender creates an FCM sender from config.
func NewSender(ctx context.Context, cfg *Config, tracerProvider tracing.TracerProvider, logger logging.Logger) (*Sender, error) {
	if cfg == nil {
		return nil, errors.New("fcm: config is required")
	}

	var opts []option.ClientOption
	if cfg.CredentialsPath != "" {
		creds, err := os.ReadFile(cfg.CredentialsPath)
		if err != nil {
			return nil, errors.Wrap(err, "fcm: credentials file not found")
		}
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, creds))
	}
	// If CredentialsPath is empty, Application Default Credentials (ADC) are used.

	app, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "fcm: initializing app")
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fcm: creating messaging client")
	}

	return &Sender{
		client: client,
		logger: logging.EnsureLogger(logger).WithName(o11yName),
		tracer: tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
	}, nil
}

// Send sends a push notification to a single device token.
func (s *Sender) Send(ctx context.Context, deviceToken, title, body string) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("title", title)

	msg := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	if _, err := s.client.Send(ctx, msg); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "sending fcm message")
	}

	return nil
}
