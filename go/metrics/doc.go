// Copyright 2025 Raywall Malheiros de Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package metrics provides a unified and simplified interface for instrumenting
applications with metrics, abstracting away the specifics of different monitoring
backends like Datadog and OpenTelemetry.

This library allows developers to easily record common metric types such as
counters, gauges, and histograms without needing to directly manage different
client libraries.

# Features

- Simple, clean API for recording metrics (`Increment`, `Gauge`, `Histogram`).
- Unified configuration for service and business context.
- Swappable backends (Datadog, OpenTelemetry) via configuration.
- Automatic tag conversion to backend-specific formats.

# Basic Usage

To get started, create a `ClientConfig` struct, then pass it to `NewMetricsClient`.
The client can then be used to send metrics. It is crucial to call `Close()` on the
client before the application terminates to ensure all buffered metrics are sent.

	// 1. Configure the client
	config := &metrics.ClientConfig{
		MetricCollectorType: metrics.DatadogCollector, // or metrics.OtelCollector
		Server: metrics.ServerConfig{
			Host:         "127.0.0.1",
			Port:         8125, // Port for Datadog Agent (statsd)
			MetricPrefix: "myapp",
		},
		Service: metrics.ServiceConfig{
			Name:    "my-awesome-service",
			Version: "1.0.2",
		},
	}

	// 2. Create a new client
	client, err := metrics.NewMetricsClient(config)
	if err != nil {
		log.Fatalf("Failed to create metrics client: %v", err)
	}
	defer client.Close() // Ensure metrics are flushed on exit

	// 3. Use the client to send metrics
	tags := types.Tags{
		{Name: "endpoint", Value: "/login"},
		{Name: "status", Value: 200},
	}
	client.Increment("http.requests.total", 1, tags)
	client.Histogram("http.request.duration_ms", 150.5, tags)
*/
package metrics
