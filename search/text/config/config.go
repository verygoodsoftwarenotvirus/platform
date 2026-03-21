package textsearchcfg

import (
	"context"
	"fmt"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"
	textsearch "github.com/verygoodsoftwarenotvirus/platform/search/text"
	"github.com/verygoodsoftwarenotvirus/platform/search/text/algolia"
	"github.com/verygoodsoftwarenotvirus/platform/search/text/elasticsearch"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ElasticsearchProvider represents the elasticsearch search index provider.
	ElasticsearchProvider = "elasticsearch"
	// AlgoliaProvider represents the algolia search index provider.
	AlgoliaProvider = "algolia"
)

// Config contains settings regarding search indices.
type Config struct {
	_              struct{}               `json:"-"`
	Algolia        *algolia.Config        `env:"init"     envPrefix:"ALGOLIA_"         json:"algolia"`
	Elasticsearch  *elasticsearch.Config  `env:"init"     envPrefix:"ELASTICSEARCH_"   json:"elasticsearch"`
	Provider       string                 `env:"PROVIDER" json:"provider"`
	CircuitBreaker circuitbreaking.Config `env:"init"     envPrefix:"CIRCUIT_BREAKER_" json:"circuitBreakerConfig"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Provider, validation.In(ElasticsearchProvider, AlgoliaProvider)),
		validation.Field(&cfg.Algolia, validation.When(cfg.Provider == AlgoliaProvider, validation.Required)),
		validation.Field(&cfg.Elasticsearch, validation.When(cfg.Provider == ElasticsearchProvider, validation.Required)),
	)
}

// ProvideIndex validates a Config struct.
func ProvideIndex[T any](
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *Config,
	indexName string,
) (textsearch.Index[T], error) {
	circuitBreaker, err := circuitbreaking.ProvideCircuitBreaker(ctx, &cfg.CircuitBreaker, logger, metricsProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize text search circuit breaker: %w", err)
	}

	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case ElasticsearchProvider:
		return elasticsearch.ProvideIndexManager[T](ctx, logger, tracerProvider, cfg.Elasticsearch, indexName, circuitBreaker)
	case AlgoliaProvider:
		return algolia.ProvideIndexManager[T](logger, tracerProvider, cfg.Algolia, indexName, circuitBreaker)
	default:
		return &textsearch.NoopIndexManager[T]{}, nil
	}
}
