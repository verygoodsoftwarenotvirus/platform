package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/llm"

	"github.com/stretchr/testify/mock"
)

var _ llm.Provider = (*Provider)(nil)

// Provider is a mock LLM provider for tests.
type Provider struct {
	mock.Mock
}

// Completion satisfies the llm.Provider interface.
func (m *Provider) Completion(ctx context.Context, params llm.CompletionParams) (*llm.CompletionResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*llm.CompletionResult), args.Error(1)
}
