package handlers

import (
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/raywall/cloud-service-pack/go/metrics/types"
)

// DatadogMetricHandler is a struct that represents a Datadog client configuration
type DatadogMetricHandler struct {
	client *statsd.Client
}

// NewDatadogMetricHandler creates a new Datadog client configuration
func NewDatadogMetricHandler(metric_prefix, server_host string, server_port int) (*DatadogMetricHandler, error) {
	prefix := ""

	if metric_prefix != "" {
		prefix = fmt.Sprintf("%s.", metric_prefix)
	}

	if _client, err := statsd.New(fmt.Sprintf("%s:%d", server_host, server_port), statsd.WithNamespace(prefix)); err != nil {
		return nil, fmt.Errorf("failed to create a StatsD client for DD: %v", err)

	} else {
		return &DatadogMetricHandler{
			client: _client,
		}, nil
	}
}

// Increment helps to increment a value in a Datadog custom metric
func (dd *DatadogMetricHandler) Increment(metric string, value int, tags types.Tags) error {
	err := dd.client.Incr(fmt.Sprintf("%s.count", metric), tags.ToStringArray(), 1.0)
	if err != nil {
		return fmt.Errorf("failed to increment the counter: %v", err)
	}

	return nil
}

// Gauge defines a static value in a Datadog custom metric
func (dd *DatadogMetricHandler) Gauge(metric string, value float64, tags types.Tags) error {
	err := dd.client.Gauge("memory.usage", value, tags.ToStringArray(), 1.0)
	if err != nil {
		return fmt.Errorf("failed to register the gauge value: %v", err)
	}

	return nil
}

// Histogram defines a histogram value (e.g. latency)
func (dd *DatadogMetricHandler) Histogram(metric, suffix string, value float64, tags types.Tags) error {
	err := dd.client.Histogram(fmt.Sprintf("%s.%s", metric, suffix), value, tags.ToStringArray(), 1.0)
	if err != nil {
		return fmt.Errorf("failed to register the histogram value: %v", err)
	}

	return nil
}

// Distribution sends a distribution for percentis analisys
func (dd *DatadogMetricHandler) Distribution(metric string, value float64, tags types.Tags) error {
	err := dd.client.Distribution(fmt.Sprintf("%s.time", metric), value, tags.ToStringArray(), 1.0)
	if err != nil {
		return fmt.Errorf("failed to register the distribution value: %v", err)
	}

	return nil
}

// NewDatadogTag creates a new Datadog metric Tag object
func (dd *DatadogMetricHandler) NewDatadogTag(name, value string) types.Tag {
	return types.Tag{
		Name:  name,
		Value: value,
	}
}

// Event helps to register an event
func (dd *DatadogMetricHandler) Event(title, value string, tags types.Tags) error {
	err := dd.client.Event(&statsd.Event{
		Title: title,
		Text:  value,
		Tags:  tags.ToStringArray(),
	})
	if err != nil {
		return fmt.Errorf("failed to send a new event: %v", err)
	}

	return nil
}

// Close helps to close a client connection with Datadog server via StatsD
func (dd *DatadogMetricHandler) Close() error {
	return dd.client.Close()
}
