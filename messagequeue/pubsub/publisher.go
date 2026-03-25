package pubsub

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/verygoodsoftwarenotvirus/platform/v2/encoding"
	"github.com/verygoodsoftwarenotvirus/platform/v2/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"cloud.google.com/go/pubsub/v2"
)

type (
	messagePublisher interface {
		Stop()
		Publish(context.Context, *pubsub.Message) *pubsub.PublishResult
	}

	pubSubPublisher struct {
		tracer    tracing.Tracer
		encoder   encoding.ClientEncoder
		logger    logging.Logger
		publisher messagePublisher
	}
)

// buildPubSubPublisher provides a Pub/Sub-backed pubSubPublisher.
func buildPubSubPublisher(logger logging.Logger, pubsubClient *pubsub.Publisher, tracerProvider tracing.TracerProvider, topic string) *pubSubPublisher {
	return &pubSubPublisher{
		encoder:   encoding.ProvideClientEncoder(logger, tracerProvider, encoding.ContentTypeJSON),
		logger:    logging.EnsureLogger(logger),
		publisher: pubsubClient,
		tracer:    tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(fmt.Sprintf("%s_publisher", topic))),
	}
}

// Stop calls Stop on the topic.
func (p *pubSubPublisher) Stop() {
	p.publisher.Stop()
}

type publisherProvider struct {
	logger            logging.Logger
	publisherCache    map[string]messagequeue.Publisher
	pubsubClient      *pubsub.Client
	tracerProvider    tracing.TracerProvider
	projectID         string
	publisherCacheHat sync.RWMutex
}

// ProvidePubSubPublisherProvider returns a PublisherProvider for a given address.
func ProvidePubSubPublisherProvider(logger logging.Logger, tracerProvider tracing.TracerProvider, client *pubsub.Client, projectID string) messagequeue.PublisherProvider {
	return &publisherProvider{
		logger:         logging.EnsureLogger(logger),
		pubsubClient:   client,
		publisherCache: map[string]messagequeue.Publisher{},
		tracerProvider: tracerProvider,
		projectID:      projectID,
	}
}

// Ping is a no-op for GCP Pub/Sub (managed service).
func (p *publisherProvider) Ping(context.Context) error { return nil }

// Close closes the connection topic.
func (p *publisherProvider) Close() {
	if err := p.pubsubClient.Close(); err != nil {
		p.logger.Error("closing pubsub connection", err)
	}
}

// qualifyTopicName ensures the topic name is fully qualified (projects/{project}/topics/{topic}).
func (p *publisherProvider) qualifyTopicName(topicName string) string {
	if strings.HasPrefix(topicName, "projects/") {
		return topicName
	}
	return fmt.Sprintf("projects/%s/topics/%s", p.projectID, topicName)
}

// ProvidePublisher returns a pubSubPublisher for a given topic.
func (p *publisherProvider) ProvidePublisher(ctx context.Context, topicName string) (messagequeue.Publisher, error) {
	if topicName == "" {
		return nil, messagequeue.ErrEmptyTopicName
	}

	qualifiedName := p.qualifyTopicName(topicName)

	logger := logging.EnsureLogger(p.logger.Clone())

	p.publisherCacheHat.Lock()
	defer p.publisherCacheHat.Unlock()
	if cachedPub, ok := p.publisherCache[qualifiedName]; ok {
		return cachedPub, nil
	}

	// Use Publisher directly with the qualified topic name. This avoids needing
	// pubsub.topics.get (TopicAdminClient.GetTopic); pubsub.topics.publish is sufficient.
	publisher := p.pubsubClient.Publisher(qualifiedName)

	pub := buildPubSubPublisher(logger, publisher, p.tracerProvider, qualifiedName)
	p.publisherCache[qualifiedName] = pub

	return pub, nil
}

func (p *pubSubPublisher) Publish(ctx context.Context, data any) error {
	_, span := p.tracer.StartSpan(ctx)
	defer span.End()

	logger := p.logger.Clone()

	var b bytes.Buffer
	if err := p.encoder.Encode(ctx, &b, data); err != nil {
		return observability.PrepareError(err, span, "encoding topic message")
	}

	msg := &pubsub.Message{Data: b.Bytes()}
	result := p.publisher.Publish(ctx, msg)

	<-result.Ready()

	// The Get method blocks until a server-generated ID or an error is returned for the published message.
	if _, err := result.Get(ctx); err != nil {
		observability.AcknowledgeError(err, logger, span, "publishing pubsub message")
	}

	logger.Debug("published message")

	return nil
}

func (p *pubSubPublisher) PublishAsync(ctx context.Context, data any) {
	_, span := p.tracer.StartSpan(ctx)
	defer span.End()

	logger := p.logger.Clone()

	var b bytes.Buffer
	if err := p.encoder.Encode(ctx, &b, data); err != nil {
		observability.AcknowledgeError(err, logger, span, "encoding topic message")
	}

	msg := &pubsub.Message{Data: b.Bytes()}
	result := p.publisher.Publish(ctx, msg)
	<-result.Ready()

	// The Get method blocks until a server-generated ID or an error is returned for the published message.
	if _, err := result.Get(ctx); err != nil {
		observability.AcknowledgeError(err, logger, span, "publishing pubsub message")
	}

	logger.Debug("published message")
}
