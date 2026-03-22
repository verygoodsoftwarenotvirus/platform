package apns

import (
	"context"
	"regexp"

	"github.com/verygoodsoftwarenotvirus/platform/v2/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

// apnsDeviceTokenHexPattern validates a 64-character hex string (32-byte token).
var apnsDeviceTokenHexPattern = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

const (
	o11yName = "ios_notif_sender"
)

// Config holds APNs configuration.
type Config struct {
	AuthKeyPath string
	KeyID       string
	TeamID      string
	BundleID    string
	Production  bool
}

// Sender sends push notifications to iOS devices via APNs.
type Sender struct {
	tracer tracing.Tracer
	logger logging.Logger
	client *apns2.Client
	topic  string
}

// NewSender creates an APNs sender from config.
func NewSender(cfg *Config, tracerProvider tracing.TracerProvider, logger logging.Logger) (*Sender, error) {
	if cfg == nil || cfg.AuthKeyPath == "" || cfg.KeyID == "" || cfg.TeamID == "" || cfg.BundleID == "" {
		return nil, errors.New("apns: missing required config (authKeyPath, keyID, teamID, bundleID)")
	}

	authKey, err := token.AuthKeyFromFile(cfg.AuthKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "apns: loading auth key")
	}

	t := &token.Token{
		AuthKey: authKey,
		KeyID:   cfg.KeyID,
		TeamID:  cfg.TeamID,
	}
	if _, err = t.Generate(); err != nil {
		return nil, errors.Wrap(err, "apns: generating token")
	}

	client := apns2.NewTokenClient(t)
	if cfg.Production {
		client = client.Production()
	} else {
		client = client.Development()
	}

	return &Sender{
		client: client,
		topic:  cfg.BundleID,
		tracer: tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		logger: logging.EnsureLogger(logger).WithName(o11yName),
	}, nil
}

// Send sends a push notification to a single device token.
// The device token must be a 64-character hex string (APNs format).
// badgeCount is optional; when non-nil, sets aps.badge on the app icon.
func (s *Sender) Send(ctx context.Context, deviceToken, title, body string, badgeCount *int) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if !apnsDeviceTokenHexPattern.MatchString(deviceToken) {
		return errors.Newf("apns: invalid device token format (expected 64 hex chars, got len %d)", len(deviceToken))
	}

	logger := s.logger.WithValue("title", title)

	p := payload.NewPayload().
		AlertTitle(title).
		AlertBody(body)
	if badgeCount != nil {
		p = p.Badge(*badgeCount)
	}

	n := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       s.topic,
		Payload:     p,
		Priority:    apns2.PriorityHigh,
	}

	res, err := s.client.PushWithContext(ctx, n)
	if err != nil {
		return errors.Wrap(err, "apns: push failed")
	}

	if !res.Sent() {
		err = errors.Newf("apns: %s (status %d)", res.Reason, res.StatusCode)
		logger = logger.WithValue("statusCode", res.StatusCode).
			WithValue("reason", res.Reason).
			WithValue("apnsID", res.ApnsID)
		return observability.PrepareAndLogError(err, logger, span, "sending apns notification")
	}

	return nil
}
