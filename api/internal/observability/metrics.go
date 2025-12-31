package observability

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// Metrics defines the interface for recording application metrics.
type Metrics interface {
	RecordRequest(method, path string, statusCode int, duration time.Duration)
	RecordError(method, path string, statusCode int)
}

// CloudWatchMetrics sends metrics to AWS CloudWatch.
type CloudWatchMetrics struct {
	client    *cloudwatch.Client
	namespace string
	logger    *slog.Logger
}

// NoopMetrics discards all metrics (used in development).
type NoopMetrics struct {
	logger *slog.Logger
}

// NewMetrics creates a Metrics implementation based on environment.
// In development (or when AWS_REGION is not set), returns a no-op implementation.
// In production, returns a CloudWatch metrics client.
func NewMetrics(logger *slog.Logger, environment string) Metrics {
	// Skip CloudWatch in development
	if environment == "development" || environment == "" {
		logger.Info("metrics disabled (development mode)")
		return &NoopMetrics{logger: logger}
	}

	// Check for AWS region - required for CloudWatch
	region := os.Getenv("AWS_REGION")
	if region == "" {
		logger.Warn("AWS_REGION not set, metrics disabled")
		return &NoopMetrics{logger: logger}
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		logger.Error("failed to load AWS config, metrics disabled", "error", err)
		return &NoopMetrics{logger: logger}
	}

	client := cloudwatch.NewFromConfig(cfg)

	namespace := os.Getenv("CLOUDWATCH_NAMESPACE")
	if namespace == "" {
		namespace = "App/API"
	}

	logger.Info("CloudWatch metrics enabled", "namespace", namespace, "region", region)

	return &CloudWatchMetrics{
		client:    client,
		namespace: namespace,
		logger:    logger,
	}
}

// RecordRequest records request count and latency metrics.
func (m *CloudWatchMetrics) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dimensions := []types.Dimension{
		{Name: aws.String("Method"), Value: aws.String(method)},
		{Name: aws.String("Path"), Value: aws.String(normalizePath(path))},
	}

	_, err := m.client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(m.namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("RequestCount"),
				Dimensions: dimensions,
				Value:      aws.Float64(1),
				Unit:       types.StandardUnitCount,
			},
			{
				MetricName: aws.String("RequestLatency"),
				Dimensions: dimensions,
				Value:      aws.Float64(float64(duration.Milliseconds())),
				Unit:       types.StandardUnitMilliseconds,
			},
		},
	})

	if err != nil {
		m.logger.Error("failed to put metrics", "error", err)
	}
}

// RecordError records an error metric.
func (m *CloudWatchMetrics) RecordError(method, path string, statusCode int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(m.namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("ErrorCount"),
				Dimensions: []types.Dimension{
					{Name: aws.String("Method"), Value: aws.String(method)},
					{Name: aws.String("Path"), Value: aws.String(normalizePath(path))},
					{Name: aws.String("StatusCode"), Value: aws.String(statusCodeCategory(statusCode))},
				},
				Value: aws.Float64(1),
				Unit:  types.StandardUnitCount,
			},
		},
	})

	if err != nil {
		m.logger.Error("failed to put error metric", "error", err)
	}
}

// RecordRequest logs metrics in development (no-op for CloudWatch).
func (m *NoopMetrics) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	m.logger.Debug("metric: request",
		"method", method,
		"path", path,
		"status", statusCode,
		"duration_ms", duration.Milliseconds(),
	)
}

// RecordError logs errors in development (no-op for CloudWatch).
func (m *NoopMetrics) RecordError(method, path string, statusCode int) {
	m.logger.Debug("metric: error",
		"method", method,
		"path", path,
		"status", statusCode,
	)
}

// normalizePath reduces cardinality by removing dynamic path segments.
// e.g., /api/users/123 -> /api/users/:id
func normalizePath(path string) string {
	// Keep paths as-is for now, but this could be extended
	// to replace UUIDs, numeric IDs, etc. with placeholders
	return path
}

// statusCodeCategory groups status codes for metric dimensions.
func statusCodeCategory(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	default:
		return "other"
	}
}
