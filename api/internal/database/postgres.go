package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/jmoiron/sqlx"
)

type PostgresDB struct {
	*sqlx.DB
	pool *pgxpool.Pool
}

type slogAdapter struct {
	logger *slog.Logger
}

func (a *slogAdapter) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	attrs := make([]slog.Attr, 0, len(data))
	for k, v := range data {
		attrs = append(attrs, slog.Any(k, v))
	}
	a.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
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

	// Add query tracing in debug mode
	if debug && logger != nil {
		cfg.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   &slogAdapter{logger: logger.With("db", "postgres")},
			LogLevel: tracelog.LogLevelDebug,
		}
	}

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
