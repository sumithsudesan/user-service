package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sumithsudesan/pkg/config"
)

// testYAML is a sample YAML configuration for testing
const testYAML = `
service:
  name: test-service
  port: 9090
  env: test

log:
  level: debug

database:
  provider: postgres
  host: localhost
  port: 5432
  name: testdb
  user: testuser
  password: secret
  ssl_mode: disable
  pool:
    max_open: 10
    max_idle: 2
    max_lifetime: 300
  timeout:
    connect: 5
    query: 30

queue:
  provider: rabbitmq
  host: localhost
  port: 5672
  user: guest
  password: guest
  exchange:
    name: user.events
    type: topic
    durable: true
  queue:
    name: notification.user.events
    routing_key: "user.*"
    durable: true
  retry:
    max_attempts: 3
    interval: 5
  dlq:
    enabled: true
    exchange: user.events.dlx
    queue: notification.user.events.dlq
`

// writeTempConfig creates a temporary YAML config file for testing
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	f.Close()
	return filepath.ToSlash(f.Name())
}

// TestLoad_ValidConfig tests loading a valid configuration file
func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, testYAML)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Validate some key fields
	if cfg.Service.Name != "test-service" {
		t.Errorf("expected service.name=test-service, got %q", cfg.Service.Name)
	}
	if cfg.Service.Port != 9090 {
		t.Errorf("expected service.port=9090, got %d", cfg.Service.Port)
	}
	if cfg.Database.Provider != "postgres" {
		t.Errorf("expected database.provider=postgres, got %q", cfg.Database.Provider)
	}
	if cfg.Queue.Provider != "rabbitmq" {
		t.Errorf("expected queue.provider=rabbitmq, got %q", cfg.Queue.Provider)
	}
	if cfg.Database.Pool.MaxOpen != 10 {
		t.Errorf("expected pool.max_open=10, got %d", cfg.Database.Pool.MaxOpen)
	}
}

// TestLoad_EnvVarOverride tests that environment variables
//
//	override config values
func TestLoad_EnvVarOverride(t *testing.T) {
	path := writeTempConfig(t, testYAML)

	t.Setenv("DATABASE_PASSWORD", "overridden-secret")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Database.Password != "overridden-secret" {
		t.Errorf("expected env var to override password, got %q", cfg.Database.Password)
	}
}

// TestLoad_MissingFile tests loading a non-existent configuration file
func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// TestLoad_MissingServiceName tests that the validation fails
// when service.name is missing
func TestLoad_MissingServiceName(t *testing.T) {
	yaml := `
service:
  port: 8080
`
	path := writeTempConfig(t, yaml)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing service.name, got nil")
	}
}
