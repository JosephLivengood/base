package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"

	"base/api/internal/database"
	"base/api/internal/domain/auth"
	"base/api/internal/domain/health"
	"base/api/internal/domain/organization"
	"base/api/internal/domain/ping"
	"base/api/internal/domain/project"
	"base/api/internal/domain/subscription"
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

type StripeConfig struct {
	SecretKey      string
	PublishableKey string
	WebhookSecret  string
}

type Dependencies struct {
	Logger        *slog.Logger
	Postgres      *database.PostgresDB
	Dynamo        *database.DynamoDB
	Redis         *database.RedisDB
	Metrics       observability.Metrics
	SessionSecret string
	GoogleConfig  GoogleOAuthConfig
	StripeConfig  StripeConfig
	Environment   string
}

func New(deps Dependencies) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(otelchi.Middleware("base2-api", otelchi.WithChiRoutes(r)))
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
		// Repositories
		userRepo := user.NewRepository(deps.Postgres)
		orgRepo := organization.NewRepository(deps.Postgres)

		// Auth routes
		secureCookies := deps.Environment != "development"
		authConfig := auth.NewConfig(
			deps.GoogleConfig.ClientID,
			deps.GoogleConfig.ClientSecret,
			deps.GoogleConfig.RedirectURL,
			secureCookies,
		)
		authHandler := auth.NewHandler(authConfig, userRepo, orgRepo, sessionStore)
		r.Route("/auth", func(r chi.Router) {
			auth.RegisterRoutes(r, authHandler)
		})

		// Ping route
		pingRepo := ping.NewRepository(deps.Postgres, deps.Dynamo)
		pingHandler := ping.NewHandler(pingRepo)
		r.Route("/ping", func(r chi.Router) {
			ping.RegisterRoutes(r, pingHandler)
		})

		// Auth middleware for protected routes
		authMiddleware := middleware.RequireAuth(sessionStore, userRepo)

		// Initialize Stripe service
		stripeService := subscription.NewStripeService(
			deps.StripeConfig.SecretKey,
			deps.StripeConfig.WebhookSecret,
			deps.Environment,
		)

		// Organization routes (protected)
		orgHandler := organization.NewHandler(orgRepo, userRepo, sessionStore)
		projectRepo := project.NewRepository(deps.Postgres)
		projectHandler := project.NewHandler(projectRepo, orgRepo)
		subscriptionRepo := subscription.NewRepository(deps.Postgres)
		subscriptionHandler := subscription.NewHandler(subscriptionRepo, orgRepo, stripeService, deps.Environment)

		r.Route("/organizations", func(r chi.Router) {
			r.Use(authMiddleware)
			organization.RegisterRoutes(r, orgHandler)

			// Nested routes under /organizations/{orgID}
			r.Route("/{orgID}/projects", func(r chi.Router) {
				project.RegisterRoutes(r, projectHandler)
			})
			r.Route("/{orgID}/subscription", func(r chi.Router) {
				subscription.RegisterRoutes(r, subscriptionHandler)
			})
		})

		// User invitations routes (protected)
		r.Route("/invitations", func(r chi.Router) {
			r.Use(authMiddleware)
			organization.RegisterInvitationRoutes(r, orgHandler)
		})

		// Stripe webhook route (no auth required)
		r.Route("/stripe", func(r chi.Router) {
			subscription.RegisterWebhookRoute(r, subscriptionHandler)
		})
	})

	return r
}
