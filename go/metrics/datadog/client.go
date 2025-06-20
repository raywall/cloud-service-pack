package datadog

import (
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
)

// DatadogClient is a interface that represents a Datadog StatsD client
type DatadogClient interface {
	NewDatadogTag(name, value string) DatadogTag
	Increment(metric string, value int, tags DatadogTags) error
	Gauge(metric string, value float64, tags DatadogTags) error
	Histogram(metric, suffix string, value float64, tags DatadogTags) error
	Distribution(metric string, value float64, tags DatadogTags) error
	Event(title, value string, tags DatadogTags) error
	Close() error
}

// New creates a new Datadog client configuration
func New(metric_prefix, server_host string, server_port int) (DatadogClient, error) {
	prefix := ""

	if metric_prefix != "" {
		prefix = fmt.Sprintf("%s.", metric_prefix)
	}

	if _client, err := statsd.New(fmt.Sprintf("%s:%d", server_host, server_port), statsd.WithNamespace(prefix)); err != nil {
		return nil, fmt.Errorf("failed to create a StatsD client for DD: %v", err)

	} else {
		return &datadogClient{
			client: _client,
		}, nil
	}
}

// NewDatadogTag creates a new Datadog metric Tag object
func (dd *datadogClient) NewDatadogTag(name, value string) DatadogTag {
	return DatadogTag{
		Name:  name,
		Value: value,
	}
}

// Increment helps to increment a value in a Datadog custom metric
func (dd *datadogClient) Increment(metric string, value int, tags DatadogTags) error {
	err := dd.client.Incr(fmt.Sprintf("%s.count", metric), tags.ToStringArray(), 1.0)
	if err != nil {
		return fmt.Errorf("failed to increment the counter: %v", err)
	}

	return nil
}

// Gauge defines a static value in a Datadog custom metric
func (dd *datadogClient) Gauge(metric string, value float64, tags DatadogTags) error {
	err := dd.client.Gauge("memory.usage", value, tags.ToStringArray(), 1.0)
	if err != nil {
		fmt.Errorf("failed to register the gauge value: %v", err)
	}

	return nil
}

// Histogram defines a histogram value (e.g. latency)
func (dd *datadogClient) Histogram(metric, suffix string, value float64, tags DatadogTags) error {
	err := dd.client.Histogram(fmt.Sprintf("%s.%s", metric, suffix), value, tags.ToStringArray(), 1.0)
	if err != nil {
		return fmt.Errorf("failed to register the histogram value: %v", err)
	}

	return nil
}

// Distribution sends a distribution for percentis analisys
func (dd *datadogClient) Distribution(metric string, value float64, tags DatadogTags) error {
	err := dd.client.Distribution(fmt.Sprintf("%s.time", metric), value, tags.ToStringArray(), 1.0)
	if err != nil {
		return fmt.Errorf("failed to register the distribution value: %v", err)
	}

	return nil
}

// Event helps to register an event
func (dd *datadogClient) Event(title, value string, tags DatadogTags) error {
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
func (dd *datadogClient) Close() error {
	return dd.client.Close()
}

// ToStringArray convert an array of DatadogTag into a string array
func (tags *DatadogTags) ToStringArray() []string {
	result := make([]string, 0)

	for _, tag := range *tags {
		result = append(result, fmt.Sprintf("%s:%v", tag.Name, tag.Value))
	}

	return result
}
