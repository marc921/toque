package messagebroker

import (
	"context"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sethvargo/go-retry"
	"go.uber.org/zap"
)

const (
	MaxRetries    = uint64(60)
	RetryInterval = 5 * time.Second
)

type Publisher interface {
	Publish(context.Context, string) error
	Close() error
}

type PublisherMock struct{}

func (p *PublisherMock) Publish(ctx context.Context, body string) error {
	return nil
}

func (p *PublisherMock) Close() error {
	return nil
}

type Consumer interface {
	Consume(ctx context.Context) (amqp.Delivery, error)
	Close() error
}

type RabbitMQClient struct {
	logger    *zap.Logger
	conn      *amqp.Connection
	ch        *amqp.Channel
	q         amqp.Queue
	connClose chan *amqp.Error
	url       string
}

func NewRabbitMQClient(ctx context.Context, logger *zap.Logger, url, queueName string) (*RabbitMQClient, error) {
	attempt := 0
	conn, err := retry.DoValue(
		ctx,
		retry.WithMaxRetries(MaxRetries, retry.NewConstant(RetryInterval)),
		func(ctx context.Context) (*amqp.Connection, error) {
			conn, err := amqp.Dial(url)
			attempt++
			if err != nil {
				logger.Info(
					"failed to connect to RabbitMQ, retrying...",
					zap.Int("attempt", attempt),
					zap.Uint64("maxRetries", MaxRetries),
					zap.Duration("retryInterval", RetryInterval),
				)
				return nil, retry.RetryableError(err)
			}
			logger.Info(
				"successfully connected to RabbitMQ",
				zap.Int("attempt", attempt),
				zap.Uint64("maxRetries", MaxRetries),
			)
			return conn, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after multiple attempts: %w", err)
	}

	connClose := conn.NotifyClose(make(chan *amqp.Error))

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	queue, err := channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return &RabbitMQClient{
		logger:    logger,
		conn:      conn,
		ch:        channel,
		q:         queue,
		connClose: connClose,
		url:       url,
	}, nil
}

func (r *RabbitMQClient) register(consumerName string) (<-chan amqp.Delivery, error) {
	msgs, err := r.ch.Consume(
		r.q.Name,     // queue
		consumerName, // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)

	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %w", err)
	}
	return msgs, nil
}

// Publish sends a message to the RabbitMQ server
func (r *RabbitMQClient) Publish(ctx context.Context, body []byte) error {
	err := r.ch.PublishWithContext(
		ctx,
		"",       // exchange
		r.q.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}
	return nil
}

func (r *RabbitMQClient) Close() error {
	if err := r.ch.Close(); err != nil {
		return fmt.Errorf("failed to close the channel: %w", err)
	}
	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("failed to close the connection: %w", err)
	}
	return nil
}

type RabbitMQConsumer struct {
	*RabbitMQClient
	consumerName string
	messages     <-chan amqp.Delivery
}

func NewRabbitMQConsumer(ctx context.Context, logger *zap.Logger, url, queueName, consumerName string) (*RabbitMQConsumer, error) {
	client, err := NewRabbitMQClient(ctx, logger, queueName, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ client: %v", err)
	}
	messages, err := client.register(consumerName)
	if err != nil {
		return nil, fmt.Errorf("client.register: %w", err)
	}
	return &RabbitMQConsumer{
		RabbitMQClient: client,
		messages:       messages,
	}, nil
}

func (c *RabbitMQConsumer) reconnect(ctx context.Context, connErr *amqp.Error) (bool, error) {
	if connErr == nil {
		// Connection was closed intentionally
		c.logger.Info("connection closed gracefully")
		return false, nil
	}

	c.logger.Error("connection closed with error", zap.Error(connErr))
	client, err := NewRabbitMQClient(ctx, c.logger, c.q.Name, c.url)
	if err != nil {
		return false, fmt.Errorf("NewRabbitMQClient: %w", err)
	}
	c.conn = client.conn
	c.ch = client.ch
	c.q = client.q
	c.connClose = client.connClose
	ch, err := c.register(c.consumerName)
	if err != nil {
		return false, fmt.Errorf("register: %w", err)
	}
	c.messages = ch
	return true, nil
}

func (c *RabbitMQConsumer) read(ctx context.Context) (amqp.Delivery, error) {
	select {
	case <-ctx.Done():
		return amqp.Delivery{}, ctx.Err()
	case connErr := <-c.connClose:
		ok, err := c.reconnect(ctx, connErr)
		if err != nil {
			return amqp.Delivery{}, fmt.Errorf("reconnect: %w", err)
		}
		if !ok {
			return amqp.Delivery{}, fmt.Errorf("connection closed")
		}
		return c.read(ctx)
	case msg, ok := <-c.messages:
		if !ok {
			return amqp.Delivery{}, fmt.Errorf("channel closed")
		}
		return msg, nil
	}
}

type MessageProcessFunc func(msg amqp.Delivery) error

var ErrUnprocessableMessage = errors.New("unprocessable message")

func (c *RabbitMQConsumer) Consume(ctx context.Context, process MessageProcessFunc) error {
	for {
		msg, err := c.read(ctx)
		if err != nil {
			return fmt.Errorf("failed to consume a message: %w", err)
		}
		if len(msg.Body) == 0 {
			c.logger.Warn("received an empty message, skipping")
			continue
		}

		processErr := process(msg)
		if processErr != nil {
			c.logger.Error(
				"failed to process message",
				zap.Bool("unprocessable", errors.Is(processErr, ErrUnprocessableMessage)),
				zap.Error(err),
			)
		}

		// Acknowledge processing of the message, whether it was successful or not.
		attempt := 0
		err = retry.Do(
			ctx,
			retry.WithMaxRetries(MaxRetries, retry.NewConstant(RetryInterval)),
			func(ctx context.Context) error {
				var err error
				switch processErr {
				case nil:
					// Message processed successfully, acknowledge it.
					err = msg.Ack(false)
				case ErrUnprocessableMessage:
					// Message is unprocessable, acknowledge it and log a warning.
					err = msg.Nack(false, false)
				default:
					// Message processing failed, log an error and retry.
					err = msg.Nack(false, true)
				}
				attempt++
				if err != nil {
					c.logger.Info(
						"failed to acknowledge message, retrying...",
						zap.Int("attempt", attempt),
						zap.Uint64("maxRetries", MaxRetries),
						zap.Duration("retryInterval", RetryInterval),
					)
					return retry.RetryableError(err)
				}
				return nil
			},
		)
		if err != nil {
			// Log the error and continue to the next message.
			// The message will be redelivered by the broker, which is bad but should be rare enough to be invisible
			// in the monitoring, billing and quotas.
			c.logger.Error("failed to acknowledge message after multiple attempts", zap.Error(err))
		}
	}
}
