package objectstorage

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking"
	cbmock "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob/memblob"
)

func noopUploaderMetrics(t *testing.T) (saveCounter, readCounter, saveErrCounter, readErrCounter metrics.Int64Counter, latencyHist metrics.Float64Histogram) {
	t.Helper()
	mp := metrics.NewNoopMetricsProvider()

	saveCounter, err := mp.NewInt64Counter("test_saves")
	require.NoError(t, err)

	readCounter, err = mp.NewInt64Counter("test_reads")
	require.NoError(t, err)

	saveErrCounter, err = mp.NewInt64Counter("test_save_errors")
	require.NoError(t, err)

	readErrCounter, err = mp.NewInt64Counter("test_read_errors")
	require.NoError(t, err)

	latencyHist, err = mp.NewFloat64Histogram("test_latency")
	require.NoError(t, err)

	return saveCounter, readCounter, saveErrCounter, readErrCounter, latencyHist
}

func TestUploader_ReadFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleFilename := "hello_world.txt"
		expectedContent := []byte(t.Name())

		b := memblob.OpenBucket(&memblob.Options{})
		require.NoError(t, b.WriteAll(ctx, exampleFilename, expectedContent, nil))

		saveCounter, readCounter, saveErrCounter, readErrCounter, latencyHist := noopUploaderMetrics(t)
		u := &Uploader{
			bucket:         b,
			logger:         logging.NewNoopLogger(),
			tracer:         tracing.NewTracerForTest(t.Name()),
			circuitBreaker: noop.NewCircuitBreaker(),
			saveCounter:    saveCounter,
			readCounter:    readCounter,
			saveErrCounter: saveErrCounter,
			readErrCounter: readErrCounter,
			latencyHist:    latencyHist,
		}

		x, err := u.ReadFile(ctx, exampleFilename)
		assert.NoError(t, err)
		assert.Equal(t, expectedContent, x)
	})

	T.Run("with invalid file", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleFilename := "hello_world.txt"

		saveCounter, readCounter, saveErrCounter, readErrCounter, latencyHist := noopUploaderMetrics(t)
		u := &Uploader{
			bucket:         memblob.OpenBucket(&memblob.Options{}),
			logger:         logging.NewNoopLogger(),
			tracer:         tracing.NewTracerForTest(t.Name()),
			circuitBreaker: noop.NewCircuitBreaker(),
			saveCounter:    saveCounter,
			readCounter:    readCounter,
			saveErrCounter: saveErrCounter,
			readErrCounter: readErrCounter,
			latencyHist:    latencyHist,
		}

		x, err := u.ReadFile(ctx, exampleFilename)
		assert.Error(t, err)
		assert.Nil(t, x)
	})

	T.Run("with broken circuit breaker", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cb := &cbmock.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(true)

		saveCounter, readCounter, saveErrCounter, readErrCounter, latencyHist := noopUploaderMetrics(t)
		u := &Uploader{
			bucket:         memblob.OpenBucket(&memblob.Options{}),
			logger:         logging.NewNoopLogger(),
			tracer:         tracing.NewTracerForTest(t.Name()),
			circuitBreaker: cb,
			saveCounter:    saveCounter,
			readCounter:    readCounter,
			saveErrCounter: saveErrCounter,
			readErrCounter: readErrCounter,
			latencyHist:    latencyHist,
		}

		x, err := u.ReadFile(ctx, "anything.txt")
		assert.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)
		assert.Nil(t, x)
	})

	T.Run("with mock circuit breaker on successful read", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleFilename := "hello_world.txt"
		expectedContent := []byte(t.Name())

		b := memblob.OpenBucket(&memblob.Options{})
		require.NoError(t, b.WriteAll(ctx, exampleFilename, expectedContent, nil))

		cb := &cbmock.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)
		cb.On("Succeeded").Return()

		saveCounter, readCounter, saveErrCounter, readErrCounter, latencyHist := noopUploaderMetrics(t)
		u := &Uploader{
			bucket:         b,
			logger:         logging.NewNoopLogger(),
			tracer:         tracing.NewTracerForTest(t.Name()),
			circuitBreaker: cb,
			saveCounter:    saveCounter,
			readCounter:    readCounter,
			saveErrCounter: saveErrCounter,
			readErrCounter: readErrCounter,
			latencyHist:    latencyHist,
		}

		x, err := u.ReadFile(ctx, exampleFilename)
		assert.NoError(t, err)
		assert.Equal(t, expectedContent, x)
	})
}

func TestUploader_SaveFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		saveCounter, readCounter, saveErrCounter, readErrCounter, latencyHist := noopUploaderMetrics(t)
		u := &Uploader{
			bucket:         memblob.OpenBucket(&memblob.Options{}),
			logger:         logging.NewNoopLogger(),
			tracer:         tracing.NewTracerForTest(t.Name()),
			circuitBreaker: noop.NewCircuitBreaker(),
			saveCounter:    saveCounter,
			readCounter:    readCounter,
			saveErrCounter: saveErrCounter,
			readErrCounter: readErrCounter,
			latencyHist:    latencyHist,
		}

		assert.NoError(t, u.SaveFile(ctx, "test_file.txt", []byte(t.Name())))
	})

	T.Run("with broken circuit breaker", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cb := &cbmock.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(true)

		saveCounter, readCounter, saveErrCounter, readErrCounter, latencyHist := noopUploaderMetrics(t)
		u := &Uploader{
			bucket:         memblob.OpenBucket(&memblob.Options{}),
			logger:         logging.NewNoopLogger(),
			tracer:         tracing.NewTracerForTest(t.Name()),
			circuitBreaker: cb,
			saveCounter:    saveCounter,
			readCounter:    readCounter,
			saveErrCounter: saveErrCounter,
			readErrCounter: readErrCounter,
			latencyHist:    latencyHist,
		}

		assert.ErrorIs(t, u.SaveFile(ctx, "test_file.txt", []byte(t.Name())), circuitbreaking.ErrCircuitBroken)
	})

	T.Run("with write error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cb := &cbmock.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)
		cb.On("Failed").Return()

		b := memblob.OpenBucket(&memblob.Options{})
		require.NoError(t, b.Close())

		saveCounter, readCounter, saveErrCounter, readErrCounter, latencyHist := noopUploaderMetrics(t)
		u := &Uploader{
			bucket:         b,
			logger:         logging.NewNoopLogger(),
			tracer:         tracing.NewTracerForTest(t.Name()),
			circuitBreaker: cb,
			saveCounter:    saveCounter,
			readCounter:    readCounter,
			saveErrCounter: saveErrCounter,
			readErrCounter: readErrCounter,
			latencyHist:    latencyHist,
		}

		assert.Error(t, u.SaveFile(ctx, "test_file.txt", []byte(t.Name())))
	})

	T.Run("can be read back after save", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		content := []byte("hello world")

		saveCounter, readCounter, saveErrCounter, readErrCounter, latencyHist := noopUploaderMetrics(t)
		u := &Uploader{
			bucket:         memblob.OpenBucket(&memblob.Options{}),
			logger:         logging.NewNoopLogger(),
			tracer:         tracing.NewTracerForTest(t.Name()),
			circuitBreaker: noop.NewCircuitBreaker(),
			saveCounter:    saveCounter,
			readCounter:    readCounter,
			saveErrCounter: saveErrCounter,
			readErrCounter: readErrCounter,
			latencyHist:    latencyHist,
		}

		require.NoError(t, u.SaveFile(ctx, "roundtrip.txt", content))

		actual, err := u.ReadFile(ctx, "roundtrip.txt")
		assert.NoError(t, err)
		assert.Equal(t, content, actual)
	})
}
