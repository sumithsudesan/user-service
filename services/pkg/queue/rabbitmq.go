package queue

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sumithsudesan/pkg/config"
	"github.com/sumithsudesan/pkg/logger"
)

// -- Publisher --
// Represnt the Rabitmq publisher
type rabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     config.QueueConfig
	log     logger.Logger
}

// Created new Rabitmq publisher instance
func newRabbitMQPublisher(cfg config.QueueConfig, log logger.Logger) (Publisher, error) {
	// Connect to MQ
	conn, ch, err := connectRabbitMQ(cfg)
	if err != nil {
		return nil, err
	}

	//
	if err := declareExchange(ch, cfg); err != nil {
		conn.Close()
		return nil, err
	}

	log.Info("rabbitmq publisher connected",
		"host", cfg.Host,
		"exchange", cfg.Exchange.Name,
	)

	// new instance
	return &rabbitMQPublisher{conn: conn,
		channel: ch,
		cfg:     cfg,
		log:     log}, nil
}

// Publish message
func (p *rabbitMQPublisher) Publish(
	ctx context.Context,
	msg Message) error {

	// Into JSON
	body, err := json.Marshal(msg.Body)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish
	err = p.channel.PublishWithContext(ctx,
		p.cfg.Exchange.Name,
		msg.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.log.Debug("message published", "routing_key", msg.RoutingKey)
	return nil
}

// Close connection
func (p *rabbitMQPublisher) Close() error {

	if p.channel != nil {
		p.channel.Close()
	}

	if p.conn != nil {
		p.conn.Close()
	}
	return nil
}

// -- Consumer ---
// Represnt RabitMQ consumer
type rabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     config.QueueConfig
	log     logger.Logger
}

// New RabitMQ consumer Instance
func newRabbitMQConsumer(cfg config.QueueConfig,
	log logger.Logger) (Consumer, error) {
	// Connects to Rambimq
	conn, ch, err := connectRabbitMQ(cfg)
	if err != nil {
		return nil, err
	}

	if err := declareExchange(ch, cfg); err != nil {
		conn.Close()
		return nil, err
	}

	// Declare queue
	q, err := ch.QueueDeclare(
		cfg.Queue.Name,
		cfg.Queue.Durable,
		false, false, false, nil,
	)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue %q: %w", cfg.Queue.Name, err)
	}

	if err := ch.QueueBind(q.Name, cfg.Queue.RoutingKey, cfg.Exchange.Name, false, nil); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	log.Info("rabbitmq consumer connected",
		"host", cfg.Host,
		"queue", cfg.Queue.Name,
		"routing_key", cfg.Queue.RoutingKey,
	)

	// New RabitMQ consumer instnace
	return &rabbitMQConsumer{conn: conn, channel: ch, cfg: cfg, log: log}, nil
}

// Consumes the messahe from QUEUE
func (c *rabbitMQConsumer) Consume(ctx context.Context, handler HandlerFunc) error {
	// Cosues
	msgs, err := c.channel.Consume(c.cfg.Queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	// Wait for
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-msgs:
			if !ok {
				return nil
			}
			err := handler(Message{
				RoutingKey: delivery.RoutingKey,
				Body:       delivery.Body,
			})
			if err != nil {
				c.log.Error("handler failed, nacking message",
					"routing_key", delivery.RoutingKey,
					"error", err,
				)
				delivery.Nack(false, false)
				continue
			}
			delivery.Ack(false)
		}
	}
}

// Closes connection
func (c *rabbitMQConsumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

// -- helpers functions --

// Connect Message QUEUE
func connectRabbitMQ(cfg config.QueueConfig) (*amqp.Connection,
	*amqp.Channel,
	error) {
	// prepare connection url
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port)

	// Connect MQ
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	// Open chanll
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return conn, ch, nil
}

// Declare exchange
func declareExchange(ch *amqp.Channel,
	cfg config.QueueConfig) error {
	return ch.ExchangeDeclare(
		cfg.Exchange.Name,
		cfg.Exchange.Type,
		cfg.Exchange.Durable,
		false, false, false, nil,
	)
}
