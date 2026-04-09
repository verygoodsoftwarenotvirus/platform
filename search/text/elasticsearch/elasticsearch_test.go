package elasticsearch

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	elasticsearchcontainers "github.com/testcontainers/testcontainers-go/modules/elasticsearch"
)

var runningContainerTests = strings.ToLower(os.Getenv("RUN_CONTAINER_TESTS")) == "true"

// esTestInfra holds a single shared Elasticsearch container for all container-
// backed subtests inside a package run. Subtests use unique index names to stay
// isolated, mirroring the qdrant/pgvector/distributedlock shared-container
// pattern.
type esTestInfra struct {
	cfg      *Config
	shutdown func(context.Context) error
}

func buildEsTestInfra(t *testing.T) *esTestInfra {
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

	return &esTestInfra{
		cfg:      cfg,
		shutdown: func(ctx context.Context) error { return elasticsearchContainer.Terminate(ctx) },
	}
}

// TestElasticsearch_Container holds every subtest that needs a real
// Elasticsearch container. They all share one container so we pay the
// pull/start cost once per package run. Each subtest creates its own
// index via unique identifiers.New() names to stay isolated.
func TestElasticsearch_Container(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	infra := buildEsTestInfra(T)
	T.Cleanup(func() { _ = infra.shutdown(context.Background()) })

	// --- ensureIndices ---

	T.Run("ensureIndices creates index when it does not exist", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		indexName := "ensure_create_" + identifiers.New()
		im, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, indexName, cbnoop.NewCircuitBreaker())
		require.NoError(t, err)
		assert.NotNil(t, im)

		searchable := &example{
			ID:   identifiers.New(),
			Name: "test document",
		}

		assert.NoError(t, im.Index(ctx, searchable.ID, searchable))
	})

	T.Run("ensureIndices handles existing index", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		indexName := "ensure_existing_" + identifiers.New()
		im1, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, indexName, cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		im2, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, indexName, cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		assert.NotNil(t, im1)
		assert.NotNil(t, im2)

		searchable := &example{
			ID:   identifiers.New(),
			Name: "test document",
		}

		assert.NoError(t, im1.Index(ctx, searchable.ID, searchable))
		assert.NoError(t, im2.Index(ctx, searchable.ID+"_2", searchable))
	})

	// --- ProvideIndexManager ---

	T.Run("ProvideIndexManager standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		im, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, "provide_"+identifiers.New(), cbnoop.NewCircuitBreaker())
		assert.NoError(t, err)
		assert.NotNil(t, im)
	})

	T.Run("ProvideIndexManager with logger and tracer", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()

		im, err := ProvideIndexManager[example](ctx, logger, tracerProvider, infra.cfg, "provide_lt_"+identifiers.New(), cbnoop.NewCircuitBreaker())
		assert.NoError(t, err)
		assert.NotNil(t, im)
	})

	// --- elasticsearchIsReadyToInit ---

	T.Run("elasticsearchIsReadyToInit returns true with valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		ready := elasticsearchIsReadyToInit(ctx, infra.cfg, logger, 5)
		assert.True(t, ready)
	})

	// --- provideElasticsearchClient ---

	T.Run("provideElasticsearchClient succeeds", func(t *testing.T) {
		t.Parallel()

		client, err := provideElasticsearchClient(infra.cfg)
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	// --- complete lifecycle ---

	T.Run("complete lifecycle", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		im, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, "lifecycle_"+identifiers.New(), cbnoop.NewCircuitBreaker())
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

	// --- Index ---

	T.Run("Index successful", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		im, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, "idx_ok_"+identifiers.New(), cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		searchable := &example{
			ID:   identifiers.New(),
			Name: t.Name(),
		}

		assert.NoError(t, im.Index(ctx, searchable.ID, searchable))
	})

	T.Run("Index json marshaling error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		im, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, "idx_json_"+identifiers.New(), cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		invalid := &invalidJSON{
			Channel: make(chan int),
		}

		assert.Error(t, im.Index(ctx, "test-id", invalid))
	})

	T.Run("Index with noop circuit breaker", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := cbnoop.NewCircuitBreaker()
		im, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, "idx_cb_"+identifiers.New(), cb)
		require.NoError(t, err)

		searchable := &example{
			ID:   identifiers.New(),
			Name: t.Name(),
		}

		assert.NoError(t, im.Index(ctx, searchable.ID, searchable))
	})

	// --- Search ---

	T.Run("Search successful", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		im, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, "search_ok_"+identifiers.New(), cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		searchable := &example{
			ID:   identifiers.New(),
			Name: "test search document",
		}
		require.NoError(t, im.Index(ctx, searchable.ID, searchable))

		time.Sleep(2 * time.Second)

		results, err := im.Search(ctx, "test")
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, searchable.ID, results[0].ID)
	})

	T.Run("Search empty query error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		im, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, "search_empty_"+identifiers.New(), cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		results, err := im.Search(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, ErrEmptyQueryProvided, err)
	})

	T.Run("Search no results found", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		im, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, "search_noresult_"+identifiers.New(), cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		results, err := im.Search(ctx, "nonexistent document")
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	// --- Delete ---

	T.Run("Delete successful", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		im, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, "del_ok_"+identifiers.New(), cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		searchable := &example{
			ID:   identifiers.New(),
			Name: "test delete document",
		}
		require.NoError(t, im.Index(ctx, searchable.ID, searchable))

		assert.NoError(t, im.Delete(ctx, searchable.ID))
	})

	T.Run("Delete non-existent document", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		im, err := ProvideIndexManager[example](ctx, nil, nil, infra.cfg, "del_nf_"+identifiers.New(), cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		assert.NoError(t, im.Delete(ctx, "non-existent-id"))
	})

	// --- Wipe ---

	T.Run("Wipe returns unimplemented error", func(t *testing.T) {
		t.Parallel()

		im := &indexManager[example]{}

		assert.Error(t, im.Wipe(t.Context()))
		assert.Equal(t, "unimplemented", im.Wipe(t.Context()).Error())
	})
}
