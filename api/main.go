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
	"base/api/internal/router"
)

func main() {
	// Use DEBUG level in development to see DB query logs
	level := slog.LevelInfo
	if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "" {
		level = slog.LevelDebug
	}

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      level,
		TimeFormat: time.Kitchen,
	}))

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

	// Setup router
	r := router.New(router.Dependencies{
		Logger:   logger,
		Postgres: postgres,
		Dynamo:   dynamo,
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
