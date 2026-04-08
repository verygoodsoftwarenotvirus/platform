package algolia

import (
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v5/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	textsearch "github.com/verygoodsoftwarenotvirus/platform/v5/search/text"

	algolia "github.com/algolia/algoliasearch-client-go/v3/algolia/search"
)

var (
	_ textsearch.Index[any] = (*indexManager[any])(nil)

	ErrNilConfig = platformerrors.New("nil config provided")
)

type (
	indexManager[T any] struct {
		logger         logging.Logger
		tracer         tracing.Tracer
		circuitBreaker circuitbreaking.CircuitBreaker
		client         *algolia.Index
		DataType       *T
	}
)

func ProvideIndexManager[T any](
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	cfg *Config,
	indexName string,
	circuitBreaker circuitbreaking.CircuitBreaker,
) (textsearch.Index[T], error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	im := &indexManager[T]{
		tracer:         tracing.NewNamedTracer(tracerProvider, fmt.Sprintf("search_%s", indexName)),
		logger:         logging.NewNamedLogger(logger, indexName),
		client:         algolia.NewClient(cfg.AppID, cfg.APIKey).InitIndex(indexName),
		circuitBreaker: circuitBreaker,
	}

	return im, nil
}
