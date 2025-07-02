package handlers

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/raywall/cloud-service-pack/go/metrics/types"
)

// DatadogMetricHandler implements the MetricHandler interface for Datadog.
type DatadogMetricHandler struct {
	client *statsd.Client
}

// NewDatadogMetricHandler creates a new Datadog client using the statsd protocol.
// It configures the client with a namespace (metric prefix) and server address.
func NewDatadogMetricHandler(metricPrefix, serverHost string, serverPort int) (*DatadogMetricHandler, error) {
	prefix := ""
	if metricPrefix != "" {
		prefix = fmt.Sprintf("%s.", metricPrefix)
	}

	client, err := statsd.New(fmt.Sprintf("%s:%d", serverHost, serverPort), statsd.WithNamespace(prefix))
	if err != nil {
		return nil, fmt.Errorf("failed to create a StatsD client for DD: %w", err)
	}

	return &DatadogMetricHandler{
		client: client,
	}, nil
}

// Increment increments a counter metric in Datadog.
func (dd *DatadogMetricHandler) Increment(ctx context.Context, metric string, value int64, tags types.Tags) error {
	// Datadog's Incr function already handles the value, so we pass value, not just 1.0.
	err := dd.client.Incr(metric, tags.ToStringArray(), float64(value))
	if err != nil {
		return fmt.Errorf("failed to increment the counter: %w", err)
	}
	return nil
}

// Gauge sets a gauge metric value in Datadog.
func (dd *DatadogMetricHandler) Gauge(ctx context.Context, metric string, value float64, tags types.Tags) error {
	// BUG FIX: The original code had a hardcoded metric name "memory.usage". This is now corrected.
	err := dd.client.Gauge(metric, value, tags.ToStringArray(), 1.0)
	if err != nil {
		return fmt.Errorf("failed to register the gauge value: %w", err)
	}
	return nil
}

// Histogram sends a histogram sample value to Datadog.
func (dd *DatadogMetricHandler) Histogram(ctx context.Context, metric string, value float64, tags types.Tags) error {
	err := dd.client.Histogram(metric, value, tags.ToStringArray(), 1.0)
	if err != nil {
		return fmt.Errorf("failed to register the histogram value: %w", err)
	}
	return nil
}

// Distribution sends a distribution sample, used for percentile analysis in Datadog.
// Note: This method is not part of the common MetricHandler interface.
func (dd *DatadogMetricHandler) Distribution(ctx context.Context, metric string, value float64, tags types.Tags) error {
	err := dd.client.Distribution(metric, value, tags.ToStringArray(), 1.0)
	if err != nil {
		return fmt.Errorf("failed to register the distribution value: %w", err)
	}
	return nil
}

// Event sends a custom event to the Datadog event stream.
// Note: This method is not part of the common MetricHandler interface.
func (dd *DatadogMetricHandler) Event(ctx context.Context, title, text string, tags types.Tags) error {
	err := dd.client.Event(&statsd.Event{
		Title: title,
		Text:  text,
		Tags:  tags.ToStringArray(),
	})
	if err != nil {
		return fmt.Errorf("failed to send a new event: %w", err)
	}
	return nil
}

// Close closes the underlying statsd client connection.
func (dd *DatadogMetricHandler) Close() error {
	return dd.client.Close()
}
