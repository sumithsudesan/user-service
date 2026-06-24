package queue

import (
	"context"
	"fmt"

	"github.com/sumithsudesan/pkg/config"
	"github.com/sumithsudesan/pkg/logger"
)

// Message is the unit of data exchanged between services.
type Message struct {
	RoutingKey string
	Body       []byte
}

// HandlerFunc is the function signature for consuming messages.
type HandlerFunc func(msg Message) error

// Publisher sends messages to the message broker.
// Add new cases to NewPublisher() to support additional providers.
type Publisher interface {
	Publish(ctx context.Context, msg Message) error
	Close() error
}

// Consumer receives messages from the message broker.
// Add new cases to NewConsumer() to support additional providers.
type Consumer interface {
	Consume(ctx context.Context, handler HandlerFunc) error
	Close() error
}

// NewPublisher returns a Publisher for the configured provider.
func NewPublisher(cfg config.QueueConfig, log logger.Logger) (Publisher, error) {
	switch cfg.Provider {
	// rabbitmq
	case "rabbitmq":
		return newRabbitMQPublisher(cfg, log)
	default:
		return nil, fmt.Errorf("unsupported queue provider: %q",
			cfg.Provider)
	}
}

// NewConsumer returns a Consumer for the configured provider.
func NewConsumer(cfg config.QueueConfig, log logger.Logger) (Consumer, error) {
	switch cfg.Provider {

	// rabbitmq
	case "rabbitmq":
		return newRabbitMQConsumer(cfg, log)
	default:
		return nil, fmt.Errorf("unsupported queue provider: %q",
			cfg.Provider)
	}
}
