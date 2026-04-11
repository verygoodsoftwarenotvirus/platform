package launchdarkly

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking"
	mockCircuitBreaker "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/mock"
	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/featureflags"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	"github.com/launchdarkly/go-server-sdk-evaluation/v2/ldmodel"
	ld "github.com/launchdarkly/go-server-sdk/v6"
	"github.com/launchdarkly/go-server-sdk/v6/ldcomponents"
	"github.com/launchdarkly/go-server-sdk/v6/subsystems"
	"github.com/launchdarkly/go-server-sdk/v6/subsystems/ldstoreimpl"
	"github.com/launchdarkly/go-server-sdk/v6/subsystems/ldstoretypes"
	ofld "github.com/open-feature/go-sdk-contrib/providers/launchdarkly/pkg"
	"github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func evalCtx(targetingKey string) featureflags.EvaluationContext {
	return featureflags.EvaluationContext{TargetingKey: targetingKey}
}

// fakeLaunchDarklyDataSource provides no flag data.
type fakeLaunchDarklyDataSource struct{}

func (f *fakeLaunchDarklyDataSource) Close() error             { return nil }
func (f *fakeLaunchDarklyDataSource) IsInitialized() bool      { return true }
func (f *fakeLaunchDarklyDataSource) Start(ch chan<- struct{}) { close(ch) }

type fakeLaunchDarklyDataSourceBuilder struct{}

func (b *fakeLaunchDarklyDataSourceBuilder) Build(subsystems.ClientContext) (subsystems.DataSource, error) {
	return &fakeLaunchDarklyDataSource{}, nil
}

// testDataSource pushes pre-configured flag data into the SDK on start.
type testDataSource struct {
	sink  subsystems.DataSourceUpdateSink
	flags []ldstoretypes.KeyedItemDescriptor
}

func (ds *testDataSource) Close() error        { return nil }
func (ds *testDataSource) IsInitialized() bool { return true }
func (ds *testDataSource) Start(ch chan<- struct{}) {
	ds.sink.Init([]ldstoretypes.Collection{
		{Kind: ldstoreimpl.Features(), Items: ds.flags},
		{Kind: ldstoreimpl.Segments(), Items: nil},
	})
	close(ch)
}

type testDataSourceBuilder struct {
	flags []ldstoretypes.KeyedItemDescriptor
}

func (b *testDataSourceBuilder) Build(ctx subsystems.ClientContext) (subsystems.DataSource, error) {
	return &testDataSource{
		sink:  ctx.GetDataSourceUpdateSink(),
		flags: b.flags,
	}, nil
}

func flagItem(key string, offValue, onValue *ldvalue.Value) ldstoretypes.KeyedItemDescriptor {
	flag := &ldmodel.FeatureFlag{
		Key:         key,
		On:          true,
		Variations:  []ldvalue.Value{*offValue, *onValue},
		Fallthrough: ldmodel.VariationOrRollout{Variation: ldvalue.NewOptionalInt(1)},
		Version:     1,
	}
	ldmodel.PreprocessFlag(flag)
	return ldstoretypes.KeyedItemDescriptor{
		Key:  key,
		Item: ldstoretypes.ItemDescriptor{Version: 1, Item: flag},
	}
}

func testFlagItems() []ldstoretypes.KeyedItemDescriptor {
	boolOff, boolOn := ldvalue.Bool(false), ldvalue.Bool(true)
	stringOff, stringOn := ldvalue.String("fallback"), ldvalue.String("hello-world")
	intOff, intOn := ldvalue.Int(0), ldvalue.Int(42)
	floatOff, floatOn := ldvalue.Float64(0.0), ldvalue.Float64(3.14)
	objectOff, objectOn := ldvalue.Null(), ldvalue.ObjectBuild().Set("key", ldvalue.String("value")).Build()

	return []ldstoretypes.KeyedItemDescriptor{
		flagItem("bool-flag", &boolOff, &boolOn),
		flagItem("string-flag", &stringOff, &stringOn),
		flagItem("int-flag", &intOff, &intOn),
		flagItem("float-flag", &floatOff, &floatOn),
		flagItem("object-flag", &objectOff, &objectOn),
	}
}

func buildTestManager(t *testing.T, cb circuitbreaking.CircuitBreaker) *featureFlagManager {
	t.Helper()

	cfg := &Config{SDKKey: t.Name()}

	ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, http.DefaultClient, cb, func(config ld.Config) ld.Config {
		config.DataSource = &fakeLaunchDarklyDataSourceBuilder{}
		return config
	})
	require.NoError(t, err)
	require.NotNil(t, ffm)

	return ffm.(*featureFlagManager)
}

func buildTestManagerWithFlags(t *testing.T, flags []ldstoretypes.KeyedItemDescriptor) *featureFlagManager {
	t.Helper()

	ldConfig := ld.Config{
		DataSource: &testDataSourceBuilder{flags: flags},
		Events:     ldcomponents.NoEvents(),
	}

	client, err := ld.MakeCustomClient(t.Name(), ldConfig, 5*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })

	// Use a unique domain per test to avoid global OpenFeature provider conflicts.
	domain := "test_" + strings.ReplaceAll(t.Name(), "/", "_")
	provider := ofld.NewProvider(client)
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
		ldClient:       client,
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

		cfg := &Config{SDKKey: t.Name()}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, http.DefaultClient, cbnoop.NewCircuitBreaker(), func(config ld.Config) ld.Config {
			config.DataSource = &fakeLaunchDarklyDataSourceBuilder{}
			return config
		})
		require.NoError(t, err)
		require.NotNil(t, actual)
	})

	T.Run("with missing http client", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{SDKKey: t.Name()}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, nil, cbnoop.NewCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, actual)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		actual, err := NewFeatureFlagManager(nil, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, http.DefaultClient, cbnoop.NewCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, actual)
	})

	T.Run("with missing SDK key", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, http.DefaultClient, cbnoop.NewCircuitBreaker(), func(config ld.Config) ld.Config {
			config.DataSource = &fakeLaunchDarklyDataSourceBuilder{}
			return config
		})
		require.Error(t, err)
		require.Nil(t, actual)
	})

	T.Run("with zero init timeout gets default", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{SDKKey: t.Name(), InitTimeout: 0}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, http.DefaultClient, cbnoop.NewCircuitBreaker(), func(config ld.Config) ld.Config {
			config.DataSource = &fakeLaunchDarklyDataSourceBuilder{}
			return config
		})
		require.NoError(t, err)
		require.NotNil(t, actual)
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
		ffm := buildTestManagerWithFlags(t, testFlagItems())

		result, err := ffm.CanUseFeature(ctx, "bool-flag", evalCtx("user123"))
		assert.NoError(t, err)
		assert.True(t, result)
	})

	T.Run("with flag not found", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.CircuitBreakerMock{
			CanProceedFunc: func() bool { return true },
			SucceededFunc:  func() {},
			FailedFunc:     func() {},
		}

		ffm := buildTestManager(t, cb)

		result, err := ffm.CanUseFeature(ctx, "nonexistent-flag", evalCtx("user123"))
		assert.Error(t, err)
		assert.False(t, result)
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
		ffm := buildTestManagerWithFlags(t, testFlagItems())

		result, err := ffm.GetStringValue(ctx, "string-flag", "fallback", evalCtx("user123"))
		assert.NoError(t, err)
		assert.Equal(t, "hello-world", result)
	})

	T.Run("with flag not found", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.CircuitBreakerMock{
			CanProceedFunc: func() bool { return true },
			SucceededFunc:  func() {},
			FailedFunc:     func() {},
		}

		ffm := buildTestManager(t, cb)

		result, err := ffm.GetStringValue(ctx, "nonexistent-flag", "fallback", evalCtx("user123"))
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
		ffm := buildTestManagerWithFlags(t, testFlagItems())

		result, err := ffm.GetInt64Value(ctx, "int-flag", int64(0), evalCtx("user123"))
		assert.NoError(t, err)
		assert.Equal(t, int64(42), result)
	})

	T.Run("with flag not found", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.CircuitBreakerMock{
			CanProceedFunc: func() bool { return true },
			SucceededFunc:  func() {},
			FailedFunc:     func() {},
		}

		ffm := buildTestManager(t, cb)

		result, err := ffm.GetInt64Value(ctx, "nonexistent-flag", int64(42), evalCtx("user123"))
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
		ffm := buildTestManagerWithFlags(t, testFlagItems())

		result, err := ffm.GetFloat64Value(ctx, "float-flag", 0.0, evalCtx("user123"))
		assert.NoError(t, err)
		assert.InDelta(t, 3.14, result, 1e-9)
	})

	T.Run("with flag not found", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.CircuitBreakerMock{
			CanProceedFunc: func() bool { return true },
			SucceededFunc:  func() {},
			FailedFunc:     func() {},
		}

		ffm := buildTestManager(t, cb)

		result, err := ffm.GetFloat64Value(ctx, "nonexistent-flag", 3.14, evalCtx("user123"))
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
		ffm := buildTestManagerWithFlags(t, testFlagItems())

		def := map[string]any{"default": true}
		result, err := ffm.GetObjectValue(ctx, "object-flag", def, evalCtx("user123"))
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{"key": "value"}, result)
	})

	T.Run("with flag not found", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.CircuitBreakerMock{
			CanProceedFunc: func() bool { return true },
			SucceededFunc:  func() {},
			FailedFunc:     func() {},
		}

		ffm := buildTestManager(t, cb)

		def := map[string]any{"k": "v"}
		result, err := ffm.GetObjectValue(ctx, "nonexistent-flag", def, evalCtx("user123"))
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
