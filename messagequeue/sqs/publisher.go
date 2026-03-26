package sqs

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/verygoodsoftwarenotvirus/platform/v3/encoding"
	"github.com/verygoodsoftwarenotvirus/platform/v3/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type (
	messagePublisher interface {
		SendMessage(ctx context.Context, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	}

	sqsPublisher struct {
		tracer    tracing.Tracer
		encoder   encoding.ClientEncoder
		logger    logging.Logger
		publisher messagePublisher
		topic     string
	}
)

// Stop does nothing.
func (p *sqsPublisher) Stop() {}

// Publish publishes a message onto an SQS event queue.
func (p *sqsPublisher) Publish(ctx context.Context, data any) error {
	_, span := p.tracer.StartSpan(ctx)
	defer span.End()

	logger := p.logger

	logger.Debug("publishing message")

	var b bytes.Buffer
	if err := p.encoder.Encode(ctx, &b, data); err != nil {
		return observability.PrepareError(err, span, "encoding topic message")
	}

	input := &sqs.SendMessageInput{
		MessageBody: aws.String(b.String()),
		QueueUrl:    aws.String(p.topic),
	}

	if _, err := p.publisher.SendMessage(ctx, input); err != nil {
		return observability.PrepareError(err, span, "publishing message")
	}

	return nil
}

// PublishAsync publishes a message onto an SQS event queue.
func (p *sqsPublisher) PublishAsync(ctx context.Context, data any) {
	ctx, span := p.tracer.StartSpan(ctx)
	defer span.End()

	logger := p.logger

	logger.Debug("publishing message")

	var b bytes.Buffer
	if err := p.encoder.Encode(ctx, &b, data); err != nil {
		observability.AcknowledgeError(err, logger, span, "encoding topic message")
		return
	}

	input := &sqs.SendMessageInput{
		MessageBody: aws.String(b.String()),
		QueueUrl:    aws.String(p.topic),
	}

	if _, err := p.publisher.SendMessage(ctx, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "publishing message")
	}
}

// provideSQSPublisher provides a sqs-backed Publisher.
func provideSQSPublisher(logger logging.Logger, sqsClient messagePublisher, tracerProvider tracing.TracerProvider, topic string) *sqsPublisher {
	return &sqsPublisher{
		publisher: sqsClient,
		topic:     topic,
		encoder:   encoding.ProvideClientEncoder(logger, tracerProvider, encoding.ContentTypeJSON),
		logger:    logging.EnsureLogger(logger),
		tracer:    tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(fmt.Sprintf("%s_publisher", topic))),
	}
}

type publisherProvider struct {
	logger            logging.Logger
	publisherCache    map[string]messagequeue.Publisher
	sqsClient         messagePublisher
	tracerProvider    tracing.TracerProvider
	publisherCacheHat sync.RWMutex
}

// ProvideSQSPublisherProvider returns a PublisherProvider for a given address.
func ProvideSQSPublisherProvider(ctx context.Context, logger logging.Logger, tracerProvider tracing.TracerProvider) messagequeue.PublisherProvider {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("sqs publisher provider: load default config: " + err.Error())
	}
	svc := sqs.NewFromConfig(cfg)

	return &publisherProvider{
		logger:         logging.EnsureLogger(logger),
		sqsClient:      svc,
		publisherCache: map[string]messagequeue.Publisher{},
		tracerProvider: tracerProvider,
	}
}

// ProvidePublisher returns a Publisher for a given topic.
func (p *publisherProvider) ProvidePublisher(ctx context.Context, topic string) (messagequeue.Publisher, error) {
	if topic == "" {
		return nil, messagequeue.ErrEmptyTopicName
	}
	logger := logging.EnsureLogger(p.logger).WithValue("topic", topic)

	p.publisherCacheHat.Lock()
	defer p.publisherCacheHat.Unlock()
	if cachedPub, ok := p.publisherCache[topic]; ok {
		return cachedPub, nil
	}

	pub := provideSQSPublisher(logger, p.sqsClient, p.tracerProvider, topic)
	p.publisherCache[topic] = pub

	return pub, nil
}

// Ping is a no-op for SQS (SQS is a managed service).
func (p *publisherProvider) Ping(context.Context) error { return nil }

// Close returns a Publisher for a given topic.
func (p *publisherProvider) Close() {}
