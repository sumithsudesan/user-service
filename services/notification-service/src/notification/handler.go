package notification

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sumithsudesan/pkg/logger"
)

// Represents the structure of the user event received from the queue.
type Event struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

// new Event instnace
func HandleEvent(body []byte, log logger.Logger) error {
	var event Event
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to parse event: %w", err)
	}

	log.Info("received user event",
		"service", "notification-service",
		"event_type", event.EventType,
		"user_id", event.UserID,
		"email", event.Email,
		"timestamp", event.Timestamp,
	)

	return nil
}
