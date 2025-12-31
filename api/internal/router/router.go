package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"base/api/internal/database"
	"base/api/internal/domain/health"
	"base/api/internal/domain/ping"
	"base/api/internal/middleware"
)

type Dependencies struct {
	Logger   *slog.Logger
	Postgres *database.PostgresDB
	Dynamo   *database.DynamoDB
}

func New(deps Dependencies) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logging(deps.Logger))
	r.Use(middleware.CORS(middleware.DefaultCORSConfig()))
	r.Use(chimiddleware.Recoverer)

	// Health routes (no auth required)
	healthHandler := health.NewHandler(deps.Postgres, deps.Dynamo)
	r.Route("/health", func(r chi.Router) {
		health.RegisterRoutes(r, healthHandler)
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Ping route
		pingRepo := ping.NewRepository(deps.Postgres, deps.Dynamo)
		pingHandler := ping.NewHandler(pingRepo)
		r.Route("/ping", func(r chi.Router) {
			ping.RegisterRoutes(r, pingHandler)
		})
	})

	return r
}
