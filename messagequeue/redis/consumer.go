package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/verygoodsoftwarenotvirus/platform/v3/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/keys"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/go-redis/redis/v8"
)

type (
	subscriptionProvider interface {
		Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	}

	channelProvider interface {
		Channel(...redis.ChannelOption) <-chan *redis.Message
	}

	redisConsumer struct {
		tracer       tracing.Tracer
		logger       logging.Logger
		handlerFunc  func(context.Context, []byte) error
		subscription channelProvider
	}
)

func provideRedisConsumer(ctx context.Context, logger logging.Logger, tracerProvider tracing.TracerProvider, redisClient subscriptionProvider, topic string, handlerFunc func(context.Context, []byte) error) *redisConsumer {
	subscription := redisClient.Subscribe(ctx, topic)

	logger.Debug("subscribed to topic!")

	return &redisConsumer{
		handlerFunc:  handlerFunc,
		subscription: subscription,
		logger:       logging.EnsureLogger(logger),
		tracer:       tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(fmt.Sprintf("%s_consumer", topic))),
	}
}

// Consume reads messages and applies the handler to their payloads.
// Writes errors to the error chan if it isn't nil.
func (r *redisConsumer) Consume(ctx context.Context, stopChan chan bool, errs chan error) {
	if stopChan == nil {
		stopChan = make(chan bool, 1)
	}
	subChan := r.subscription.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-subChan:
			msgCtx, span := r.tracer.StartCustomSpan(ctx, "consume_message")
			if err := r.handlerFunc(msgCtx, []byte(msg.Payload)); err != nil {
				observability.AcknowledgeError(err, r.logger, span, "handling message")
				if errs != nil {
					errs <- err
				}
			}
			span.End()
		case <-stopChan:
			return
		}
	}
}

type consumerProvider struct {
	logger          logging.Logger
	tracerProvider  tracing.TracerProvider
	consumerCache   map[string]messagequeue.Consumer
	redisClient     subscriptionProvider
	consumerCacheMu sync.RWMutex
}

// ProvideRedisConsumerProvider returns a ConsumerProvider for a given address.
func ProvideRedisConsumerProvider(logger logging.Logger, tracerProvider tracing.TracerProvider, cfg Config) messagequeue.ConsumerProvider {
	logger.WithValue("queue_addresses", cfg.QueueAddresses).
		WithValue(keys.UsernameKey, cfg.Username).
		WithValue("password", cfg.Password).Info("setting up redis consumer")

	var redisClient subscriptionProvider
	if len(cfg.QueueAddresses) > 1 {
		redisClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    cfg.QueueAddresses,
			Username: cfg.Username,
			Password: cfg.Password,
		})
	} else if len(cfg.QueueAddresses) == 1 {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.QueueAddresses[0],
			Username: cfg.Username,
			Password: cfg.Password,
		})
	}

	return &consumerProvider{
		logger:         logging.EnsureLogger(logger),
		tracerProvider: tracerProvider,
		redisClient:    redisClient,
		consumerCache:  map[string]messagequeue.Consumer{},
	}
}

// ProvideConsumer returns a Consumer for a given topic.
func (p *consumerProvider) ProvideConsumer(ctx context.Context, topic string, handlerFunc messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
	logger := logging.EnsureLogger(p.logger).WithValue("topic", topic)

	if topic == "" {
		return nil, ErrEmptyInputProvided
	}

	p.consumerCacheMu.Lock()
	defer p.consumerCacheMu.Unlock()
	if cachedPub, ok := p.consumerCache[topic]; ok {
		return cachedPub, nil
	}

	c := provideRedisConsumer(ctx, logger, p.tracerProvider, p.redisClient, topic, handlerFunc)
	p.consumerCache[topic] = c

	return c, nil
}
