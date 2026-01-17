package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			// Get request ID from chi middleware
			requestID := chimiddleware.GetReqID(r.Context())

			// Extract trace ID from OpenTelemetry span context for log correlation
			var traceID, spanID string
			spanCtx := trace.SpanContextFromContext(r.Context())
			if spanCtx.HasTraceID() {
				traceID = spanCtx.TraceID().String()
			}
			if spanCtx.HasSpanID() {
				spanID = spanCtx.SpanID().String()
			}

			logger.Info("request",
				"request_id", requestID,
				"trace_id", traceID,
				"span_id", spanID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
				"ip", r.RemoteAddr,
			)
		})
	}
}
