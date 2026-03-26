package mockmetrics

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/metric"
)

var _ metrics.Int64Counter = (*Int64Counter)(nil)

type Int64Counter struct {
	mock.Mock
}

func (m *Int64Counter) Add(ctx context.Context, incr int64, options ...metric.AddOption) {
	m.Called(ctx, incr, options)
}
