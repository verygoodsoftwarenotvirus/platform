package algolia

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking"
	mockcircuitbreaking "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/mock"
	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
}
