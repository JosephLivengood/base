package middleware

import (
	"context"
	"net/http"
	"strings"

	"base/api/pkg/response"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthConfig struct {
	// Add your auth configuration here (e.g., JWT secret, auth provider)
	SkipPaths []string
}

// Auth middleware - placeholder for your auth implementation
func Auth(cfg AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for certain paths
			for _, path := range cfg.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "missing authorization header")
				return
			}

			// Extract Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Unauthorized(w, "invalid authorization header format")
				return
			}

			token := parts[1]

			// TODO: Validate token and extract user info
			// This is a placeholder - implement your actual auth logic here
			_ = token

			// Add user to context (replace with actual user data)
			ctx := context.WithValue(r.Context(), UserContextKey, nil)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
