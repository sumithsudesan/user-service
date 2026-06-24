package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sumithsudesan/notification-service/src/notification"
	"github.com/sumithsudesan/pkg/config"
	"github.com/sumithsudesan/pkg/logger"
	"github.com/sumithsudesan/pkg/queue"
)

func main() {
	// Load condiguration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// new loger
	log, err := logger.New(cfg.Log.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}

	log.Info("starting service",
		"service", cfg.Service.Name,
		"env", cfg.Service.Env,
	)

	// new consumer
	consumer, err := queue.NewConsumer(cfg.Queue, log)
	if err != nil {
		log.Error("failed to connect to queue", "error", err)
		os.Exit(1)
	}
	defer consumer.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// start consume
	if err := consumer.Consume(ctx, func(msg queue.Message) error {
		return notification.HandleEvent(msg.Body, log)
	}); err != nil {
		log.Error("consumer error", "error", err)
		os.Exit(1)
	}

	log.Info("notification-service stopped")
}
