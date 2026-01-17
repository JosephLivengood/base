package observability

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// OTel holds the OpenTelemetry providers for graceful shutdown.
type OTel struct {
	tracerProvider *sdktrace.TracerProvider
	logger         *slog.Logger
}

// InitOTel initializes OpenTelemetry tracing with OTLP HTTP exporter.
// Configuration is read from environment variables:
//   - OTEL_EXPORTER_OTLP_ENDPOINT: The OTLP endpoint URL (e.g., http://localhost:4318)
//   - OTEL_SERVICE_NAME: The service name for traces (e.g., base2-api)
//
// Returns nil if OTEL_EXPORTER_OTLP_ENDPOINT is not set (disabled).
func InitOTel(logger *slog.Logger) (*OTel, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		logger.Info("OpenTelemetry disabled (OTEL_EXPORTER_OTLP_ENDPOINT not set)")
		return nil, nil
	}

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "base2-api"
	}

	ctx := context.Background()

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create OTLP trace exporter
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(stripProtocol(endpoint)),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create tracer provider with batch processor
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("OpenTelemetry initialized",
		"endpoint", endpoint,
		"service", serviceName,
	)

	return &OTel{
		tracerProvider: tracerProvider,
		logger:         logger,
	}, nil
}

// Shutdown gracefully shuts down the OpenTelemetry providers.
func (o *OTel) Shutdown() {
	if o == nil || o.tracerProvider == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := o.tracerProvider.Shutdown(ctx); err != nil {
		o.logger.Error("failed to shutdown tracer provider", "error", err)
	}
}

// stripProtocol removes http:// or https:// prefix from the endpoint.
// OTLP HTTP client expects just the host:port.
func stripProtocol(endpoint string) string {
	if len(endpoint) > 8 && endpoint[:8] == "https://" {
		return endpoint[8:]
	}
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		return endpoint[7:]
	}
	return endpoint
}
