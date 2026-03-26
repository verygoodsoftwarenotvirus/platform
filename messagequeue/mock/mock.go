package mockpublishers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/messagequeue"

	"github.com/stretchr/testify/mock"
)

var _ messagequeue.Publisher = (*Publisher)(nil)

// Publisher implements our interface.
type Publisher struct {
	mock.Mock
}

// Publish implements our interface.
func (m *Publisher) Publish(ctx context.Context, data any) error {
	return m.Called(ctx, data).Error(0)
}

// PublishAsync implements our interface.
func (m *Publisher) PublishAsync(ctx context.Context, data any) {
	m.Called(ctx, data)
}

// Stop implements our interface.
func (m *Publisher) Stop() {
	m.Called()
}

// PublisherProvider implements our interface.
type PublisherProvider struct {
	mock.Mock
}

// ProvidePublisher implements our interface.
func (m *PublisherProvider) ProvidePublisher(ctx context.Context, topic string) (messagequeue.Publisher, error) {
	args := m.Called(topic)

	return args.Get(0).(messagequeue.Publisher), args.Error(1)
}

// Close implements our interface.
func (m *PublisherProvider) Close() {
	m.Called()
}

// Ping implements our interface.
func (m *PublisherProvider) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
