// Copyright 2025 Raywall Malheiros de Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"context"
	"errors"

	"github.com/raywall/cloud-service-pack/go/metrics/handlers"
	"github.com/raywall/cloud-service-pack/go/metrics/types"
)

var (
	// ErrHandlerNotFound is returned when the specified metric collector type is not supported.
	ErrHandlerNotFound = errors.New("metric handler was not found")
	// ErrConfigNotFound is returned when a nil configuration is provided to the client constructor.
	ErrConfigNotFound = errors.New("client configuration was not found")
)

type (
	// ServerConfig holds the configuration for the metrics collector server.
	ServerConfig struct {
		// Host is the hostname or IP address of the metrics collector.
		Host string `json:"host"`
		// Port is the port number used by the metrics collector.
		Port int `json:"port"`
		// MetricPrefix is a string prefixed to all metric names.
		MetricPrefix string `json:"metricPrefix"`
	}

	// ServiceConfig contains information about the service emitting the metrics.
	ServiceConfig struct {
		// Name of the service responsible for the metrics.
		Name string `json:"service"`
		// Version is the current version of the service.
		Version string `json:"version"`
	}

	// SolutionConfig provides business context for the metrics.
	SolutionConfig struct {
		// Team indicates the name of the team/squad responsible for this service.
		Team string `json:"team"`
		// Solution indicates the name of the solution or application.
		Solution string `json:"solution"`
		// Domain indicates the business domain of the solution (DDD).
		Domain string `json:"domain"`
		// Product indicates the name of the product this service belongs to.
		Product string `json:"product"`
		// CustomTags is an additional field to create custom tags not covered by the default object.
		CustomTags types.Tags `json:"tags"`
	}

	// ClientConfig holds all configuration needed to initialize a metrics client.
	ClientConfig struct {
		// MetricCollectorType indicates which metrics backend will be used (e.g., Datadog, OTEL).
		MetricCollectorType CollectorType
		// Server holds the server connection details.
		Server ServerConfig `json:"server"`
		// Service holds information about the emitting service.
		Service ServiceConfig `json:"service"`
		// Solution holds business context for the metrics.
		Solution SolutionConfig `json:"solution"`
	}

	// Client is the main metrics client used to send data to a collector.
	Client struct {
		ctx     context.Context
		config  *ClientConfig
		handler handlers.MetricHandler
	}

	// CollectorType defines the type of metrics collector backend.
	CollectorType string
)

const (
	// DatadogCollector represents the Datadog collector type.
	DatadogCollector CollectorType = "datadog"
	// OtelCollector represents the OpenTelemetry collector type.
	OtelCollector CollectorType = "otel"
)

// NewMetricsClient creates and initializes a new metrics client based on the provided configuration.
// It selects the appropriate backend handler (e.g., Datadog, OpenTelemetry) based on the
// MetricCollectorType specified in the config.
func NewMetricsClient(c *ClientConfig) (*Client, error) {
	var (
		handler handlers.MetricHandler
		err     error
	)

	if c == nil {
		return nil, ErrConfigNotFound
	}

	switch c.MetricCollectorType {
	case DatadogCollector:
		handler, err = handlers.NewDatadogMetricHandler(
			c.Server.MetricPrefix,
			c.Server.Host,
			c.Server.Port,
		)
		if err != nil {
			return nil, err
		}
	case OtelCollector:
		handler, err = handlers.NewOtelMetricHandler(
			c.Service.Name,
			c.Service.Version,
			c.Server.Host,
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrHandlerNotFound
	}

	return &Client{
		ctx:     context.Background(),
		config:  c,
		handler: handler,
	}, nil
}

// Increment sends a counter metric, which adds a value to a running total.
// The value is typically 1, but can be any integer.
func (c *Client) Increment(metric string, value int64, tags types.Tags) error {
	if c.handler == nil {
		return ErrHandlerNotFound
	}
	return c.handler.Increment(c.ctx, metric, value, tags)
}

// Gauge sends a gauge metric, which represents a single numerical value that can arbitrarily go up and down.
func (c *Client) Gauge(metric string, value float64, tags types.Tags) error {
	if c.handler == nil {
		return ErrHandlerNotFound
	}
	return c.handler.Gauge(c.ctx, metric, value, tags)
}

// Histogram sends a histogram metric, which samples observations (e.g., request durations)
// and calculates statistical distributions.
func (c *Client) Histogram(metric string, value float64, tags types.Tags) error {
	if c.handler == nil {
		return ErrHandlerNotFound
	}
	return c.handler.Histogram(c.ctx, metric, value, tags)
}

// Close gracefully shuts down the metrics client and flushes any buffered metrics.
// It should be called before the application exits.
func (c *Client) Close() error {
	if c.handler == nil {
		return ErrHandlerNotFound
	}
	return c.handler.Close()
}
