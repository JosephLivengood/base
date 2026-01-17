package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type RedisDB struct {
	Client *redis.Client
}

func NewRedis(host string, port string) (*RedisDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
	})

	// Add OpenTelemetry tracing and metrics instrumentation
	if err := redisotel.InstrumentTracing(client); err != nil {
		return nil, fmt.Errorf("failed to instrument redis tracing: %w", err)
	}
	if err := redisotel.InstrumentMetrics(client); err != nil {
		return nil, fmt.Errorf("failed to instrument redis metrics: %w", err)
	}

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisDB{Client: client}, nil
}

func (r *RedisDB) Health(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

func (r *RedisDB) Close() error {
	return r.Client.Close()
}
