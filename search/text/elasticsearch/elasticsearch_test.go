package elasticsearch

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	elasticsearchcontainers "github.com/testcontainers/testcontainers-go/modules/elasticsearch"
)

var (
	runningContainerTests = strings.ToLower(os.Getenv("RUN_CONTAINER_TESTS")) == "true"
)

func buildContainerBackedElasticsearchConfig(t *testing.T) (config *Config, shutdownFunction func(context.Context) error) {
	t.Helper()

	elasticsearchContainer, err := elasticsearchcontainers.Run(
		t.Context(),
		"elasticsearch:8.10.2",
		elasticsearchcontainers.WithPassword("arbitraryPassword"),
	)
	require.NoError(t, err)
	require.NotNil(t, elasticsearchContainer)

	cfg := &Config{
		Address:               elasticsearchContainer.Settings.Address,
		IndexOperationTimeout: 0,
		Username:              "elastic",
		Password:              elasticsearchContainer.Settings.Password,
		CACert:                elasticsearchContainer.Settings.CACert,
	}

	return cfg, func(ctx context.Context) error { return elasticsearchContainer.Terminate(ctx) }
}

func Test_ensureIndices(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	T.Run("creates index when it doesn't exist", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		// Create index manager with a unique index name
		indexName := "ensure_indices_test_" + t.Name()
		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, indexName, circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)
		assert.NotNil(t, im)

		// The ensureIndices method is called during ProvideIndexManager
		// This test verifies that the index was created successfully
		// by attempting to index a document
		searchable := &example{
			ID:   identifiers.New(),
			Name: "test document",
		}

		err = im.Index(ctx, searchable.ID, searchable)
		assert.NoError(t, err)
	})

	T.Run("handles existing index", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		// Create first index manager
		indexName := "ensure_indices_existing_test_" + t.Name()
		im1, err := ProvideIndexManager[example](ctx, nil, nil, cfg, indexName, circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)

		// Create second index manager with same index name
		im2, err := ProvideIndexManager[example](ctx, nil, nil, cfg, indexName, circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)

		assert.NotNil(t, im1)
		assert.NotNil(t, im2)

		// Both should work fine since the index already exists
		searchable := &example{
			ID:   identifiers.New(),
			Name: "test document",
		}

		err = im1.Index(ctx, searchable.ID, searchable)
		assert.NoError(t, err)

		err = im2.Index(ctx, searchable.ID+"_2", searchable)
		assert.NoError(t, err)
	})
}

func Test_ProvideIndexManager_Container(T *testing.T) {
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

		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, t.Name(), circuitbreaking.NewNoopCircuitBreaker())
		assert.NoError(t, err)
		assert.NotNil(t, im)
	})

	T.Run("without available instance", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{}

		im, err := ProvideIndexManager[example](ctx, nil, nil, cfg, t.Name(), circuitbreaking.NewNoopCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, im)
	})
}

func Test_elasticsearchIsReadyToInit(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	T.Run("returns true with valid config and server", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		logger := logging.NewNoopLogger()

		ready := elasticsearchIsReadyToInit(ctx, cfg, logger, 5)
		assert.True(t, ready)
	})

	T.Run("returns false with invalid address", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		invalidCfg := &Config{
			Address: "http://localhost:9999", // Non-existent server
		}
		logger := logging.NewNoopLogger()

		ready := elasticsearchIsReadyToInit(ctx, invalidCfg, logger, 1)
		assert.False(t, ready)
	})
}

func Test_provideElasticsearchClient(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	T.Run("successful client creation", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg, shutdownFunc := buildContainerBackedElasticsearchConfig(t)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		client, err := provideElasticsearchClient(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	T.Run("handles invalid address", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Address: "invalid://address",
		}

		client, err := provideElasticsearchClient(cfg)
		// The Elasticsearch client doesn't validate addresses during creation
		// It only fails when making actual requests
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}
