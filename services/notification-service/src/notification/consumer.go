package notification

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Constants for RabbitMQ exchange,
// queue, and routing key.
// TODO: Consider moving these to a configuration file
//
//	or environment variables for flexibility.
const (
	exchangeName = "user.events"
	queueName    = "notification.user.events"
	routingKey   = "user.*"
)

// Represents a RabbitMQ consumer that listens for user events
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	log     *slog.Logger
}

// New creates a new Consumer instance,
// establishes a connection to RabbitMQ,
func New(url string, log *slog.Logger) (*Consumer, error) {
	//
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Open a channel to communicate with RabbitMQ.
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare the exchange, queue, and bind them together.
	if err := ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare the queue and bind it to the exchange with the routing key.
	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	//
	if err := ch.QueueBind(q.Name, routingKey, exchangeName, false, nil); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	// Set the Quality of Service (QoS) to ensure that the consumer
	//
	if err := ch.Qos(1, 0, false); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	// Return the initialized Consumer instance.
	return &Consumer{conn: conn, channel: ch, log: log}, nil
}

// Start begins consuming messages from the RabbitMQ queue.
func (c *Consumer) Start(ctx context.Context) error {
	//
	msgs, err := c.channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.log.Info("notification-service ready", "queue", queueName)

	// wait for messages in a loop, processing each one as it arrives.
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := HandleEvent(msg.Body, c.log); err != nil {
				c.log.Error("failed to handle event", "error", err)
				msg.Nack(false, false)
				continue
			}
			msg.Ack(false)
		}
	}
}

// Close gracefully shuts down the consumer by closing
// the RabbitMQ channel and connection.
func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	//
	if c.conn != nil {
		c.conn.Close()
	}
}
