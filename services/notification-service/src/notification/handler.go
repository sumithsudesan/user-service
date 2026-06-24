package notification

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// Represents the structure of the incoming event data.
type Event struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

// HandleEvent processes the incoming event data,
// logs the event details, and returns an error
// if the event cannot be parsed.
func HandleEvent(body []byte, log *slog.Logger) error {
	var event Event
	// unmarshal the JSON into the Event struct.
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to parse event: %w", err)
	}

	// Log the event details using structured logging.
	log.Info("received user event",
		"service", "notification-service",
		"event_type", event.EventType,
		"user_id", event.UserID,
		"email", event.Email,
		"timestamp", event.Timestamp,
	)

	return nil
}
