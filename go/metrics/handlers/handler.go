package handlers

import (
	"context"

	"github.com/raywall/cloud-service-pack/go/metrics/types"
)

// MetricHandler defines the interface for a metrics backend client.
// It abstracts the specific implementation details of different metric collectors like Datadog or OpenTelemetry.
type MetricHandler interface {
	// Increment sends a counter metric.
	Increment(ctx context.Context, metric string, value int64, tags types.Tags) error
	// Gauge sends a gauge metric.
	Gauge(ctx context.Context, metric string, value float64, tags types.Tags) error
	// Histogram sends a histogram metric.
	Histogram(ctx context.Context, metric string, value float64, tags types.Tags) error
	// Close flushes and closes the connection to the backend.
	Close() error
}
