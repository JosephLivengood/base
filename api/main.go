package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"

	"base/api/config"
	"base/api/internal/database"
	"base/api/internal/observability"
	"base/api/internal/router"
)

func main() {
	env := os.Getenv("ENVIRONMENT")
	isDev := env == "development" || env == ""

	// Use DEBUG level in development to see DB query logs
	level := slog.LevelInfo
	if isDev {
		level = slog.LevelDebug
	}

	// Use colored output for development, JSON for production (CloudWatch)
	var handler slog.Handler
	if isDev {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      level,
			TimeFormat: time.Kitchen,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}
	logger := slog.New(handler)

	if err := run(logger); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger.Info("starting server",
		"port", cfg.Port,
		"environment", cfg.Environment,
	)

	debug := cfg.Environment == "development"

	// Initialize metrics client (no-op in development, CloudWatch in production)
	metrics := observability.NewMetrics(logger, cfg.Environment)

	// Initialize PostgreSQL
	postgres, err := database.NewPostgres(cfg.PostgresDSN(), logger, debug)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer postgres.Close()
	logger.Info("connected to postgres")

	// Initialize DynamoDB
	dynamo, err := database.NewDynamo(database.DynamoConfig{
		Endpoint:  cfg.DynamoEndpoint,
		Region:    cfg.DynamoRegion,
		AccessKey: cfg.AWSAccessKey,
		SecretKey: cfg.AWSSecretKey,
		Logger:    logger,
		Debug:     debug,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to dynamodb: %w", err)
	}
	logger.Info("connected to dynamodb")

	// Initialize Redis
	redisDB, err := database.NewRedis(cfg.RedisHost, cfg.RedisPort)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	defer redisDB.Close()
	logger.Info("connected to redis")

	// Setup router
	r := router.New(router.Dependencies{
		Logger:        logger,
		Postgres:      postgres,
		Dynamo:        dynamo,
		Redis:         redisDB,
		Metrics:       metrics,
		SessionSecret: cfg.SessionSecret,
		GoogleConfig: router.GoogleOAuthConfig{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
		},
		Environment: cfg.Environment,
	})

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	logger.Info("server started", "addr", srv.Addr)

	<-done
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	logger.Info("server stopped")
	return nil
}
