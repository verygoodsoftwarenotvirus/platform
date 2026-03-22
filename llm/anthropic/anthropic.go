package anthropic

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v2/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v2/llm"
	"github.com/verygoodsoftwarenotvirus/platform/v2/pointer"

	anyllm "github.com/mozilla-ai/any-llm-go"
	anyllmanthropic "github.com/mozilla-ai/any-llm-go/providers/anthropic"
)

// NewProvider creates a new Anthropic-backed LLM provider.
func NewProvider(cfg *Config) (llm.Provider, error) {
	if cfg == nil {
		return nil, errors.New("anthropic config is required")
	}

	opts := []anyllm.Option{
		anyllm.WithAPIKey(cfg.APIKey),
	}
	if cfg.BaseURL != "" {
		opts = append(opts, anyllm.WithBaseURL(cfg.BaseURL))
	}
	if cfg.Timeout > 0 {
		opts = append(opts, anyllm.WithTimeout(cfg.Timeout))
	}

	provider, err := anyllmanthropic.New(opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create anthropic provider")
	}

	return &anthropicProvider{
		provider:     provider,
		defaultModel: cfg.DefaultModel,
	}, nil
}

type anthropicProvider struct {
	provider     *anyllmanthropic.Provider
	defaultModel string
}

// Completion implements llm.Provider.
func (p *anthropicProvider) Completion(ctx context.Context, params llm.CompletionParams) (*llm.CompletionResult, error) {
	model := params.Model
	if model == "" {
		model = p.defaultModel
	}
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	anyllmParams := anyllm.CompletionParams{
		Model:    model,
		Messages: toAnyLLMMessages(pointer.ToSlice(params.Messages)),
	}

	resp, err := p.provider.Completion(ctx, anyllmParams)
	if err != nil {
		return nil, err
	}

	return toCompletionResult(resp), nil
}

func toAnyLLMMessages(msgs []*llm.Message) []anyllm.Message {
	out := make([]anyllm.Message, len(msgs))
	for i, m := range msgs {
		out[i] = anyllm.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return out
}

func toCompletionResult(resp *anyllm.ChatCompletion) *llm.CompletionResult {
	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.ContentString()
	}
	return &llm.CompletionResult{Content: content}
}
