package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// OtelClient is an interface that represents an OpenTelemetry client
type OtelClient interface {
	NewOtelTag(name string, value interface{}) OtelTag
	Increment(ctx context.Context, metric string, value int64, tags OtelTags) error
	Gauge(ctx context.Context, metric string, value float64, tags OtelTags) error
	Histogram(ctx context.Context, metric string, value float64, tags OtelTags) error
	UpDownCounter(ctx context.Context, metric string, value int64, tags OtelTags) error
	StartSpan(ctx context.Context, name string, tags OtelTags) (context.Context, trace.Span)
	RecordEvent(ctx context.Context, name string, tags OtelTags)
	Close() error
}

// New creates a new OpenTelemetry client configuration
func New(serviceName, serviceVersion, otelEndpoint string) (OtelClient, error) {
	ctx := context.Background()

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %v", err)
	}

	// Configure trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(otelEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %v", err)
	}

	// Configure trace provider
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Configure metric exporter
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(otelEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %v", err)
	}

	// Configure metric provider
	metricProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(10*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(metricProvider)

	// Create tracer and meter
	tracer := otel.Tracer(serviceName)
	meter := otel.Meter(serviceName)

	return &otelClient{
		tracer:         tracer,
		meter:          meter,
		counters:       make(map[string]metric.Int64Counter),
		gauges:         make(map[string]metric.Float64Gauge),
		histograms:     make(map[string]metric.Float64Histogram),
		upDownCounter:  make(map[string]metric.Int64UpDownCounter),
		serviceName:    serviceName,
		serviceVersion: serviceVersion,
	}, nil
}

// NewOtelTag creates a new OpenTelemetry metric Tag object
func (c *otelClient) NewOtelTag(name string, value interface{}) OtelTag {
	return OtelTag{
		Name:  name,
		Value: value,
	}
}

// Increment helps to increment a value in an OpenTelemetry custom metric
func (c *otelClient) Increment(ctx context.Context, metricName string, value int64, tags OtelTags) error {
	counter, exists := c.counters[metricName]
	if !exists {
		var err error
		counter, err = c.meter.Int64Counter(
			metricName,
			metric.WithDescription(fmt.Sprintf("Counter metric for %s", metricName)),
		)
		if err != nil {
			return fmt.Errorf("failed to create counter: %v", err)
		}
		c.counters[metricName] = counter
	}

	counter.Add(ctx, value, metric.WithAttributes(c.tagsToAttributes(tags)...))
	return nil
}

// Gauge defines a static value in an OpenTelemetry custom metric
func (c *otelClient) Gauge(ctx context.Context, metricName string, value float64, tags OtelTags) error {
	gauge, exists := c.gauges[metricName]
	if !exists {
		var err error
		gauge, err = c.meter.Float64Gauge(
			metricName,
			metric.WithDescription(fmt.Sprintf("Gauge metric for %s", metricName)),
		)
		if err != nil {
			return fmt.Errorf("failed to create gauge: %v", err)
		}
		c.gauges[metricName] = gauge
	}

	gauge.Record(ctx, value, metric.WithAttributes(c.tagsToAttributes(tags)...))
	return nil
}

// Histogram defines a histogram value (e.g. latency)
func (c *otelClient) Histogram(ctx context.Context, metricName string, value float64, tags OtelTags) error {
	histogram, exists := c.histograms[metricName]
	if !exists {
		var err error
		histogram, err = c.meter.Float64Histogram(
			metricName,
			metric.WithDescription(fmt.Sprintf("Histogram metric for %s", metricName)),
		)
		if err != nil {
			return fmt.Errorf("failed to create histogram: %v", err)
		}
		c.histograms[metricName] = histogram
	}

	histogram.Record(ctx, value, metric.WithAttributes(c.tagsToAttributes(tags)...))
	return nil
}

// UpDownCounter for values that can go up and down (like queue size, memory usage)
func (c *otelClient) UpDownCounter(ctx context.Context, metricName string, value int64, tags OtelTags) error {
	upDownCounter, exists := c.upDownCounter[metricName]
	if !exists {
		var err error
		upDownCounter, err = c.meter.Int64UpDownCounter(
			metricName,
			metric.WithDescription(fmt.Sprintf("UpDownCounter metric for %s", metricName)),
		)
		if err != nil {
			return fmt.Errorf("failed to create up-down counter: %v", err)
		}
		c.upDownCounter[metricName] = upDownCounter
	}

	upDownCounter.Add(ctx, value, metric.WithAttributes(c.tagsToAttributes(tags)...))
	return nil
}

// StartSpan creates a new span for tracing
func (c *otelClient) StartSpan(ctx context.Context, name string, tags OtelTags) (context.Context, trace.Span) {
	ctx, span := c.tracer.Start(ctx, name)

	// Add attributes to span
	for _, tag := range tags {
		span.SetAttributes(c.tagToAttribute(tag))
	}

	return ctx, span
}

// RecordEvent records a span event with attributes
func (c *otelClient) RecordEvent(ctx context.Context, name string, tags OtelTags) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(c.tagsToAttributes(tags)...))
	}
}

// Close helps to close the OpenTelemetry providers
func (c *otelClient) Close() error {
	ctx := context.Background()

	// Get providers and shutdown
	if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		if err := tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown trace provider: %v", err)
		}
	}

	if mp, ok := otel.GetMeterProvider().(*sdkmetric.MeterProvider); ok {
		if err := mp.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown meter provider: %v", err)
		}
	}

	return nil
}

// tagsToAttributes converts OtelTags to OpenTelemetry attributes
func (c *otelClient) tagsToAttributes(tags OtelTags) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(tags))
	for _, tag := range tags {
		attrs = append(attrs, c.tagToAttribute(tag))
	}
	return attrs
}

// tagToAttribute converts a single OtelTag to an OpenTelemetry attribute
func (c *otelClient) tagToAttribute(tag OtelTag) attribute.KeyValue {
	switch v := tag.Value.(type) {
	case string:
		return attribute.String(tag.Name, v)
	case int:
		return attribute.Int(tag.Name, v)
	case int64:
		return attribute.Int64(tag.Name, v)
	case float64:
		return attribute.Float64(tag.Name, v)
	case bool:
		return attribute.Bool(tag.Name, v)
	default:
		return attribute.String(tag.Name, fmt.Sprintf("%v", v))
	}
}
