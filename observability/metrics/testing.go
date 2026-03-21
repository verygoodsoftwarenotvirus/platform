package metrics

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func Int64CounterForTest(name string) metric.Int64Counter {
	x, err := otel.Meter("testing").Int64Counter(name)
	if err != nil {
		panic(err)
	}

	return x
}
