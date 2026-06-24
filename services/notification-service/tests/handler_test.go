package tests

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/sumithsudesan/notification-service/src/notification"
)

// newTestLogger creates a new logger instance for testing purposes.
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(bytes.NewBuffer(nil), nil))
}

// TestHandleEvent_ValidPayload tests the HandleEvent function with a valid JSON payload.
func TestHandleEvent_ValidPayload(t *testing.T) {
	body := []byte(`{
		"user_id": "user-123",
		"email": "test@example.com",
		"event_type": "user.created",
		"timestamp": "2026-06-24T10:00:00Z"
	}`)

	if err := notification.HandleEvent(body, newTestLogger()); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// TestHandleEvent_InvalidJSON tests the HandleEvent function with invalid JSON input.
func TestHandleEvent_InvalidJSON(t *testing.T) {
	if err := notification.HandleEvent([]byte(`not-json`), newTestLogger()); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestHandleEvent_EmptyBody(t *testing.T) {
	if err := notification.HandleEvent([]byte(`{}`), newTestLogger()); err != nil {
		t.Fatalf("empty payload should parse without error, got: %v", err)
	}
}

func TestHandleEvent_AllEventTypes(t *testing.T) {
	types := []string{"user.created", "user.updated", "user.deleted"}
	log := newTestLogger()

	for _, et := range types {
		body := []byte(`{"user_id":"u1","email":"a@b.com","event_type":"` + et + `","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`)
		if err := notification.HandleEvent(body, log); err != nil {
			t.Fatalf("event_type %q: unexpected error: %v", et, err)
		}
	}
}
