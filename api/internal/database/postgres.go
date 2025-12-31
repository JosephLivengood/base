package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDB struct {
	*sqlx.DB
}

func NewPostgres(dsn string) (*PostgresDB, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresDB{DB: db}, nil
}

func (db *PostgresDB) Health(ctx context.Context) error {
	return db.PingContext(ctx)
}

func (db *PostgresDB) Close() error {
	return db.DB.Close()
}
