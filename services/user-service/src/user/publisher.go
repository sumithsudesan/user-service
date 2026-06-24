package user

import "time"

// Event represents a domain event published after a user mutation.
type Event struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	EventType string    `json:"event_type"` // user.created | user.updated | user.deleted
	Timestamp time.Time `json:"timestamp"`
}

// Publisher defines the event publishing contract for the user domain.
// Queue infeface Impl
type Publisher interface {
	//
	Publish(event Event) error
}
