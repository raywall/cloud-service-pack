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

// OtelMetricHandler implements the MetricHandler interface for OpenTelemetry.
type OtelMetricHandler struct {
	tracer     trace.Tracer
	meter      metric.Meter
	counters   map[string]metric.Int64Counter
	gauges     map[string]metric.Float64Gauge
	histograms map[string]metric.Float64Histogram
}

// NewOtelMetricHandler creates a new OpenTelemetry client, including trace and metric providers.
// It establishes a GRPC connection to an OTEL collector.
func NewOtelMetricHandler(serviceName, serviceVersion, otelEndpoint string) (*OtelMetricHandler, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTEL resource: %w", err)
	}

	// Configure and create a trace exporter and provider.
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(otelEndpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create OTEL trace exporter: %w", err)
	}
	traceProvider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter), sdktrace.WithResource(res))
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Configure and create a metric exporter and provider.
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(otelEndpoint), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create OTEL metric exporter: %w", err)
	}
	metricProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(10*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(metricProvider)

	return &OtelMetricHandler{
		tracer:     otel.Tracer(serviceName),
		meter:      otel.Meter(serviceName),
		counters:   make(map[string]metric.Int64Counter),
		gauges:     make(map[string]metric.Float64Gauge),
		histograms: make(map[string]metric.Float64Histogram),
	}, nil
}

// Increment adds a value to a counter metric in OpenTelemetry.
// It creates the counter on-demand if it doesn't already exist.
func (c *OtelMetricHandler) Increment(ctx context.Context, metricName string, value int64, tags types.Tags) error {
	counter, exists := c.counters[metricName]
	if !exists {
		var err error
		counter, err = c.meter.Int64Counter(
			metricName,
			metric.WithDescription(fmt.Sprintf("Counter for %s", metricName)),
		)
		if err != nil {
			return fmt.Errorf("failed to create OTEL counter: %w", err)
		}
		c.counters[metricName] = counter
	}

	counter.Add(ctx, value, metric.WithAttributes(tags.ToAttributes()...))
	return nil
}

// Gauge records a value for a gauge metric in OpenTelemetry.
// It creates the gauge on-demand if it doesn't already exist.
func (c *OtelMetricHandler) Gauge(ctx context.Context, metricName string, value float64, tags types.Tags) error {
	gauge, exists := c.gauges[metricName]
	if !exists {
		var err error
		gauge, err = c.meter.Float64Gauge(
			metricName,
			metric.WithDescription(fmt.Sprintf("Gauge for %s", metricName)),
		)
		if err != nil {
			return fmt.Errorf("failed to create OTEL gauge: %w", err)
		}
		c.gauges[metricName] = gauge
	}

	gauge.Record(ctx, value, metric.WithAttributes(tags.ToAttributes()...))
	return nil
}

// Histogram records a value for a histogram metric in OpenTelemetry.
// It creates the histogram on-demand if it doesn't already exist.
func (c *OtelMetricHandler) Histogram(ctx context.Context, metricName string, value float64, tags types.Tags) error {
	histogram, exists := c.histograms[metricName]
	if !exists {
		var err error
		histogram, err = c.meter.Float64Histogram(
			metricName,
			metric.WithDescription(fmt.Sprintf("Histogram for %s", metricName)),
		)
		if err != nil {
			return fmt.Errorf("failed to create OTEL histogram: %w", err)
		}
		c.histograms[metricName] = histogram
	}

	histogram.Record(ctx, value, metric.WithAttributes(tags.ToAttributes()...))
	return nil
}

// Close gracefully shuts down the OpenTelemetry trace and meter providers,
// ensuring all buffered telemetry is exported.
func (c *OtelMetricHandler) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var shutdownErr error
	if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		if err := tp.Shutdown(ctx); err != nil {
			shutdownErr = fmt.Errorf("failed to shutdown trace provider: %w", err)
		}
	}
	if mp, ok := otel.GetMeterProvider().(*sdkmetric.MeterProvider); ok {
		if err := mp.Shutdown(ctx); err != nil {
			if shutdownErr != nil {
				shutdownErr = fmt.Errorf("%v; failed to shutdown meter provider: %w", shutdownErr, err)
			} else {
				shutdownErr = fmt.Errorf("failed to shutdown meter provider: %w", err)
			}
		}
	}
	return shutdownErr
}
