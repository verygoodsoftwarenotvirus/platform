package circuitbreaking_test

import (
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v3/circuitbreaking"
)

func ExampleNewNoopCircuitBreaker() {
	cb := circuitbreaking.NewNoopCircuitBreaker()

	fmt.Println(cb.CanProceed())

	cb.Failed()
	fmt.Println(cb.CanProceed())
	// Output:
	// true
	// true
}
