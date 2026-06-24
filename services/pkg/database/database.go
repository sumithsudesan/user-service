package database

import (
	"context"
	"fmt"

	"github.com/sumithsudesan/pkg/config"
	"github.com/sumithsudesan/pkg/logger"
)

// DB is the database abstraction interface.
// support additional providers (mysql, mongodb, dynamodb).
type DB interface {
	QueryRow(ctx context.Context, query string, args ...any) Row
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Close() error
}

// Row represents a single result row returned by QueryRow.
type Row interface {
	Scan(dest ...any) error
}

// Rows represents multiple result rows returned by Query.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
}

// Result holds the outcome of an Exec call.
type Result interface {
	RowsAffected() int64
}

// New returns a DB connection for the configured provider.
// The provider is selected by cfg.Provider — add new cases here
// when introducing new database providers.
func New(ctx context.Context, cfg config.DatabaseConfig, log logger.Logger) (DB, error) {
	switch cfg.Provider {
	case "postgres":
		return newPostgres(ctx, cfg, log)
	default:
		return nil, fmt.Errorf("unsupported database provider: %q", cfg.Provider)
	}
}
