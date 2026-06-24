package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sumithsudesan/pkg/config"
	"github.com/sumithsudesan/pkg/logger"
)

// Sentinel errors — callers check against these, never against pgx types.
var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("record already exists")
)

// postgress
type postgresDB struct {
	pool *pgxpool.Pool
	log  logger.Logger
}

// New returns a DB connection for the configured provider.
func newPostgres(ctx context.Context,
	cfg config.DatabaseConfig,
	log logger.Logger) (DB, error) {

	// Construct the connection string.
	connString := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.User,
		cfg.Password,
		cfg.SSLMode,
	)

	// parse Config to set pool and timeout settings
	poolPgx, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	//
	poolPgx.MaxConns = int32(cfg.Pool.MaxOpen)
	poolPgx.MinConns = int32(cfg.Pool.MaxIdle)
	poolPgx.MaxConnLifetime = time.Duration(cfg.Pool.MaxLifetime) * time.Second
	poolPgx.ConnConfig.ConnectTimeout = time.Duration(cfg.Timeout.Connect) * time.Second

	// Create the connection pool.
	pool, err := pgxpool.NewWithConfig(ctx, poolPgx)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	// check if the connection is alive
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	log.Info("connected to postgres",
		"host", cfg.Host,
		"port", cfg.Port,
		"db", cfg.Name,
	)

	// New posgrees instance
	return &postgresDB{pool: pool, log: log}, nil
}

// QueryRow executes a query that is expected to return at most one row.
func (db *postgresDB) QueryRow(ctx context.Context,
	query string,
	args ...any) Row {

	if db.pool == nil {
		panic("database pool is not initialised")
	}
	return db.pool.QueryRow(ctx, query, args...)
}

// Query executes a query that returns multiple rows.
func (db *postgresDB) Query(ctx context.Context,
	query string,
	args ...any) (Rows, error) {
	if db.pool == nil {
		return nil, fmt.Errorf("database pool is not initialised")
	}

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, mapError(err)
	}
	return rows, nil
}

// Exec executes a query without returning any rows. For example, an INSERT or UPDATE.
func (db *postgresDB) Exec(ctx context.Context,
	query string,
	args ...any) (Result, error) {

	if db.pool == nil {
		return nil, fmt.Errorf("database pool is not initialised")
	}

	// Execute the query.
	tag, err := db.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, mapError(err)
	}
	return tag, nil
}

// Close closes the database connection pool.
func (db *postgresDB) Close() error {
	if db.pool == nil {
		return nil
	}
	db.pool.Close()
	db.pool = nil
	return nil
}

// mapError translates pgx-specific errors into package-level sentinel errors.
func mapError(err error) error {
	if err == nil {
		return nil
	}

	// no rows returned
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	// postgres error codes
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%w: %s", ErrConflict, pgErr.Detail)
		case "23503": // foreign_key_violation
			return fmt.Errorf("%w: %s", ErrNotFound, pgErr.Detail)
		}
	}

	return fmt.Errorf("database error: %w", err)
}
