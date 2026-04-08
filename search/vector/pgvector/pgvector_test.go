package pgvector

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/database"
	"github.com/verygoodsoftwarenotvirus/platform/v5/identifiers"
	vectorsearch "github.com/verygoodsoftwarenotvirus/platform/v5/search/vector"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const pgvectorImage = "pgvector/pgvector:pg16"

var runningContainerTests = strings.ToLower(os.Getenv("RUN_CONTAINER_TESTS")) == "true"

// testDBClient is a minimal database.Client backed by a single *sql.DB. It exists
// only to avoid pulling in database/postgres for tests in this leaf package.
type testDBClient struct {
	db *sql.DB
}

func (c *testDBClient) WriteDB() *sql.DB { return c.db }
func (c *testDBClient) ReadDB() *sql.DB  { return c.db }
func (c *testDBClient) Close() error     { return c.db.Close() }
func (c *testDBClient) CurrentTime() time.Time {
	return time.Now()
}
func (c *testDBClient) RollbackTransaction(_ context.Context, tx database.SQLQueryExecutorAndTransactionManager) {
	_ = tx.Rollback()
}

func buildContainerBackedPgvector(t *testing.T) (client *testDBClient, shutdown func(context.Context) error) {
	t.Helper()

	ctx := t.Context()
	container, err := postgrescontainer.Run(
		ctx,
		pgvectorImage,
		postgrescontainer.WithDatabase("vectortest"),
		postgrescontainer.WithUsername("vectortest"),
		postgrescontainer.WithPassword("vectortest"),
		testcontainers.WithWaitStrategyAndDeadline(2*time.Minute, wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
	)
	require.NoError(t, err)
	require.NotNil(t, container)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	require.NoError(t, db.PingContext(ctx))

	return &testDBClient{db: db}, func(ctx context.Context) error {
		_ = db.Close()
		return container.Terminate(ctx)
	}
}

type doc struct {
	Kind  string `json:"kind"`
	Title string `json:"title"`
}

func provideTestIndex(t *testing.T, client database.Client, indexName string, dim int, metric vectorsearch.DistanceMetric) vectorsearch.Index[doc] {
	t.Helper()

	cfg := &Config{
		Dimension: dim,
		Metric:    metric,
	}
	im, err := ProvideIndex[doc](t.Context(), nil, nil, nil, cfg, client, indexName, cbnoop.NewCircuitBreaker())
	require.NoError(t, err)
	require.NotNil(t, im)
	return im
}

func TestProvideIndex(T *testing.T) {
	T.Parallel()

	T.Run("nil config", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideIndex[doc](t.Context(), nil, nil, nil, nil, &testDBClient{}, "idx", cbnoop.NewCircuitBreaker())
		require.ErrorIs(t, err, vectorsearch.ErrNilConfig)
	})

	T.Run("nil database", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideIndex[doc](t.Context(), nil, nil, nil, &Config{Dimension: 3, Metric: vectorsearch.DistanceCosine}, nil, "idx", cbnoop.NewCircuitBreaker())
		require.ErrorIs(t, err, vectorsearch.ErrNilDatabaseClient)
	})

	T.Run("invalid dimension", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideIndex[doc](t.Context(), nil, nil, nil, &Config{Dimension: 0, Metric: vectorsearch.DistanceCosine}, &testDBClient{}, "idx", cbnoop.NewCircuitBreaker())
		require.Error(t, err)
	})

	T.Run("invalid metric", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideIndex[doc](t.Context(), nil, nil, nil, &Config{Dimension: 3, Metric: "weird"}, &testDBClient{}, "idx", cbnoop.NewCircuitBreaker())
		require.Error(t, err)
	})

	T.Run("invalid index name", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideIndex[doc](t.Context(), nil, nil, nil, &Config{Dimension: 3, Metric: vectorsearch.DistanceCosine}, &testDBClient{}, "no-dashes", cbnoop.NewCircuitBreaker())
		require.ErrorIs(t, err, ErrInvalidIdentifier)
	})

	T.Run("invalid metadata column", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideIndex[doc](t.Context(), nil, nil, nil, &Config{Dimension: 3, Metric: vectorsearch.DistanceCosine, MetadataColumn: "weird-col"}, &testDBClient{}, "idx", cbnoop.NewCircuitBreaker())
		require.ErrorIs(t, err, ErrInvalidIdentifier)
	})
}

func TestEncodeVector(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "[0.1,0.2,0.3]", encodeVector([]float32{0.1, 0.2, 0.3}))
	})

	T.Run("empty", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "[]", encodeVector(nil))
	})

	T.Run("integer-valued", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "[1,2,3]", encodeVector([]float32{1, 2, 3}))
	})
}

func TestQuoteIdent(T *testing.T) {
	T.Parallel()

	T.Run("simple", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `"users"`, quoteIdent("users"))
	})

	T.Run("with embedded quote", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `"foo""bar"`, quoteIdent(`foo"bar`))
	})
}

func TestPgTextArray(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `{"a","b","c"}`, pgTextArray([]string{"a", "b", "c"}))
	})

	T.Run("with quotes", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `{"a\"b","c"}`, pgTextArray([]string{`a"b`, "c"}))
	})
}

func TestMarshalUnmarshalMetadata(T *testing.T) {
	T.Parallel()

	T.Run("nil round-trip", func(t *testing.T) {
		t.Parallel()
		raw, err := marshalMetadata[doc](nil)
		require.NoError(t, err)
		assert.Equal(t, []byte(`{}`), raw)

		out, err := unmarshalMetadata[doc](raw)
		require.NoError(t, err)
		require.NotNil(t, out)
	})

	T.Run("populated round-trip", func(t *testing.T) {
		t.Parallel()
		original := &doc{Kind: "doc", Title: "hello"}
		raw, err := marshalMetadata(original)
		require.NoError(t, err)

		out, err := unmarshalMetadata[doc](raw)
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, *original, *out)
	})

	T.Run("null is treated as nil", func(t *testing.T) {
		t.Parallel()
		out, err := unmarshalMetadata[doc]([]byte("null"))
		require.NoError(t, err)
		assert.Nil(t, out)
	})
}

func TestPgvectorIndex_Container(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	client, shutdown := buildContainerBackedPgvector(T)
	T.Cleanup(func() {
		_ = shutdown(context.Background())
	})

	T.Run("Upsert and Query roundtrip", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provideTestIndex(t, client, "rt_"+identifiers.New(), 3, vectorsearch.DistanceCosine)

		require.NoError(t, idx.Upsert(ctx,
			vectorsearch.Vector[doc]{ID: "a", Embedding: []float32{1, 0, 0}, Metadata: &doc{Kind: "doc", Title: "alpha"}},
			vectorsearch.Vector[doc]{ID: "b", Embedding: []float32{0, 1, 0}, Metadata: &doc{Kind: "doc", Title: "beta"}},
			vectorsearch.Vector[doc]{ID: "c", Embedding: []float32{0, 0, 1}, Metadata: &doc{Kind: "doc", Title: "gamma"}},
		))

		results, err := idx.Query(ctx, vectorsearch.QueryRequest{
			Embedding: []float32{1, 0, 0},
			TopK:      3,
		})
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, "a", results[0].ID)
		require.NotNil(t, results[0].Metadata)
		assert.Equal(t, "alpha", results[0].Metadata.Title)
		assert.InDelta(t, 0.0, results[0].Distance, 1e-5)
	})

	T.Run("Upsert updates existing row", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provideTestIndex(t, client, "upd_"+identifiers.New(), 3, vectorsearch.DistanceCosine)

		require.NoError(t, idx.Upsert(ctx, vectorsearch.Vector[doc]{ID: "x", Embedding: []float32{1, 0, 0}, Metadata: &doc{Title: "first"}}))
		require.NoError(t, idx.Upsert(ctx, vectorsearch.Vector[doc]{ID: "x", Embedding: []float32{0, 1, 0}, Metadata: &doc{Title: "second"}}))

		results, err := idx.Query(ctx, vectorsearch.QueryRequest{
			Embedding: []float32{0, 1, 0},
			TopK:      1,
		})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "x", results[0].ID)
		require.NotNil(t, results[0].Metadata)
		assert.Equal(t, "second", results[0].Metadata.Title)
	})

	T.Run("TopK is respected", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provideTestIndex(t, client, "topk_"+identifiers.New(), 3, vectorsearch.DistanceCosine)

		require.NoError(t, idx.Upsert(ctx,
			vectorsearch.Vector[doc]{ID: "a", Embedding: []float32{1, 0, 0}},
			vectorsearch.Vector[doc]{ID: "b", Embedding: []float32{0, 1, 0}},
			vectorsearch.Vector[doc]{ID: "c", Embedding: []float32{0, 0, 1}},
		))

		results, err := idx.Query(ctx, vectorsearch.QueryRequest{
			Embedding: []float32{1, 0, 0},
			TopK:      2,
		})
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	T.Run("filter clause is applied", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provideTestIndex(t, client, "filt_"+identifiers.New(), 3, vectorsearch.DistanceCosine)

		require.NoError(t, idx.Upsert(ctx,
			vectorsearch.Vector[doc]{ID: "a", Embedding: []float32{1, 0, 0}, Metadata: &doc{Kind: "doc"}},
			vectorsearch.Vector[doc]{ID: "b", Embedding: []float32{1, 0, 0}, Metadata: &doc{Kind: "image"}},
		))

		results, err := idx.Query(ctx, vectorsearch.QueryRequest{
			Embedding: []float32{1, 0, 0},
			TopK:      10,
			Filter:    "metadata->>'kind' = 'doc'",
		})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "a", results[0].ID)
	})

	T.Run("Query rejects empty embedding", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provideTestIndex(t, client, "emb_"+identifiers.New(), 3, vectorsearch.DistanceCosine)

		_, err := idx.Query(ctx, vectorsearch.QueryRequest{Embedding: nil, TopK: 5})
		require.ErrorIs(t, err, vectorsearch.ErrEmptyEmbedding)
	})

	T.Run("Query rejects wrong dimension", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provideTestIndex(t, client, "dim_"+identifiers.New(), 3, vectorsearch.DistanceCosine)

		_, err := idx.Query(ctx, vectorsearch.QueryRequest{Embedding: []float32{1, 0}, TopK: 5})
		require.ErrorIs(t, err, vectorsearch.ErrDimensionMismatch)
	})

	T.Run("Upsert rejects wrong dimension", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provideTestIndex(t, client, "udim_"+identifiers.New(), 3, vectorsearch.DistanceCosine)

		err := idx.Upsert(ctx, vectorsearch.Vector[doc]{ID: "a", Embedding: []float32{1, 0}})
		require.ErrorIs(t, err, vectorsearch.ErrDimensionMismatch)
	})

	T.Run("Delete removes specific rows", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provideTestIndex(t, client, "del_"+identifiers.New(), 3, vectorsearch.DistanceCosine)

		require.NoError(t, idx.Upsert(ctx,
			vectorsearch.Vector[doc]{ID: "a", Embedding: []float32{1, 0, 0}},
			vectorsearch.Vector[doc]{ID: "b", Embedding: []float32{0, 1, 0}},
			vectorsearch.Vector[doc]{ID: "c", Embedding: []float32{0, 0, 1}},
		))
		require.NoError(t, idx.Delete(ctx, "a", "c"))

		results, err := idx.Query(ctx, vectorsearch.QueryRequest{Embedding: []float32{0, 1, 0}, TopK: 10})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "b", results[0].ID)
	})

	T.Run("Wipe empties the index", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provideTestIndex(t, client, "wipe_"+identifiers.New(), 3, vectorsearch.DistanceCosine)

		require.NoError(t, idx.Upsert(ctx,
			vectorsearch.Vector[doc]{ID: "a", Embedding: []float32{1, 0, 0}},
			vectorsearch.Vector[doc]{ID: "b", Embedding: []float32{0, 1, 0}},
		))
		require.NoError(t, idx.Wipe(ctx))

		results, err := idx.Query(ctx, vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 10})
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	T.Run("ProvideIndex is idempotent for the same index", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		name := "idem_" + identifiers.New()
		idx1 := provideTestIndex(t, client, name, 3, vectorsearch.DistanceCosine)
		idx2 := provideTestIndex(t, client, name, 3, vectorsearch.DistanceCosine)
		assert.NotNil(t, idx1)
		assert.NotNil(t, idx2)

		require.NoError(t, idx1.Upsert(ctx, vectorsearch.Vector[doc]{ID: "shared", Embedding: []float32{1, 0, 0}}))

		results, err := idx2.Query(ctx, vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 1})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "shared", results[0].ID)
	})
}
