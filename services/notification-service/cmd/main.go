package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sumithsudesan/notification-service/src/notification"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	c, err := notification.New(amqpURL, log)
	if err != nil {
		log.Error("failed to initialise consumer", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := c.Start(ctx); err != nil {
		log.Error("consumer error", "error", err)
		os.Exit(1)
	}

	log.Info("notification-service stopped")
}
