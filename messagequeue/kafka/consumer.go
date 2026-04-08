package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/segmentio/kafka-go"
)

type (
	kafkaReader interface {
		FetchMessage(ctx context.Context) (kafka.Message, error)
		CommitMessages(ctx context.Context, msgs ...kafka.Message) error
		Close() error
	}

	kafkaConsumer struct {
		tracer          tracing.Tracer
		logger          logging.Logger
		consumedCounter metrics.Int64Counter
		handlerFunc     func(context.Context, []byte) error
		reader          kafkaReader
	}
)

var _ messagequeue.Consumer = (*kafkaConsumer)(nil)

func provideKafkaConsumer(logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider, brokers []string, groupID, topic string, handlerFunc func(context.Context, []byte) error) *kafkaConsumer {
	mp := metrics.EnsureMetricsProvider(metricsProvider)

	consumedCounter, err := mp.NewInt64Counter(fmt.Sprintf("%s_consumed", topic))
	if err != nil {
		panic(fmt.Sprintf("creating consumed counter: %v", err))
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
	})

	return &kafkaConsumer{
		handlerFunc:     handlerFunc,
		reader:          reader,
		logger:          logging.EnsureLogger(logger),
		tracer:          tracing.NewNamedTracer(tracerProvider, fmt.Sprintf("%s_consumer", topic)),
		consumedCounter: consumedCounter,
	}
}

// Consume reads messages from Kafka and applies the handler to their payloads.
func (c *kafkaConsumer) Consume(ctx context.Context, stopChan chan bool, errs chan error) {
	if stopChan == nil {
		stopChan = make(chan bool, 1)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-stopChan:
			return
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if errs != nil {
				errs <- err
			}
			continue
		}

		msgCtx, span := c.tracer.StartCustomSpan(ctx, "consume_message")
		c.consumedCounter.Add(msgCtx, 1)

		if err = c.handlerFunc(msgCtx, msg.Value); err != nil {
			observability.AcknowledgeError(err, c.logger, span, "handling message")
			if errs != nil {
				errs <- err
			}
		} else if err = c.reader.CommitMessages(msgCtx, msg); err != nil {
			observability.AcknowledgeError(err, c.logger, span, "committing message")
		}

		span.End()
	}
}

type consumerProvider struct {
	logger          logging.Logger
	tracerProvider  tracing.TracerProvider
	metricsProvider metrics.Provider
	consumerCache   map[string]messagequeue.Consumer
	groupID         string
	brokers         []string
	consumerCacheMu sync.RWMutex
}

var _ messagequeue.ConsumerProvider = (*consumerProvider)(nil)

// ProvideKafkaConsumerProvider returns a ConsumerProvider backed by Kafka.
func ProvideKafkaConsumerProvider(logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider, cfg Config) messagequeue.ConsumerProvider {
	logger.WithValue("brokers", cfg.Brokers).WithValue("group_id", cfg.GroupID).Info("setting up kafka consumer")

	return &consumerProvider{
		logger:          logging.EnsureLogger(logger),
		tracerProvider:  tracerProvider,
		metricsProvider: metricsProvider,
		brokers:         cfg.Brokers,
		groupID:         cfg.GroupID,
		consumerCache:   map[string]messagequeue.Consumer{},
	}
}

// ProvideConsumer returns a Consumer for the given topic.
func (p *consumerProvider) ProvideConsumer(_ context.Context, topic string, handlerFunc messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
	if topic == "" {
		return nil, ErrEmptyInputProvided
	}

	p.consumerCacheMu.Lock()
	defer p.consumerCacheMu.Unlock()
	if cached, ok := p.consumerCache[topic]; ok {
		return cached, nil
	}

	c := provideKafkaConsumer(p.logger, p.tracerProvider, p.metricsProvider, p.brokers, p.groupID, topic, handlerFunc)
	p.consumerCache[topic] = c

	return c, nil
}
