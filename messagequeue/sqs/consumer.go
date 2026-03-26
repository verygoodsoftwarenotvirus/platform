package sqs

import (
	"context"
	"fmt"
	"sync"

	"github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	longPollWaitSeconds = 20
	maxNumberOfMessages = 10
)

type (
	messageReceiver interface {
		ReceiveMessage(ctx context.Context, input *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
		DeleteMessage(ctx context.Context, input *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	}

	sqsConsumer struct {
		tracer          tracing.Tracer
		logger          logging.Logger
		consumedCounter metrics.Int64Counter
		receiver        messageReceiver
		handlerFunc     func(context.Context, []byte) error
		queueURL        string
	}
)

func provideSQSConsumer(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	receiver messageReceiver,
	queueURL string,
	handlerFunc func(context.Context, []byte) error,
) *sqsConsumer {
	mp := metrics.EnsureMetricsProvider(metricsProvider)

	consumedCounter, err := mp.NewInt64Counter(fmt.Sprintf("%s_consumed", queueURL))
	if err != nil {
		panic(fmt.Sprintf("creating consumed counter: %v", err))
	}

	return &sqsConsumer{
		logger:          logging.EnsureLogger(logger),
		receiver:        receiver,
		queueURL:        queueURL,
		handlerFunc:     handlerFunc,
		tracer:          tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(fmt.Sprintf("%s_consumer", queueURL))),
		consumedCounter: consumedCounter,
	}
}

// Consume polls the SQS queue and processes messages until stopChan is signaled.
// On handler success, the message is deleted from the queue.
// On handler failure, the message is not deleted (it returns after visibility timeout).
func (c *sqsConsumer) Consume(ctx context.Context, stopChan chan bool, errs chan error) {
	if stopChan == nil {
		stopChan = make(chan bool, 1)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-stopChan
		cancel()
	}()

	for ctx.Err() == nil {
		output, err := c.receiver.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: maxNumberOfMessages,
			WaitTimeSeconds:     longPollWaitSeconds,
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("receiving SQS messages", err)
			if errs != nil {
				errs <- err
			}
			continue
		}

		for i := range output.Messages {
			msg := &output.Messages[i]
			if msg.Body == nil {
				continue
			}
			body := []byte(aws.ToString(msg.Body))

			msgCtx, span := c.tracer.StartCustomSpan(ctx, "consume_message")
			c.consumedCounter.Add(msgCtx, 1)
			if err = c.handlerFunc(msgCtx, body); err != nil {
				observability.AcknowledgeError(err, c.logger, span, "handling SQS message")
				if errs != nil {
					errs <- err
				}
				span.End()
				continue
			}

			if _, err = c.receiver.DeleteMessage(msgCtx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(c.queueURL),
				ReceiptHandle: msg.ReceiptHandle,
			}); err != nil {
				observability.AcknowledgeError(err, c.logger, span, "deleting SQS message")
				if errs != nil {
					errs <- err
				}
			}
			span.End()
		}
	}
}

type consumerProvider struct {
	logger          logging.Logger
	tracerProvider  tracing.TracerProvider
	metricsProvider metrics.Provider
	consumerCache   map[string]messagequeue.Consumer
	sqsClient       messageReceiver
	consumerCacheMu sync.RWMutex
}

// ProvideSQSConsumerProvider returns a ConsumerProvider for SQS.
func ProvideSQSConsumerProvider(ctx context.Context, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider, _ Config) messagequeue.ConsumerProvider {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("sqs consumer provider: load default config: " + err.Error())
	}
	svc := sqs.NewFromConfig(cfg)

	return &consumerProvider{
		logger:          logging.EnsureLogger(logger),
		tracerProvider:  tracerProvider,
		metricsProvider: metricsProvider,
		sqsClient:       svc,
		consumerCache:   map[string]messagequeue.Consumer{},
	}
}

// ProvideConsumer returns a Consumer for the given topic (queue URL).
func (p *consumerProvider) ProvideConsumer(_ context.Context, topic string, handlerFunc messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
	if topic == "" {
		return nil, messagequeue.ErrEmptyTopicName
	}

	p.consumerCacheMu.Lock()
	defer p.consumerCacheMu.Unlock()
	if cached, ok := p.consumerCache[topic]; ok {
		return cached, nil
	}

	c := provideSQSConsumer(p.logger.WithValue("queue_url", topic), p.tracerProvider, p.metricsProvider, p.sqsClient, topic, handlerFunc)
	p.consumerCache[topic] = c

	return c, nil
}
