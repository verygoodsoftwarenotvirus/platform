package mock

import (
	"context"

	vectorsearch "github.com/verygoodsoftwarenotvirus/platform/v5/search/vector"

	"github.com/stretchr/testify/mock"
)

var _ vectorsearch.Index[any] = (*Index[any])(nil)

// Index is a testify-backed mock of vectorsearch.Index.
type Index[T any] struct {
	mock.Mock
}

// Upsert implements the vectorsearch.Index interface.
func (m *Index[T]) Upsert(ctx context.Context, vectors ...vectorsearch.Vector[T]) error {
	return m.Called(ctx, vectors).Error(0)
}

// Delete implements the vectorsearch.Index interface.
func (m *Index[T]) Delete(ctx context.Context, ids ...string) error {
	return m.Called(ctx, ids).Error(0)
}

// Wipe implements the vectorsearch.Index interface.
func (m *Index[T]) Wipe(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// Query implements the vectorsearch.Index interface.
func (m *Index[T]) Query(ctx context.Context, req vectorsearch.QueryRequest) ([]vectorsearch.QueryResult[T], error) {
	args := m.Called(ctx, req)
	if v := args.Get(0); v != nil {
		return v.([]vectorsearch.QueryResult[T]), args.Error(1)
	}
	return nil, args.Error(1)
}
