package textsearch

import (
	"context"
)

var _ Index[any] = (*NoopIndexManager[any])(nil)

// NoopIndexManager is a noop Index.
type NoopIndexManager[T any] struct{}

// Search is a no-op method.
func (*NoopIndexManager[T]) Search(context.Context, string) ([]*T, error) {
	return []*T{}, nil
}

// Index is a no-op method.
func (*NoopIndexManager[T]) Index(context.Context, string, any) error {
	return nil
}

// Delete is a no-op method.
func (*NoopIndexManager[T]) Delete(context.Context, string) error {
	return nil
}

// Wipe is a no-op method.
func (*NoopIndexManager[T]) Wipe(context.Context) error {
	return nil
}
