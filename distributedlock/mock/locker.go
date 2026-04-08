package mock

import (
	"context"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/distributedlock"

	"github.com/stretchr/testify/mock"
)

var (
	_ distributedlock.Locker = (*Locker)(nil)
	_ distributedlock.Lock   = (*Lock)(nil)
)

// Locker is a testify-backed mock of distributedlock.Locker.
type Locker struct {
	mock.Mock
}

// Acquire implements distributedlock.Locker.
func (m *Locker) Acquire(ctx context.Context, key string, ttl time.Duration) (distributedlock.Lock, error) {
	args := m.Called(ctx, key, ttl)
	if v := args.Get(0); v != nil {
		return v.(distributedlock.Lock), args.Error(1)
	}
	return nil, args.Error(1)
}

// Ping implements distributedlock.Locker.
func (m *Locker) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// Close implements distributedlock.Locker.
func (m *Locker) Close() error {
	return m.Called().Error(0)
}

// Lock is a testify-backed mock of distributedlock.Lock.
type Lock struct {
	mock.Mock
}

// Key implements distributedlock.Lock.
func (m *Lock) Key() string {
	return m.Called().String(0)
}

// TTL implements distributedlock.Lock.
func (m *Lock) TTL() time.Duration {
	args := m.Called()
	if v, ok := args.Get(0).(time.Duration); ok {
		return v
	}
	return 0
}

// Release implements distributedlock.Lock.
func (m *Lock) Release(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// Refresh implements distributedlock.Lock.
func (m *Lock) Refresh(ctx context.Context, ttl time.Duration) error {
	return m.Called(ctx, ttl).Error(0)
}
