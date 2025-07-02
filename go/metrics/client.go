package metrics

import (
	"errors"

	"github.com/raywall/cloud-service-pack/go/metrics/handlers"
	"github.com/raywall/cloud-service-pack/go/metrics/types"
)

var (
	ErrHandlerNotFound = errors.New("metric handler was not found")
)

type Client struct {
	handler handlers.MetricHandler
}

func NewMetricsClient(handler handlers.MetricHandler) *Client {
	return &Client{
		handler: handler,
	}
}

func (c *Client) Increment(metric string, value int, tags types.Tags) error {
	if c.handler == nil {
		return ErrHandlerNotFound
	}
	return c.handler.Increment(metric, value, tags)
}

func (c *Client) Gauge(metric string, value float64, tags types.Tags) error {
	if c.handler == nil {
		return ErrHandlerNotFound
	}
	return c.handler.Gauge(metric, value, tags)
}

func (c *Client) Histogram(metric, suffix string, value float64, tags types.Tags) error {
	if c.handler == nil {
		return ErrHandlerNotFound
	}
	return c.handler.Histogram(metric, suffix, value, tags)
}

func (c *Client) Close() error {
	if c.handler == nil {
		return ErrHandlerNotFound
	}
	return c.handler.Close()
}
