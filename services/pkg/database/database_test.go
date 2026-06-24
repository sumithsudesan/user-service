package database_test

import (
	"context"
	"testing"

	"github.com/sumithsudesan/pkg/config"
	"github.com/sumithsudesan/pkg/database"
	"github.com/sumithsudesan/pkg/logger"
)

func newTestLogger(t *testing.T) logger.Logger {
	t.Helper()
	l, err := logger.New("info")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	return l
}

func TestNew_UnsupportedProvider(t *testing.T) {
	cfg := config.DatabaseConfig{Provider: "oracle"}
	_, err := database.New(context.Background(), cfg, newTestLogger(t))
	if err == nil {
		t.Fatal("expected error for unsupported provider, got nil")
	}
}

func TestNew_EmptyProvider(t *testing.T) {
	cfg := config.DatabaseConfig{Provider: ""}
	_, err := database.New(context.Background(), cfg, newTestLogger(t))
	if err == nil {
		t.Fatal("expected error for empty provider, got nil")
	}
}

// TestNew_Postgres is an integration test — requires a real PostgreSQL instance.
// Run with: go test -run TestNew_Postgres -tags integration
//
// func TestNew_Postgres(t *testing.T) {
//     cfg := config.DatabaseConfig{
//         Provider: "postgres",
//         Host:     "localhost", Port: 5432,
//         Name: "testdb", User: "postgres", Password: "secret",
//         SSLMode: "disable",
//         Pool:    config.PoolConfig{MaxOpen: 5, MaxIdle: 1, MaxLifetime: 60},
//         Timeout: config.TimeoutConfig{Connect: 5, Query: 10},
//     }
//     db, err := database.New(context.Background(), cfg, newTestLogger(t))
//     if err != nil { t.Fatalf("unexpected error: %v", err) }
//     defer db.Close()
// }
