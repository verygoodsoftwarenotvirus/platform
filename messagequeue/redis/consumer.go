package redis

import (
	"context"
	"sync"

	"github.com/verygoodsoftwarenotvirus/platform/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"

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
		logger       logging.Logger
		handlerFunc  func(context.Context, []byte) error
		subscription channelProvider
	}
)

func provideRedisConsumer(ctx context.Context, logger logging.Logger, redisClient subscriptionProvider, topic string, handlerFunc func(context.Context, []byte) error) *redisConsumer {
	subscription := redisClient.Subscribe(ctx, topic)

	logger.Debug("subscribed to topic!")

	return &redisConsumer{
		handlerFunc:  handlerFunc,
		subscription: subscription,
		logger:       logging.EnsureLogger(logger),
	}
}

// Consume reads messages and applies the handler to their payloads.
// Writes errors to the error chan if it isn't nil.
func (r *redisConsumer) Consume(stopChan chan bool, errs chan error) {
	if stopChan == nil {
		stopChan = make(chan bool, 1)
	}
	subChan := r.subscription.Channel()

	for {
		select {
		case msg := <-subChan:
			if err := r.handlerFunc(context.Background(), []byte(msg.Payload)); err != nil {
				r.logger.Error("handling message", err)
				if errs != nil {
					errs <- err
				}
			}
		case <-stopChan:
			return
		}
	}
}

type consumerProvider struct {
	logger           logging.Logger
	consumerCache    map[string]messagequeue.Consumer
	redisClient      subscriptionProvider
	consumerCacheHat sync.RWMutex
}

// ProvideRedisConsumerProvider returns a ConsumerProvider for a given address.
func ProvideRedisConsumerProvider(logger logging.Logger, cfg Config) messagequeue.ConsumerProvider {
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
		logger:        logging.EnsureLogger(logger),
		redisClient:   redisClient,
		consumerCache: map[string]messagequeue.Consumer{},
	}
}

// ProvideConsumer returns a Consumer for a given topic.
func (p *consumerProvider) ProvideConsumer(ctx context.Context, topic string, handlerFunc messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
	logger := logging.EnsureLogger(p.logger).WithValue("topic", topic)

	if topic == "" {
		return nil, ErrEmptyInputProvided
	}

	p.consumerCacheHat.Lock()
	defer p.consumerCacheHat.Unlock()
	if cachedPub, ok := p.consumerCache[topic]; ok {
		return cachedPub, nil
	}

	c := provideRedisConsumer(ctx, logger, p.redisClient, topic, handlerFunc)
	p.consumerCache[topic] = c

	return c, nil
}
