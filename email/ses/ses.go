package ses

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v5/email"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v5/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

const (
	name = "ses_emailer"
)

var (
	_ email.Emailer = (*Emailer)(nil)

	// ErrNilConfig indicates a nil config was provided.
	ErrNilConfig = platformerrors.New("ses config is nil")
	// ErrEmptyRegion indicates an empty AWS region was provided.
	ErrEmptyRegion = platformerrors.New("empty AWS region")
	// ErrNilHTTPClient indicates a nil HTTP client was provided.
	ErrNilHTTPClient = platformerrors.New("nil HTTP client")
)

// Emailer uses AWS SES v2 to send email.
type Emailer struct {
	logger         logging.Logger
	tracer         tracing.Tracer
	sendCounter    metrics.Int64Counter
	errorCounter   metrics.Int64Counter
	latencyHist    metrics.Float64Histogram
	client         *sesv2.Client
	circuitBreaker circuitbreaking.CircuitBreaker
}

// NewSESEmailer returns a new AWS SES-backed Emailer.
func NewSESEmailer(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, client *http.Client, circuitBreaker circuitbreaking.CircuitBreaker, metricsProvider metrics.Provider) (*Emailer, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	if cfg.Region == "" {
		return nil, ErrEmptyRegion
	}

	if client == nil {
		return nil, ErrNilHTTPClient
	}

	mp := metrics.EnsureMetricsProvider(metricsProvider)

	sendCounter, err := mp.NewInt64Counter(fmt.Sprintf("%s_sends", name))
	if err != nil {
		return nil, platformerrors.Wrap(err, "creating send counter")
	}

	errorCounter, err := mp.NewInt64Counter(fmt.Sprintf("%s_errors", name))
	if err != nil {
		return nil, platformerrors.Wrap(err, "creating error counter")
	}

	latencyHist, err := mp.NewFloat64Histogram(fmt.Sprintf("%s_latency_ms", name))
	if err != nil {
		return nil, platformerrors.Wrap(err, "creating latency histogram")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithHTTPClient(client),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "loading AWS config")
	}

	return &Emailer{
		logger:         logging.NewNamedLogger(logger, name),
		tracer:         tracing.NewNamedTracer(tracerProvider, name),
		sendCounter:    sendCounter,
		errorCounter:   errorCounter,
		latencyHist:    latencyHist,
		client:         sesv2.NewFromConfig(awsCfg),
		circuitBreaker: circuitBreaker,
	}, nil
}

// SendEmail sends an email via AWS SES v2.
func (e *Emailer) SendEmail(ctx context.Context, details *email.OutboundEmailMessage) error {
	ctx, span := e.tracer.StartSpan(ctx)
	defer span.End()

	startTime := time.Now()
	defer func() {
		e.latencyHist.Record(ctx, float64(time.Since(startTime).Milliseconds()))
	}()

	logger := e.logger.WithValue("email.subject", details.Subject).WithValue("email.to_address", details.ToAddress)
	tracing.AttachToSpan(span, "to_email", details.ToAddress)

	if e.circuitBreaker.CannotProceed() {
		return circuitbreaking.ErrCircuitBroken
	}

	from := details.FromAddress
	if details.FromName != "" {
		from = fmt.Sprintf("%s <%s>", details.FromName, details.FromAddress)
	}

	to := details.ToAddress
	if details.ToName != "" {
		to = fmt.Sprintf("%s <%s>", details.ToName, details.ToAddress)
	}

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(from),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: aws.String(details.Subject),
				},
				Body: &types.Body{
					Html: &types.Content{
						Data: aws.String(details.HTMLContent),
					},
				},
			},
		},
	}

	if _, err := e.client.SendEmail(ctx, input); err != nil {
		e.circuitBreaker.Failed()
		e.errorCounter.Add(ctx, 1)
		return observability.PrepareAndLogError(err, logger, span, "sending email")
	}

	e.circuitBreaker.Succeeded()
	e.sendCounter.Add(ctx, 1)
	return nil
}
