package queue_test

import (
	"testing"

	"github.com/sumithsudesan/pkg/config"
	"github.com/sumithsudesan/pkg/logger"
	"github.com/sumithsudesan/pkg/queue"
)

// new test loger instance
func newTestLogger(t *testing.T) logger.Logger {
	t.Helper()
	l, err := logger.New("info")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	return l
}

// publihser - unsupported provider
func TestNewPublisher_UnsupportedProvider(t *testing.T) {
	cfg := config.QueueConfig{Provider: "kafka"}
	_, err := queue.NewPublisher(cfg, newTestLogger(t))
	if err == nil {
		t.Fatal("expected error for unsupported provider, got nil")
	}
}

// consueme - unsupported provider
func TestNewConsumer_UnsupportedProvider(t *testing.T) {
	cfg := config.QueueConfig{Provider: "sqs"}
	_, err := queue.NewConsumer(cfg, newTestLogger(t))
	if err == nil {
		t.Fatal("expected error for unsupported provider, got nil")
	}
}

// epty provider
func TestNewPublisher_EmptyProvider(t *testing.T) {
	cfg := config.QueueConfig{Provider: ""}
	_, err := queue.NewPublisher(cfg, newTestLogger(t))
	if err == nil {
		t.Fatal("expected error for empty provider, got nil")
	}
}
