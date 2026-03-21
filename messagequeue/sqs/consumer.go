package sqs

import (
	"context"
	"sync"

	"github.com/verygoodsoftwarenotvirus/platform/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"

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
		logger      logging.Logger
		receiver    messageReceiver
		handlerFunc func(context.Context, []byte) error
		queueURL    string
	}
)

func provideSQSConsumer(
	logger logging.Logger,
	receiver messageReceiver,
	queueURL string,
	handlerFunc func(context.Context, []byte) error,
) *sqsConsumer {
	return &sqsConsumer{
		logger:      logging.EnsureLogger(logger),
		receiver:    receiver,
		queueURL:    queueURL,
		handlerFunc: handlerFunc,
	}
}

// Consume polls the SQS queue and processes messages until stopChan is signaled.
// On handler success, the message is deleted from the queue.
// On handler failure, the message is not deleted (it returns after visibility timeout).
func (c *sqsConsumer) Consume(stopChan chan bool, errs chan error) {
	if stopChan == nil {
		stopChan = make(chan bool, 1)
	}

	ctx, cancel := context.WithCancel(context.Background())
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
			if err = c.handlerFunc(ctx, body); err != nil {
				c.logger.Error("handling SQS message", err)
				if errs != nil {
					errs <- err
				}
				continue
			}

			if _, err = c.receiver.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(c.queueURL),
				ReceiptHandle: msg.ReceiptHandle,
			}); err != nil {
				c.logger.Error("deleting SQS message", err)
				if errs != nil {
					errs <- err
				}
			}
		}
	}
}

type consumerProvider struct {
	logger           logging.Logger
	consumerCache    map[string]messagequeue.Consumer
	sqsClient        messageReceiver
	consumerCacheHat sync.RWMutex
}

// ProvideSQSConsumerProvider returns a ConsumerProvider for SQS.
func ProvideSQSConsumerProvider(ctx context.Context, logger logging.Logger, _ Config) messagequeue.ConsumerProvider {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("sqs consumer provider: load default config: " + err.Error())
	}
	svc := sqs.NewFromConfig(cfg)

	return &consumerProvider{
		logger:        logging.EnsureLogger(logger),
		sqsClient:     svc,
		consumerCache: map[string]messagequeue.Consumer{},
	}
}

// ProvideConsumer returns a Consumer for the given topic (queue URL).
func (p *consumerProvider) ProvideConsumer(_ context.Context, topic string, handlerFunc messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
	if topic == "" {
		return nil, messagequeue.ErrEmptyTopicName
	}

	p.consumerCacheHat.Lock()
	defer p.consumerCacheHat.Unlock()
	if cached, ok := p.consumerCache[topic]; ok {
		return cached, nil
	}

	c := provideSQSConsumer(p.logger.WithValue("queue_url", topic), p.sqsClient, topic, handlerFunc)
	p.consumerCache[topic] = c

	return c, nil
}
