package posthog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking"
	mockCircuitBreaker "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/mock"
	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/featureflags"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	openfeatureposthog "github.com/dhaus67/openfeature-posthog-go"
	"github.com/open-feature/go-sdk/openfeature"
	"github.com/posthog/posthog-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testFlags = map[string]any{
	"bool-flag":   true,
	"string-flag": "hello-world",
	"int-flag":    "42",
	"float-flag":  "3.14",
	"object-flag": `{"key":"value"}`,
}

func evalCtx(targetingKey string) featureflags.EvaluationContext {
	return featureflags.EvaluationContext{TargetingKey: targetingKey}
}

func posthogFlagsHandler(flags map[string]any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/feature_flag/local_evaluation"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"flags":              []any{},
				"group_type_mapping": map[string]string{},
			})
		case strings.HasPrefix(r.URL.Path, "/flags/"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"featureFlags":        flags,
				"featureFlagPayloads": map[string]any{},
			})
		default:
			w.WriteHeader(http.StatusOK)
		}
	})
}

func posthogErrorHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/feature_flag/local_evaluation"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"flags":              []any{},
				"group_type_mapping": map[string]string{},
			})
		case strings.HasPrefix(r.URL.Path, "/flags/"):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusOK)
		}
	})
}

func buildTestManager(t *testing.T, cb circuitbreaking.CircuitBreaker, configModifiers ...func(config *posthog.Config)) *featureFlagManager {
	t.Helper()

	cfg := &Config{
		ProjectAPIKey:  t.Name(),
		PersonalAPIKey: t.Name(),
	}

	ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, cb, configModifiers...)
	require.NoError(t, err)
	require.NotNil(t, ffm)

	return ffm.(*featureFlagManager)
}

func buildTestManagerWithHandler(t *testing.T, handler http.Handler) *featureFlagManager {
	t.Helper()

	ts := httptest.NewServer(handler)

	phConfig := posthog.Config{
		PersonalApiKey: t.Name(),
		Endpoint:       ts.URL,
	}

	client, err := posthog.NewWithConfig(t.Name(), phConfig)
	require.NoError(t, err)

	t.Cleanup(func() {
		client.Close()
		ts.Close()
	})

	// Use a unique domain per test to avoid global OpenFeature provider conflicts.
	domain := "test_" + strings.ReplaceAll(t.Name(), "/", "_")
	provider := openfeatureposthog.NewProvider(client)
	err = openfeature.SetNamedProviderAndWait(domain, provider)
	require.NoError(t, err)

	ofClient := openfeature.NewClient(domain)

	mp := metrics.EnsureMetricsProvider(nil)
	evalCounter, err := mp.NewInt64Counter(fmt.Sprintf("%s_evaluations", serviceName))
	require.NoError(t, err)
	errorCounter, err := mp.NewInt64Counter(fmt.Sprintf("%s_errors", serviceName))
	require.NoError(t, err)
	latencyHist, err := mp.NewFloat64Histogram(fmt.Sprintf("%s_latency_ms", serviceName))
	require.NoError(t, err)

	return &featureFlagManager{
		posthogClient:  client,
		ofClient:       ofClient,
		circuitBreaker: cbnoop.NewCircuitBreaker(),
		logger:         logging.NewNoopLogger(),
		tracer:         tracing.NewNamedTracer(tracing.NewNoopTracerProvider(), serviceName),
		evalCounter:    evalCounter,
		errorCounter:   errorCounter,
		latencyHist:    latencyHist,
	}
}

func TestNewFeatureFlagManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			ProjectAPIKey:  t.Name(),
			PersonalAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, cbnoop.NewCircuitBreaker())
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		actual, err := NewFeatureFlagManager(nil, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, cbnoop.NewCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with missing project API key", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, cbnoop.NewCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with missing personal API key", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			ProjectAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, cbnoop.NewCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			ProjectAPIKey:  t.Name(),
			PersonalAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, cbnoop.NewCircuitBreaker(), func(config *posthog.Config) {
			config.Interval = -1
		})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestToOpenFeatureContext(T *testing.T) {
	T.Parallel()

	T.Run("with attributes", func(t *testing.T) {
		t.Parallel()

		ec := featureflags.EvaluationContext{
			TargetingKey: "user123",
			Attributes:   map[string]any{"plan": "pro", "region": "us-east"},
		}

		result := toOpenFeatureContext(ec)

		assert.Equal(t, "user123", result.TargetingKey())
		assert.Equal(t, "pro", result.Attribute("plan"))
		assert.Equal(t, "us-east", result.Attribute("region"))
	})

	T.Run("with nil attributes", func(t *testing.T) {
		t.Parallel()

		ec := featureflags.EvaluationContext{
			TargetingKey: "user456",
		}

		result := toOpenFeatureContext(ec)

		assert.Equal(t, "user456", result.TargetingKey())
	})
}

func TestFeatureFlagManager_CanUseFeature(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		ffm := buildTestManagerWithHandler(t, posthogFlagsHandler(testFlags))

		actual, err := ffm.CanUseFeature(ctx, "bool-flag", evalCtx("user123"))
		assert.NoError(t, err)
		assert.True(t, actual)
	})

	T.Run("with error executing request", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		ffm := buildTestManagerWithHandler(t, posthogErrorHandler())

		actual, err := ffm.CanUseFeature(ctx, "bool-flag", evalCtx("user123"))
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with broken circuit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.CircuitBreakerMock{
			CanProceedFunc: func() bool { return false },
		}

		ffm := buildTestManager(t, cb)

		result, err := ffm.CanUseFeature(ctx, "some-flag", evalCtx("user123"))
		assert.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)
		assert.False(t, result)
	})
}

func TestFeatureFlagManager_GetStringValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		ffm := buildTestManagerWithHandler(t, posthogFlagsHandler(testFlags))

		result, err := ffm.GetStringValue(ctx, "string-flag", "fallback", evalCtx("user123"))
		assert.NoError(t, err)
		assert.Equal(t, "hello-world", result)
	})

	T.Run("with error executing request", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		ffm := buildTestManagerWithHandler(t, posthogErrorHandler())

		result, err := ffm.GetStringValue(ctx, "string-flag", "fallback", evalCtx("user123"))
		assert.Error(t, err)
		assert.Equal(t, "fallback", result)
	})

	T.Run("with broken circuit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.CircuitBreakerMock{
			CanProceedFunc: func() bool { return false },
		}

		ffm := buildTestManager(t, cb)

		result, err := ffm.GetStringValue(ctx, "some-flag", "fallback", evalCtx("user123"))
		assert.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)
		assert.Equal(t, "fallback", result)
	})
}

func TestFeatureFlagManager_GetInt64Value(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		ffm := buildTestManagerWithHandler(t, posthogFlagsHandler(testFlags))

		result, err := ffm.GetInt64Value(ctx, "int-flag", int64(0), evalCtx("user123"))
		assert.NoError(t, err)
		assert.Equal(t, int64(42), result)
	})

	T.Run("with error executing request", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		ffm := buildTestManagerWithHandler(t, posthogErrorHandler())

		result, err := ffm.GetInt64Value(ctx, "int-flag", int64(42), evalCtx("user123"))
		assert.Error(t, err)
		assert.Equal(t, int64(42), result)
	})

	T.Run("with broken circuit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.CircuitBreakerMock{
			CanProceedFunc: func() bool { return false },
		}

		ffm := buildTestManager(t, cb)

		result, err := ffm.GetInt64Value(ctx, "some-flag", int64(42), evalCtx("user123"))
		assert.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)
		assert.Equal(t, int64(42), result)
	})
}

func TestFeatureFlagManager_GetFloat64Value(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		ffm := buildTestManagerWithHandler(t, posthogFlagsHandler(testFlags))

		result, err := ffm.GetFloat64Value(ctx, "float-flag", 0.0, evalCtx("user123"))
		assert.NoError(t, err)
		assert.InDelta(t, 3.14, result, 1e-9)
	})

	T.Run("with error executing request", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		ffm := buildTestManagerWithHandler(t, posthogErrorHandler())

		result, err := ffm.GetFloat64Value(ctx, "float-flag", 3.14, evalCtx("user123"))
		assert.Error(t, err)
		assert.InDelta(t, 3.14, result, 1e-9)
	})

	T.Run("with broken circuit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.CircuitBreakerMock{
			CanProceedFunc: func() bool { return false },
		}

		ffm := buildTestManager(t, cb)

		result, err := ffm.GetFloat64Value(ctx, "some-flag", 3.14, evalCtx("user123"))
		assert.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)
		assert.InDelta(t, 3.14, result, 1e-9)
	})
}

func TestFeatureFlagManager_GetObjectValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		ffm := buildTestManagerWithHandler(t, posthogFlagsHandler(testFlags))

		def := map[string]any{"default": true}
		result, err := ffm.GetObjectValue(ctx, "object-flag", def, evalCtx("user123"))
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{"key": "value"}, result)
	})

	T.Run("with error executing request", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		ffm := buildTestManagerWithHandler(t, posthogErrorHandler())

		def := map[string]any{"k": "v"}
		result, err := ffm.GetObjectValue(ctx, "object-flag", def, evalCtx("user123"))
		assert.Error(t, err)
		assert.Equal(t, def, result)
	})

	T.Run("with broken circuit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.CircuitBreakerMock{
			CanProceedFunc: func() bool { return false },
		}

		ffm := buildTestManager(t, cb)

		def := map[string]any{"k": "v"}
		result, err := ffm.GetObjectValue(ctx, "some-flag", def, evalCtx("user123"))
		assert.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)
		assert.Equal(t, def, result)
	})
}

func TestFeatureFlagManager_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ffm := buildTestManager(t, cbnoop.NewCircuitBreaker())

		err := ffm.Close()
		assert.NoError(t, err)
	})
}
