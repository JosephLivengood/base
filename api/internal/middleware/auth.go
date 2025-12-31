package middleware

import (
	"context"
	"net/http"

	"base/api/internal/domain/user"
	"base/api/internal/session"
	"base/api/pkg/response"
)

type contextKey string

const (
	UserContextKey    contextKey = "user"
	SessionContextKey contextKey = "session"
)

// RequireAuth is middleware that requires a valid session
func RequireAuth(sessionStore *session.Store, userRepo *user.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				response.Unauthorized(w, "authentication required")
				return
			}

			sess, err := sessionStore.Get(r.Context(), cookie.Value)
			if err != nil {
				response.Unauthorized(w, "invalid or expired session")
				return
			}

			usr, err := userRepo.GetByID(r.Context(), sess.UserID)
			if err != nil {
				response.Unauthorized(w, "user not found")
				return
			}

			// Add session and user to context
			ctx := context.WithValue(r.Context(), SessionContextKey, sess)
			ctx = context.WithValue(ctx, UserContextKey, usr)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) *user.User {
	if usr, ok := ctx.Value(UserContextKey).(*user.User); ok {
		return usr
	}
	return nil
}

// GetSessionFromContext retrieves the session from the request context
func GetSessionFromContext(ctx context.Context) *session.Session {
	if sess, ok := ctx.Value(SessionContextKey).(*session.Session); ok {
		return sess
	}
	return nil
}
