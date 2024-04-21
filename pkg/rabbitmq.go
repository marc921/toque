package pkg

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
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
	Consume() (<-chan amqp.Delivery, error)
	Close() error
}

type RabbitMQClient struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

func NewRabbitMQClient() (*RabbitMQClient, error) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq-service:5672/")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ server: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	queue, err := channel.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %v", err)
	}

	return &RabbitMQClient{
		conn: conn,
		ch:   channel,
		q:    queue,
	}, nil
}

func (r *RabbitMQClient) Publish(ctx context.Context, body string) error {
	err := r.ch.PublishWithContext(
		ctx,
		"",       // exchange
		r.q.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}
	return nil
}

func (r *RabbitMQClient) Consume() (<-chan amqp.Delivery, error) {
	msgs, err := r.ch.Consume(
		r.q.Name, // queue
		"",       // consumer
		false,    // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)

	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %v", err)
	}
	return msgs, nil
}

func (r *RabbitMQClient) Close() error {
	if err := r.ch.Close(); err != nil {
		return fmt.Errorf("failed to close the channel: %v", err)
	}
	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("failed to close the connection: %v", err)
	}
	return nil
}
