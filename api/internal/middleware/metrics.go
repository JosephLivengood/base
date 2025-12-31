package middleware

import (
	"net/http"
	"time"

	"base/api/internal/observability"
)

func Metrics(metrics observability.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			// Record request metrics
			metrics.RecordRequest(r.Method, r.URL.Path, wrapped.statusCode, duration)

			// Record error metrics for 4xx and 5xx responses
			if wrapped.statusCode >= 400 {
				metrics.RecordError(r.Method, r.URL.Path, wrapped.statusCode)
			}
		})
	}
}
