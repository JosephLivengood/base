package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type PostgresDB struct {
	*sqlx.DB
	pool *pgxpool.Pool
}

func NewPostgres(dsn string, logger *slog.Logger, debug bool) (*PostgresDB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	// Configure connection pool
	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 5 * time.Minute

	// Add OpenTelemetry tracing (uses global tracer provider)
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	// Suppress unused parameter warnings (debug logging handled by OTel)
	_ = debug
	_ = logger

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Wrap pool for sqlx compatibility
	db := sqlx.NewDb(stdlib.OpenDBFromPool(pool), "pgx")

	return &PostgresDB{DB: db, pool: pool}, nil
}

func (db *PostgresDB) Health(ctx context.Context) error {
	return db.PingContext(ctx)
}

func (db *PostgresDB) Close() error {
	db.pool.Close()
	return db.DB.Close()
}
