package memory

import (
	"context"
	"sync"

	"github.com/verygoodsoftwarenotvirus/platform/v2/cache"
)

type inMemoryCacheImpl[T any] struct {
	cache   map[string]*T
	cacheMu sync.RWMutex
}

// NewInMemoryCache builds an in-memory cache.
func NewInMemoryCache[T any]() cache.Cache[T] {
	return &inMemoryCacheImpl[T]{
		cache: make(map[string]*T),
	}
}

func (i *inMemoryCacheImpl[T]) Get(_ context.Context, key string) (*T, error) {
	i.cacheMu.RLock()
	defer i.cacheMu.RUnlock()

	if val, ok := i.cache[key]; ok {
		return val, nil
	}

	return nil, cache.ErrNotFound
}

func (i *inMemoryCacheImpl[T]) Set(_ context.Context, key string, value *T) error {
	i.cacheMu.Lock()
	defer i.cacheMu.Unlock()

	i.cache[key] = value

	return nil
}

func (i *inMemoryCacheImpl[T]) Delete(_ context.Context, key string) error {
	i.cacheMu.Lock()
	defer i.cacheMu.Unlock()

	delete(i.cache, key)

	return nil
}
