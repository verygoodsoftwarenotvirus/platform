package posthog

import (
	"context"
	"fmt"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v5/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v5/featureflags"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/keys"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	openfeatureposthog "github.com/dhaus67/openfeature-posthog-go"
	"github.com/open-feature/go-sdk/openfeature"
	"github.com/posthog/posthog-go"
)

const (
	serviceName  = "posthog_feature_flag_manager"
	clientDomain = "posthog_feature_flags"
)

var (
	ErrNilConfig          = platformerrors.New("missing config")
	ErrMissingCredentials = platformerrors.New("missing PostHog credentials")
)

type (
	// featureFlagManager implements the feature flag interface using OpenFeature.
	featureFlagManager struct {
		logger         logging.Logger
		tracer         tracing.Tracer
		posthogClient  posthog.Client
		ofClient       *openfeature.Client
		circuitBreaker circuitbreaking.CircuitBreaker
		evalCounter    metrics.Int64Counter
		errorCounter   metrics.Int64Counter
		latencyHist    metrics.Float64Histogram
	}
)

// NewFeatureFlagManager constructs a new featureFlagManager backed by OpenFeature.
func NewFeatureFlagManager(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider, circuitBreaker circuitbreaking.CircuitBreaker, configModifiers ...func(config *posthog.Config)) (featureflags.FeatureFlagManager, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	cfg.CircuitBreakerConfig.EnsureDefaults()

	if cfg.ProjectAPIKey == "" {
		return nil, platformerrors.Wrap(ErrMissingCredentials, "missing credential 'ProjectAPIKey'")
	}

	if cfg.PersonalAPIKey == "" {
		return nil, platformerrors.Wrap(ErrMissingCredentials, "missing credential 'PersonalAPIKey'")
	}

	phc := posthog.Config{
		PersonalApiKey: cfg.PersonalAPIKey,
	}

	for _, modifier := range configModifiers {
		modifier(&phc)
	}

	client, err := posthog.NewWithConfig(
		cfg.ProjectAPIKey,
		phc,
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "failed to create posthog client")
	}

	provider := openfeatureposthog.NewProvider(client)
	if err = openfeature.SetNamedProviderAndWait(clientDomain, provider); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			logger.Error("error closing OpenFeatureFlag client", closeErr)
		}
		return nil, platformerrors.Wrap(err, "failed to set OpenFeature provider")
	}

	ofClient := openfeature.NewClient(clientDomain)

	mp := metrics.EnsureMetricsProvider(metricsProvider)

	evalCounter, err := mp.NewInt64Counter(fmt.Sprintf("%s_evaluations", serviceName))
	if err != nil {
		return nil, platformerrors.Wrap(err, "creating eval counter")
	}

	errorCounter, err := mp.NewInt64Counter(fmt.Sprintf("%s_errors", serviceName))
	if err != nil {
		return nil, platformerrors.Wrap(err, "creating error counter")
	}

	latencyHist, err := mp.NewFloat64Histogram(fmt.Sprintf("%s_latency_ms", serviceName))
	if err != nil {
		return nil, platformerrors.Wrap(err, "creating latency histogram")
	}

	ffm := &featureFlagManager{
		posthogClient:  client,
		ofClient:       ofClient,
		circuitBreaker: circuitBreaker,
		logger:         logging.NewNamedLogger(logger, serviceName),
		tracer:         tracing.NewNamedTracer(tracerProvider, serviceName),
		evalCounter:    evalCounter,
		errorCounter:   errorCounter,
		latencyHist:    latencyHist,
	}

	return ffm, nil
}

// CanUseFeature returns whether a user can use a feature or not.
func (f *featureFlagManager) CanUseFeature(ctx context.Context, userID, feature string) (bool, error) {
	_, span := f.tracer.StartSpan(ctx)
	defer span.End()

	logger := f.logger.WithValue(keys.UserIDKey, userID).WithValue("feature", feature)

	if !f.circuitBreaker.CanProceed() {
		return false, circuitbreaking.ErrCircuitBroken
	}

	startTime := time.Now()
	evalCtx := openfeature.NewEvaluationContext(userID, nil)
	flagEnabled, err := f.ofClient.BooleanValue(ctx, feature, false, evalCtx)
	f.latencyHist.Record(ctx, float64(time.Since(startTime).Milliseconds()))
	if err != nil {
		f.errorCounter.Add(ctx, 1)
		f.circuitBreaker.Failed()
		return false, observability.PrepareAndLogError(err, logger, span, "checking feature flag eligibility")
	}

	f.evalCounter.Add(ctx, 1)
	f.circuitBreaker.Succeeded()
	return flagEnabled, nil
}

// GetStringValue returns the string value of a feature flag for a user.
func (f *featureFlagManager) GetStringValue(ctx context.Context, userID, feature string) (string, error) {
	_, span := f.tracer.StartSpan(ctx)
	defer span.End()

	logger := f.logger.WithValue(keys.UserIDKey, userID).WithValue("feature", feature)

	if !f.circuitBreaker.CanProceed() {
		return "", circuitbreaking.ErrCircuitBroken
	}

	startTime := time.Now()
	evalCtx := openfeature.NewEvaluationContext(userID, nil)
	result, err := f.ofClient.StringValue(ctx, feature, "", evalCtx)
	f.latencyHist.Record(ctx, float64(time.Since(startTime).Milliseconds()))
	if err != nil {
		f.errorCounter.Add(ctx, 1)
		f.circuitBreaker.Failed()
		return "", observability.PrepareAndLogError(err, logger, span, "checking feature flag string variation")
	}

	f.evalCounter.Add(ctx, 1)
	f.circuitBreaker.Succeeded()
	return result, nil
}

// GetInt64Value returns the int64 value of a feature flag for a user.
func (f *featureFlagManager) GetInt64Value(ctx context.Context, userID, feature string) (int64, error) {
	_, span := f.tracer.StartSpan(ctx)
	defer span.End()

	logger := f.logger.WithValue(keys.UserIDKey, userID).WithValue("feature", feature)

	if !f.circuitBreaker.CanProceed() {
		return 0, circuitbreaking.ErrCircuitBroken
	}

	startTime := time.Now()
	evalCtx := openfeature.NewEvaluationContext(userID, nil)
	result, err := f.ofClient.IntValue(ctx, feature, 0, evalCtx)
	f.latencyHist.Record(ctx, float64(time.Since(startTime).Milliseconds()))
	if err != nil {
		f.errorCounter.Add(ctx, 1)
		f.circuitBreaker.Failed()
		return 0, observability.PrepareAndLogError(err, logger, span, "checking feature flag int variation")
	}

	f.evalCounter.Add(ctx, 1)
	f.circuitBreaker.Succeeded()
	return result, nil
}

// Close closes the PostHog client.
func (f *featureFlagManager) Close() error {
	return f.posthogClient.Close()
}
