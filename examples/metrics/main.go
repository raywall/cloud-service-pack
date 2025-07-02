package main

import (
	"log"
	"time"

	"github.com/raywall/cloud-service-pack/go/metrics"
	"github.com/raywall/cloud-service-pack/go/metrics/types"
)

func main() {
	// =========================================================================
	// Example 1: Using the Datadog Collector
	// =========================================================================
	log.Println("--- Running Datadog Example ---")
	runDatadogExample()

	log.Println()

	// =========================================================================
	// Example 2: Using the OpenTelemetry Collector
	// NOTE: Requires an OTEL Collector running and listening on localhost:4317
	// =========================================================================
	log.Println("--- Running OpenTelemetry Example ---")
	runOtelExample()

	log.Println("\nExamples finished. Check your monitoring backend.")
}

func runDatadogExample() {
	// Configure the client for Datadog.
	// Assumes a Datadog agent is running locally and listening on port 8125.
	config := &metrics.ClientConfig{
		MetricCollectorType: metrics.DatadogCollector,
		Server: metrics.ServerConfig{
			Host:         "127.0.0.1",
			Port:         8125,
			MetricPrefix: "myshop",
		},
		Service: metrics.ServiceConfig{
			Name:    "checkout-service",
			Version: "1.0.0",
		},
	}

	// Create a new client
	client, err := metrics.NewMetricsClient(config)
	if err != nil {
		log.Fatalf("Failed to create Datadog client: %v", err)
	}
	// Use defer to ensure Close is called to flush metrics.
	defer client.Close()

	// Define some tags
	tags := types.Tags{
		{Name: "payment_method", Value: "credit_card"},
		{Name: "country", Value: "br"},
	}

	// Send some metrics
	log.Println("Sending Datadog metrics...")
	client.Increment("orders.placed", 1, tags)
	client.Gauge("items.in_cart", 5, tags)
	client.Histogram("order.value", 250.75, tags)

	// Keep the app running for a moment to ensure metrics are sent via UDP
	time.Sleep(2 * time.Second)
	log.Println("Datadog metrics sent.")
}

func runOtelExample() {
	// Configure the client for OpenTelemetry.
	// Assumes an OTEL collector is running and listening for GRPC on localhost:4317.
	config := &metrics.ClientConfig{
		MetricCollectorType: metrics.OtelCollector,
		Server: metrics.ServerConfig{
			Host: "localhost:4317", // For OTEL, host includes port
		},
		Service: metrics.ServiceConfig{
			Name:    "inventory-service",
			Version: "2.1.0",
		},
	}

	// Create a new client
	client, err := metrics.NewMetricsClient(config)
	if err != nil {
		log.Fatalf("Failed to create OTEL client: %v", err)
	}
	// Use defer to ensure Close is called to flush metrics.
	defer client.Close()

	// Define some tags
	tags := types.Tags{
		{Name: "product_id", Value: 12345},
		{Name: "warehouse", Value: "sp-01"},
		{Name: "in_stock", Value: true},
	}

	// Send some metrics
	log.Println("Sending OpenTelemetry metrics...")
	client.Increment("products.viewed", 1, tags)
	client.Gauge("stock.level", 150, tags)
	client.Histogram("api.latency.ms", 45.6, tags)

	// For OTEL, the provider pushes metrics periodically. We wait here to
	// allow at least one push cycle to complete before closing.
	log.Println("Waiting for OTEL to export metrics (12 seconds)...")
	time.Sleep(12 * time.Second)
	log.Println("OpenTelemetry metrics should be exported.")
}
