package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"base/api/internal/database"
	"base/api/internal/domain/health"
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

	// API routes (with auth)
	r.Route("/api/v1", func(r chi.Router) {
		// Add auth middleware for API routes
		// r.Use(middleware.Auth(middleware.AuthConfig{
		// 	SkipPaths: []string{"/api/v1/auth"},
		// }))

		// Mount domain routes here
		// Example:
		// usersService := users.NewService(deps.Postgres)
		// usersHandler := users.NewHandler(usersService)
		// r.Route("/users", func(r chi.Router) {
		// 	users.RegisterRoutes(r, usersHandler)
		// })
	})

	return r
}
