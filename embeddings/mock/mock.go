package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v5/embeddings"

	"github.com/stretchr/testify/mock"
)

var _ embeddings.Embedder = (*Embedder)(nil)

// Embedder is a mock embeddings.Embedder for use in tests.
type Embedder struct {
	mock.Mock
}

// GenerateEmbedding satisfies the embeddings.Embedder interface.
func (m *Embedder) GenerateEmbedding(ctx context.Context, input *embeddings.Input) (*embeddings.Embedding, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*embeddings.Embedding), args.Error(1)
}
