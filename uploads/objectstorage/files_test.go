package objectstorage

import (
	"os"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob/memblob"
)

func noopUploaderMetrics(t *testing.T) (metrics.Int64Counter, metrics.Int64Counter, metrics.Int64Counter, metrics.Int64Counter, metrics.Float64Histogram) {
	t.Helper()
	mp := metrics.NewNoopMetricsProvider()
	saveCounter, _ := mp.NewInt64Counter("test_saves")
	readCounter, _ := mp.NewInt64Counter("test_reads")
	saveErrCounter, _ := mp.NewInt64Counter("test_save_errors")
	readErrCounter, _ := mp.NewInt64Counter("test_read_errors")
	latencyHist, _ := mp.NewFloat64Histogram("test_latency")
	return saveCounter, readCounter, saveErrCounter, readErrCounter, latencyHist
}

func TestUploader_ReadFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleFilename := "hello_world.txt"

		b := memblob.OpenBucket(&memblob.Options{})
		require.NoError(t, b.WriteAll(ctx, exampleFilename, []byte(t.Name()), nil))

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
		assert.NotNil(t, x)
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
}

func TestUploader_SaveFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		tempFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

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

		assert.NoError(t, u.SaveFile(ctx, tempFile.Name(), []byte(t.Name())))
	})
}
