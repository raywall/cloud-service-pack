package handlers

import "github.com/raywall/cloud-service-pack/go/metrics/types"

type MetricHandler interface {
	Increment(metric string, value int, tags types.Tags) error
	Gauge(metric string, value float64, tags types.Tags) error
	Histogram(metric, suffix string, value float64, tags types.Tags) error
	// Distribution(metric string, value float64, tags types.Tags) error
	// Event(title, value string, tags types.Tags) error
	Close() error
}
