package ratelimiting_test

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v3/ratelimiting"
)

func ExampleNewInMemoryRateLimiter() {
	limiter := ratelimiting.NewInMemoryRateLimiter(10.0, 5)

	allowed, err := limiter.Allow(context.Background(), "user-123")
	if err != nil {
		panic(err)
	}

	fmt.Println(allowed)
	// Output: true
}
