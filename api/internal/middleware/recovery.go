package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					requestID := chimiddleware.GetReqID(r.Context())
					stack := string(debug.Stack())

					logger.Error("panic recovered",
						"request_id", requestID,
						"error", rec,
						"method", r.Method,
						"path", r.URL.Path,
						"ip", r.RemoteAddr,
						"stack", stack,
					)

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
