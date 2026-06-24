package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sumithsudesan/pkg/config"
	"github.com/sumithsudesan/pkg/database"
	"github.com/sumithsudesan/pkg/logger"
	"github.com/sumithsudesan/pkg/queue"
	"github.com/sumithsudesan/user-service/src/user"
)

func main() {
	// load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// New logger
	log, err := logger.New(cfg.Log.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}

	log.Info("starting service",
		"service", cfg.Service.Name,
		"env", cfg.Service.Env,
		"port", cfg.Service.Port,
	)

	ctx := context.Background()

	// DNB connection
	db, err := database.New(ctx, cfg.Database, log)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// New queure
	pub, err := queue.NewPublisher(cfg.Queue, log)
	if err != nil {
		log.Error("failed to connect to queue", "error", err)
		os.Exit(1)
	}
	defer pub.Close()

	repo := user.NewPostgresRepository(db, log)
	publisher := user.NewQueuePublisher(pub, log)
	svc := user.NewService(repo, publisher, log)
	handler := user.NewHandler(svc)
	router := user.NewRouter(handler)

	// Server start
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Service.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background
	go func() {
		log.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Block until shutdown signal
	quit, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-quit.Done()

	log.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "error", err)
	}

	log.Info("service stopped")
}
