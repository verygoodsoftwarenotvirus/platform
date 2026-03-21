package memory

import (
	"context"
	"sync"

	"github.com/verygoodsoftwarenotvirus/platform/cache"
)

type inMemoryCacheImpl[T any] struct {
	cache    map[string]*T
	cacheHat sync.RWMutex
}

// NewInMemoryCache builds an in-memory cache.
func NewInMemoryCache[T any]() cache.Cache[T] {
	return &inMemoryCacheImpl[T]{
		cache: make(map[string]*T),
	}
}

func (i *inMemoryCacheImpl[T]) Get(_ context.Context, key string) (*T, error) {
	i.cacheHat.RLock()
	defer i.cacheHat.RUnlock()

	if val, ok := i.cache[key]; ok {
		return val, nil
	}

	return nil, cache.ErrNotFound
}

func (i *inMemoryCacheImpl[T]) Set(_ context.Context, key string, value *T) error {
	i.cacheHat.Lock()
	defer i.cacheHat.Unlock()

	i.cache[key] = value

	return nil
}

func (i *inMemoryCacheImpl[T]) Delete(_ context.Context, key string) error {
	i.cacheHat.Lock()
	defer i.cacheHat.Unlock()

	delete(i.cache, key)

	return nil
}
