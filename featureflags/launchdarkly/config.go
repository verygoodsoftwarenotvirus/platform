package launchdarkly

import (
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v2/circuitbreaking"
)

type (
	Config struct {
		SDKKey               string                 `env:"SDK_KEY"                 json:"sdkKey"`
		CircuitBreakerConfig circuitbreaking.Config `envPrefix:"CIRCUIT_BREAKING_" json:"circuitBreakerConfig"`
		InitTimeout          time.Duration          `env:"INIT_TIMEOUT"            json:"initTimeout"`
	}
)
