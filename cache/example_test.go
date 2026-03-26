package cache_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v3/cache"
	"github.com/verygoodsoftwarenotvirus/platform/v3/cache/memory"
)

func ExampleCache_setAndGet() {
	ctx := context.Background()
	c := memory.NewInMemoryCache[string]()

	value := "cached-value"
	if err := c.Set(ctx, "my-key", &value); err != nil {
		panic(err)
	}

	result, err := c.Get(ctx, "my-key")
	if err != nil {
		panic(err)
	}

	fmt.Println(*result)
	// Output: cached-value
}

func ExampleCache_notFound() {
	ctx := context.Background()
	c := memory.NewInMemoryCache[string]()

	_, err := c.Get(ctx, "nonexistent")
	fmt.Println(err)
	fmt.Println(errors.Is(err, cache.ErrNotFound))
	// Output:
	// not found
	// true
}
