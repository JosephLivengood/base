package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"base/api/internal/database"
	"base/api/internal/domain/auth"
	"base/api/internal/domain/health"
	"base/api/internal/domain/ping"
	"base/api/internal/domain/user"
	"base/api/internal/middleware"
	"base/api/internal/observability"
	"base/api/internal/session"
)

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Dependencies struct {
	Logger        *slog.Logger
	Postgres      *database.PostgresDB
	Dynamo        *database.DynamoDB
	Redis         *database.RedisDB
	Metrics       observability.Metrics
	SessionSecret string
	GoogleConfig  GoogleOAuthConfig
	Environment   string
}

func New(deps Dependencies) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Recovery(deps.Logger))
	r.Use(middleware.Logging(deps.Logger))
	r.Use(middleware.Metrics(deps.Metrics))
	r.Use(middleware.CORS(middleware.DefaultCORSConfig()))

	// Initialize session store
	sessionStore := session.NewStore(deps.Redis, deps.SessionSecret)

	// Health routes (no auth required)
	healthHandler := health.NewHandler(deps.Postgres, deps.Dynamo, deps.Redis)
	r.Route("/health", func(r chi.Router) {
		health.RegisterRoutes(r, healthHandler)
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes
		userRepo := user.NewRepository(deps.Postgres)
		secureCookies := deps.Environment != "development"
		authConfig := auth.NewConfig(
			deps.GoogleConfig.ClientID,
			deps.GoogleConfig.ClientSecret,
			deps.GoogleConfig.RedirectURL,
			secureCookies,
		)
		authHandler := auth.NewHandler(authConfig, userRepo, sessionStore)
		r.Route("/auth", func(r chi.Router) {
			auth.RegisterRoutes(r, authHandler)
		})

		// Ping route
		pingRepo := ping.NewRepository(deps.Postgres, deps.Dynamo)
		pingHandler := ping.NewHandler(pingRepo)
		r.Route("/ping", func(r chi.Router) {
			ping.RegisterRoutes(r, pingHandler)
		})
	})

	return r
}
