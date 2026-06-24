package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumithsudesan/pkg/logger"
	"github.com/sumithsudesan/pkg/queue"
)

// Represent message Queue publisher
type queuePublisher struct {
	pub queue.Publisher
	log logger.Logger
}

// Create new instance of queuePublisher
func NewQueuePublisher(pub queue.Publisher,
	log logger.Logger) Publisher {
	return &queuePublisher{pub: pub, log: log}
}

// Publsh message to queue
func (p *queuePublisher) Publish(event Event) error {
	// Convert to JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// pubish to queue
	if err := p.pub.Publish(context.Background(), queue.Message{
		RoutingKey: event.EventType,
		Body:       body,
	}); err != nil {
		p.log.Error("failed to publish event",
			"event_type", event.EventType,
			"user_id", event.UserID,
			"error", err,
		)
		return err
	}

	p.log.Info("event published",
		"event_type", event.EventType,
		"user_id", event.UserID,
	)
	return nil
}
