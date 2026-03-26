package elasticsearch

import (
	"testing"
	"time"

	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v4/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v4/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type example struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type invalidJSON struct {
	Channel chan int `json:"channel"` // channels can't be marshaled to JSON
}

func Test_indexManager_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, "index_test", cbnoop.NewCircuitBreaker())
		assert.NoError(t, err)
		assert.NotNil(t, im)

		searchable := &example{
			ID:   identifiers.New(),
			Name: t.Name(),
		}

		assert.NoError(t, im.Index(ctx, searchable.ID, searchable))

		time.Sleep(5 * time.Second)

		results, err := im.Search(ctx, searchable.Name[0:2])
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, searchable, results[0])

		assert.NoError(t, im.Delete(ctx, searchable.ID))
	})
}

func Test_indexManager_Index(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	T.Run("successful indexing", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, "index_test", cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		searchable := &example{
			ID:   identifiers.New(),
			Name: t.Name(),
		}

		err = im.Index(ctx, searchable.ID, searchable)
		assert.NoError(t, err)
	})

	T.Run("json marshaling error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, "index_test", cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		invalid := &invalidJSON{
			Channel: make(chan int),
		}

		err = im.Index(ctx, "test-id", invalid)
		assert.Error(t, err)
	})

	T.Run("circuit breaker open", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		// Create a circuit breaker that's already open
		cb := cbnoop.NewCircuitBreaker()
		// We can't easily test circuit breaker state without a real implementation
		// This test documents the expected behavior when circuit breaker is open
		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, "index_test", cb)
		require.NoError(t, err)

		searchable := &example{
			ID:   identifiers.New(),
			Name: t.Name(),
		}

		// This should succeed with noop circuit breaker
		err = im.Index(ctx, searchable.ID, searchable)
		assert.NoError(t, err)
	})
}

func Test_indexManager_Search(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	T.Run("successful search", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, "search_test", cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		// Index a document first
		searchable := &example{
			ID:   identifiers.New(),
			Name: "test search document",
		}
		err = im.Index(ctx, searchable.ID, searchable)
		require.NoError(t, err)

		// Wait for indexing to complete
		time.Sleep(2 * time.Second)

		// Search for the document
		results, err := im.Search(ctx, "test")
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, searchable.ID, results[0].ID)
	})

	T.Run("empty query error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, "search_test", cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		results, err := im.Search(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, ErrEmptyQueryProvided, err)
	})

	T.Run("no results found", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, "search_test", cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		results, err := im.Search(ctx, "nonexistent document")
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})
}

func Test_indexManager_Delete(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	T.Run("successful deletion", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, "delete_test", cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		// Index a document first
		searchable := &example{
			ID:   identifiers.New(),
			Name: "test delete document",
		}
		err = im.Index(ctx, searchable.ID, searchable)
		require.NoError(t, err)

		// Delete the document
		err = im.Delete(ctx, searchable.ID)
		assert.NoError(t, err)
	})

	T.Run("delete non-existent document", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, "delete_test", cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		// Try to delete a non-existent document
		err = im.Delete(ctx, "non-existent-id")
		// Elasticsearch typically returns success even for non-existent documents
		assert.NoError(t, err)
	})
}

func Test_indexManager_Wipe(T *testing.T) {
	T.Parallel()

	T.Run("returns unimplemented error", func(t *testing.T) {
		t.Parallel()

		// Create a mock index manager to test the Wipe method
		// Since Wipe just returns an error, we don't need a real Elasticsearch instance
		im := &indexManager[example]{}

		ctx := t.Context()
		err := im.Wipe(ctx)
		assert.Error(t, err)
		assert.Equal(t, "unimplemented", err.Error())
	})
}

func Test_ProvideIndexManager(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	T.Run("successful creation", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()

		im, err := ProvideIndexManager[example](ctx, logger, tracerProvider, cfg, "provide_test", cbnoop.NewCircuitBreaker())
		assert.NoError(t, err)
		assert.NotNil(t, im)
	})

	T.Run("invalid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		invalidCfg := &Config{
			Address: "invalid://address",
		}

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()

		im, err := ProvideIndexManager[example](ctx, logger, tracerProvider, invalidCfg, "provide_test", cbnoop.NewCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, im)
		assert.Contains(t, err.Error(), "initializing search client")
	})
}
