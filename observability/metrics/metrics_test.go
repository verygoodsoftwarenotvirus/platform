package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

func TestEnsureMetricsProvider(T *testing.T) {
	T.Parallel()

	T.Run("returns provided provider when non-nil", func(t *testing.T) {
		t.Parallel()

		p := NewNoopMetricsProvider()
		actual := EnsureMetricsProvider(p)
		assert.Equal(t, p, actual)
	})

	T.Run("returns noop provider when nil", func(t *testing.T) {
		t.Parallel()

		actual := EnsureMetricsProvider(nil)
		assert.NotNil(t, actual)
	})
}

func TestNoopProvider(T *testing.T) {
	T.Parallel()

	T.Run("NewFloat64Counter", func(t *testing.T) {
		t.Parallel()
		p := NewNoopMetricsProvider()
		c, err := p.NewFloat64Counter("test_counter")
		require.NoError(t, err)
		assert.NotNil(t, c)
		c.Add(context.Background(), 1.0)
	})

	T.Run("NewFloat64Gauge", func(t *testing.T) {
		t.Parallel()
		p := NewNoopMetricsProvider()
		g, err := p.NewFloat64Gauge("test_gauge")
		require.NoError(t, err)
		assert.NotNil(t, g)
		g.Record(context.Background(), 1.0)
	})

	T.Run("NewFloat64UpDownCounter", func(t *testing.T) {
		t.Parallel()
		p := NewNoopMetricsProvider()
		c, err := p.NewFloat64UpDownCounter("test_updown")
		require.NoError(t, err)
		assert.NotNil(t, c)
		c.Add(context.Background(), -1.0)
	})

	T.Run("NewFloat64Histogram", func(t *testing.T) {
		t.Parallel()
		p := NewNoopMetricsProvider()
		h, err := p.NewFloat64Histogram("test_histogram")
		require.NoError(t, err)
		assert.NotNil(t, h)
		h.Record(context.Background(), 1.0)
	})

	T.Run("NewInt64Counter", func(t *testing.T) {
		t.Parallel()
		p := NewNoopMetricsProvider()
		c, err := p.NewInt64Counter("test_counter")
		require.NoError(t, err)
		assert.NotNil(t, c)
		c.Add(context.Background(), 1)
	})

	T.Run("NewInt64Gauge", func(t *testing.T) {
		t.Parallel()
		p := NewNoopMetricsProvider()
		g, err := p.NewInt64Gauge("test_gauge")
		require.NoError(t, err)
		assert.NotNil(t, g)
		g.Record(context.Background(), 1)
	})

	T.Run("NewInt64UpDownCounter", func(t *testing.T) {
		t.Parallel()
		p := NewNoopMetricsProvider()
		c, err := p.NewInt64UpDownCounter("test_updown")
		require.NoError(t, err)
		assert.NotNil(t, c)
		c.Add(context.Background(), -1)
	})

	T.Run("NewInt64Histogram", func(t *testing.T) {
		t.Parallel()
		p := NewNoopMetricsProvider()
		h, err := p.NewInt64Histogram("test_histogram")
		require.NoError(t, err)
		assert.NotNil(t, h)
		h.Record(context.Background(), 1)
	})

	T.Run("Shutdown", func(t *testing.T) {
		t.Parallel()
		p := NewNoopMetricsProvider()
		err := p.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	T.Run("MeterProvider", func(t *testing.T) {
		t.Parallel()
		p := NewNoopMetricsProvider()
		mp := p.MeterProvider()
		assert.NotNil(t, mp)
	})
}

func TestInt64CounterForTest(T *testing.T) {
	T.Parallel()

	T.Run("returns a counter", func(t *testing.T) {
		t.Parallel()
		c := Int64CounterForTest("test_counter")
		assert.NotNil(t, c)
	})
}

func TestImplWrappers(T *testing.T) {
	T.Parallel()

	ctx := context.Background()
	meter := otel.Meter("test")

	T.Run("Float64CounterImpl", func(t *testing.T) {
		t.Parallel()
		x, err := meter.Float64Counter("test_f64_counter")
		require.NoError(t, err)
		impl := &Float64CounterImpl{X: x}
		impl.Add(ctx, 1.0)
	})

	T.Run("Float64GaugeImpl", func(t *testing.T) {
		t.Parallel()
		x, err := meter.Float64Gauge("test_f64_gauge")
		require.NoError(t, err)
		impl := &Float64GaugeImpl{X: x}
		impl.Record(ctx, 1.0)
	})

	T.Run("Float64UpDownCounterImpl", func(t *testing.T) {
		t.Parallel()
		x, err := meter.Float64UpDownCounter("test_f64_updown")
		require.NoError(t, err)
		impl := &Float64UpDownCounterImpl{X: x}
		impl.Add(ctx, -1.0)
	})

	T.Run("Float64HistogramImpl", func(t *testing.T) {
		t.Parallel()
		x, err := meter.Float64Histogram("test_f64_histogram")
		require.NoError(t, err)
		impl := &Float64HistogramImpl{X: x}
		impl.Record(ctx, 1.0)
	})

	T.Run("Int64CounterImpl", func(t *testing.T) {
		t.Parallel()
		x, err := meter.Int64Counter("test_i64_counter")
		require.NoError(t, err)
		impl := &Int64CounterImpl{X: x}
		impl.Add(ctx, 1)
	})

	T.Run("Int64GaugeImpl", func(t *testing.T) {
		t.Parallel()
		x, err := meter.Int64Gauge("test_i64_gauge")
		require.NoError(t, err)
		impl := &Int64GaugeImpl{X: x}
		impl.Record(ctx, 1)
	})

	T.Run("Int64UpDownCounterImpl", func(t *testing.T) {
		t.Parallel()
		x, err := meter.Int64UpDownCounter("test_i64_updown")
		require.NoError(t, err)
		impl := &Int64UpDownCounterImpl{X: x}
		impl.Add(ctx, -1)
	})

	T.Run("Int64HistogramImpl", func(t *testing.T) {
		t.Parallel()
		x, err := meter.Int64Histogram("test_i64_histogram")
		require.NoError(t, err)
		impl := &Int64HistogramImpl{X: x}
		impl.Record(ctx, 1)
	})
}
