package otel

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// OtelTag is a structure that represents an OpenTelemetry metric Tag/Attribute
type OtelTag struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// OtelTags is a struct that represents an array of OpenTelemetry metric Tags
type OtelTags []OtelTag

// otelClient is a struct that represents an OpenTelemetry client configuration
type otelClient struct {
	tracer         trace.Tracer
	meter          metric.Meter
	counters       map[string]metric.Int64Counter
	gauges         map[string]metric.Float64Gauge
	histograms     map[string]metric.Float64Histogram
	upDownCounter  map[string]metric.Int64UpDownCounter
	serviceName    string
	serviceVersion string
}
