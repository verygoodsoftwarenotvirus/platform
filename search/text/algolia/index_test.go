package algolia

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking"
	mockcircuitbreaking "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/mock"
	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	algoliasearch "github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	algoliatransport "github.com/algolia/algoliasearch-client-go/v3/algolia/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ algoliatransport.Requester = (*testRequester)(nil)

type testRequester struct {
	handler http.Handler
}

func (r *testRequester) Request(req *http.Request) (*http.Response, error) {
	recorder := &responseRecorder{
		headers: http.Header{},
		body:    &strings.Builder{},
		code:    http.StatusOK,
	}
	r.handler.ServeHTTP(recorder, req)

	return &http.Response{
		StatusCode: recorder.code,
		Header:     recorder.headers,
		Body:       io.NopCloser(strings.NewReader(recorder.body.String())),
	}, nil
}

type responseRecorder struct {
	headers http.Header
	body    *strings.Builder
	code    int
}

func (r *responseRecorder) Header() http.Header {
	return r.headers
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.code = statusCode
}

func buildTestIndexManagerWithMockServer(t *testing.T, handler http.Handler, cb circuitbreaking.CircuitBreaker) *indexManager[example] {
	t.Helper()

	client := algoliasearch.NewClientWithConfig(algoliasearch.Configuration{
		AppID:     "fake",
		APIKey:    "fake",
		Hosts:     []string{"localhost"},
		Requester: &testRequester{handler: handler},
	})

	return &indexManager[example]{
		logger:         logging.NewNoopLogger(),
		tracer:         tracing.NewTracerForTest("test"),
		circuitBreaker: cb,
		client:         client.InitIndex("test"),
	}
}

func buildTestIndexManager(t *testing.T) *indexManager[example] {
	t.Helper()

	im, err := ProvideIndexManager[example](
		logging.NewNoopLogger(),
		tracing.NewNoopTracerProvider(),
		&Config{AppID: "fake", APIKey: "fake"},
		"test",
		cbnoop.NewCircuitBreaker(),
	)
	if err != nil {
		t.Fatal(err)
	}

	return im.(*indexManager[example])
}

func buildTestIndexManagerWithCircuitBreaker(t *testing.T, cb circuitbreaking.CircuitBreaker) *indexManager[example] {
	t.Helper()

	im, err := ProvideIndexManager[example](
		logging.NewNoopLogger(),
		tracing.NewNoopTracerProvider(),
		&Config{AppID: "fake", APIKey: "fake"},
		"test",
		cb,
	)
	if err != nil {
		t.Fatal(err)
	}

	return im.(*indexManager[example])
}

func TestIndexManager_Index(T *testing.T) {
	T.Parallel()

	T.Run("with broken circuit breaker", func(t *testing.T) {
		t.Parallel()

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(true)

		im := buildTestIndexManagerWithCircuitBreaker(t, cb)

		err := im.Index(context.Background(), "id", map[string]string{"id": "test"})
		assert.Error(t, err)
		assert.Equal(t, circuitbreaking.ErrCircuitBroken, err)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with unmarshalable value", func(t *testing.T) {
		t.Parallel()

		im := buildTestIndexManager(t)

		err := im.Index(context.Background(), "id", make(chan int))
		assert.Error(t, err)
	})

	T.Run("with valid value but invalid credentials", func(t *testing.T) {
		t.Parallel()

		im := buildTestIndexManager(t)

		err := im.Index(context.Background(), "id", map[string]string{"id": "test", "name": "example"})
		assert.Error(t, err)
	})

	T.Run("with non-object JSON value", func(t *testing.T) {
		t.Parallel()

		im := buildTestIndexManager(t)

		err := im.Index(context.Background(), "id", "just a string")
		assert.Error(t, err)
	})

	T.Run("with successful index", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"createdAt":"2021-01-01T00:00:00Z","objectID":"123","taskID":123}`))
		})

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)

		im := buildTestIndexManagerWithMockServer(t, handler, cb)

		err := im.Index(context.Background(), "123", map[string]string{"id": "123", "name": "example"})
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, cb)
	})
}

func TestIndexManager_Search(T *testing.T) {
	T.Parallel()

	T.Run("with broken circuit breaker", func(t *testing.T) {
		t.Parallel()

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(true)

		im := buildTestIndexManagerWithCircuitBreaker(t, cb)

		results, err := im.Search(context.Background(), "query")
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, circuitbreaking.ErrCircuitBroken, err)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with empty query", func(t *testing.T) {
		t.Parallel()

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)

		im := buildTestIndexManagerWithCircuitBreaker(t, cb)

		results, err := im.Search(context.Background(), "")
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, ErrEmptyQueryProvided, err)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with valid query but invalid credentials", func(t *testing.T) {
		t.Parallel()

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)
		cb.On("Failed").Return()

		im := buildTestIndexManagerWithCircuitBreaker(t, cb)

		results, err := im.Search(context.Background(), "test query")
		assert.Error(t, err)
		assert.Nil(t, results)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with successful search results", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"hits":[{"objectID":"123"}],"nbHits":1,"page":0,"nbPages":1,"hitsPerPage":20,"processingTimeMS":1}`))
		})

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)
		cb.On("Succeeded").Return()

		im := buildTestIndexManagerWithMockServer(t, handler, cb)

		results, err := im.Search(context.Background(), "test query")
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results, 1)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with empty search results", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"hits":[],"nbHits":0,"page":0,"nbPages":0,"hitsPerPage":20,"processingTimeMS":1}`))
		})

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)
		cb.On("Succeeded").Return()

		im := buildTestIndexManagerWithMockServer(t, handler, cb)

		results, err := im.Search(context.Background(), "test query")
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Empty(t, results)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with multiple search results", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"hits":[{"objectID":"abc","name":"first"},{"objectID":"def","name":"second"},{"objectID":"ghi","name":"third"}],"nbHits":3,"page":0,"nbPages":1,"hitsPerPage":20,"processingTimeMS":1}`))
		})

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)
		cb.On("Succeeded").Return()

		im := buildTestIndexManagerWithMockServer(t, handler, cb)

		results, err := im.Search(context.Background(), "test query")
		assert.NoError(t, err)
		assert.Len(t, results, 3)
		assert.Equal(t, "abc", results[0].ID)
		assert.Equal(t, "first", results[0].Name)
		assert.Equal(t, "def", results[1].ID)
		assert.Equal(t, "second", results[1].Name)
		assert.Equal(t, "ghi", results[2].ID)
		assert.Equal(t, "third", results[2].Name)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("when unmarshalling search result fails", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"hits":[{"objectID":"123","name":["not","a","string"]}],"nbHits":1,"page":0,"nbPages":1,"hitsPerPage":20,"processingTimeMS":1}`))
		})

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)

		im := buildTestIndexManagerWithMockServer(t, handler, cb)

		results, err := im.Search(context.Background(), "test query")
		assert.Error(t, err)
		assert.Nil(t, results)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with successful search results without objectID", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"hits":[{"name":"example"}],"nbHits":1,"page":0,"nbPages":1,"hitsPerPage":20,"processingTimeMS":1}`))
		})

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)
		cb.On("Succeeded").Return()

		im := buildTestIndexManagerWithMockServer(t, handler, cb)

		results, err := im.Search(context.Background(), "test query")
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results, 1)

		mock.AssertExpectationsForObjects(t, cb)
	})
}

func TestIndexManager_Delete(T *testing.T) {
	T.Parallel()

	T.Run("with broken circuit breaker", func(t *testing.T) {
		t.Parallel()

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(true)

		im := buildTestIndexManagerWithCircuitBreaker(t, cb)

		err := im.Delete(context.Background(), "id")
		assert.Error(t, err)
		assert.Equal(t, circuitbreaking.ErrCircuitBroken, err)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with invalid credentials", func(t *testing.T) {
		t.Parallel()

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)
		cb.On("Failed").Return()

		im := buildTestIndexManagerWithCircuitBreaker(t, cb)

		err := im.Delete(context.Background(), "some-id")
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with successful deletion", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"deletedAt":"2021-01-01T00:00:00Z","taskID":123}`))
		})

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)
		cb.On("Succeeded").Return()

		im := buildTestIndexManagerWithMockServer(t, handler, cb)

		err := im.Delete(context.Background(), "some-id")
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, cb)
	})
}

func TestIndexManager_Wipe(T *testing.T) {
	T.Parallel()

	T.Run("with broken circuit breaker", func(t *testing.T) {
		t.Parallel()

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(true)

		im := buildTestIndexManagerWithCircuitBreaker(t, cb)

		err := im.Wipe(context.Background())
		assert.Error(t, err)
		assert.Equal(t, circuitbreaking.ErrCircuitBroken, err)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with invalid credentials", func(t *testing.T) {
		t.Parallel()

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)
		cb.On("Failed").Return()

		im := buildTestIndexManagerWithCircuitBreaker(t, cb)

		err := im.Wipe(context.Background())
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with successful wipe", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"updatedAt":"2021-01-01T00:00:00Z","taskID":123}`))
		})

		cb := &mockcircuitbreaking.MockCircuitBreaker{}
		cb.On("CannotProceed").Return(false)
		cb.On("Succeeded").Return()

		im := buildTestIndexManagerWithMockServer(t, handler, cb)

		err := im.Wipe(context.Background())
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, cb)
	})
}
