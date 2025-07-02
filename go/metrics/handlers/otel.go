package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/raywall/cloud-service-pack/go/metrics/types"
	"go.opentelemetry.io/otel"
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

// OtelMetricHandler is a struct that represents an OpenTelemetry client configuration
type OtelMetricHandler struct {
	tracer         trace.Tracer
	meter          metric.Meter
	counters       map[string]metric.Int64Counter
	gauges         map[string]metric.Float64Gauge
	histograms     map[string]metric.Float64Histogram
	upDownCounter  map[string]metric.Int64UpDownCounter
	serviceName    string
	serviceVersion string
}

// NewOtelMetricHandler creates a new OpenTelemetry client configuration
func NewOtelMetricHandler(serviceName, serviceVersion, otelEndpoint string) (*OtelMetricHandler, error) {
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

	return &OtelMetricHandler{
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

// Increment helps to increment a value in an OpenTelemetry custom metric
func (c *OtelMetricHandler) Increment(ctx context.Context, metricName string, value int64, tags types.Tags) error {
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

	counter.Add(ctx, value, metric.WithAttributes(tags.ToAttributes()...))
	return nil
}

// Gauge defines a static value in an OpenTelemetry custom metric
func (c *OtelMetricHandler) Gauge(ctx context.Context, metricName string, value float64, tags types.Tags) error {
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

	gauge.Record(ctx, value, metric.WithAttributes(tags.ToAttributes()...))
	return nil
}

// Histogram defines a histogram value (e.g. latency)
func (c *OtelMetricHandler) Histogram(ctx context.Context, metricName string, value float64, tags types.Tags) error {
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

	histogram.Record(ctx, value, metric.WithAttributes(tags.ToAttributes()...))
	return nil
}

// Close helps to close the OpenTelemetry providers
func (c *OtelMetricHandler) Close() error {
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

// UpDownCounter for values that can go up and down (like queue size, memory usage)
func (c *OtelMetricHandler) UpDownCounter(ctx context.Context, metricName string, value int64, tags types.Tags) error {
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

	upDownCounter.Add(ctx, value, metric.WithAttributes(tags.ToAttributes()...))
	return nil
}

// StartSpan creates a new span for tracing
func (c *OtelMetricHandler) StartSpan(ctx context.Context, name string, tags types.Tags) (context.Context, trace.Span) {
	ctx, span := c.tracer.Start(ctx, name)

	// Add attributes to span
	for _, tag := range tags {
		span.SetAttributes(tag.ToAttribute())
	}

	return ctx, span
}

// RecordEvent records a span event with attributes
func (c *OtelMetricHandler) RecordEvent(ctx context.Context, name string, tags types.Tags) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(tags.ToAttributes()...))
	}
}
