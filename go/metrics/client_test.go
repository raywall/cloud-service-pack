package metrics

import (
	"context"
	"errors"
	"testing"

	"github.com/raywall/cloud-service-pack/go/metrics/types"
	"github.com/stretchr/testify/assert"
)

// mockMetricHandler is a mock implementation of the MetricHandler interface for testing.
type mockMetricHandler struct {
	incrementCalled bool
	gaugeCalled     bool
	histogramCalled bool
	closeCalled     bool
	lastMetric      string
	lastTags        types.Tags
	shouldFail      bool
}

func (m *mockMetricHandler) Increment(ctx context.Context, metric string, value int64, tags types.Tags) error {
	m.incrementCalled = true
	m.lastMetric = metric
	m.lastTags = tags
	if m.shouldFail {
		return errors.New("mock increment error")
	}
	return nil
}

func (m *mockMetricHandler) Gauge(ctx context.Context, metric string, value float64, tags types.Tags) error {
	m.gaugeCalled = true
	m.lastMetric = metric
	m.lastTags = tags
	if m.shouldFail {
		return errors.New("mock gauge error")
	}
	return nil
}

func (m *mockMetricHandler) Histogram(ctx context.Context, metric string, value float64, tags types.Tags) error {
	m.histogramCalled = true
	m.lastMetric = metric
	m.lastTags = tags
	if m.shouldFail {
		return errors.New("mock histogram error")
	}
	return nil
}

func (m *mockMetricHandler) Close() error {
	m.closeCalled = true
	if m.shouldFail {
		return errors.New("mock close error")
	}
	return nil
}

func TestNewMetricsClient(t *testing.T) {
	t.Run("should return error when config is nil", func(t *testing.T) {
		client, err := NewMetricsClient(nil)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, ErrConfigNotFound, err)
	})

	t.Run("should return error for unknown collector type", func(t *testing.T) {
		config := &ClientConfig{
			MetricCollectorType: "unknown",
		}
		client, err := NewMetricsClient(config)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, ErrHandlerNotFound, err)
	})

	// Note: Testing successful creation of Datadog and Otel handlers would require
	// either a running collector or mocking the handler constructors, which is more complex.
	// For simplicity, we focus on the client's logic here.
}

func TestClientOperations(t *testing.T) {
	mockHandler := &mockMetricHandler{}
	client := &Client{
		ctx:     context.Background(),
		config:  &ClientConfig{},
		handler: mockHandler,
	}
	tags := types.Tags{{Name: "test", Value: "true"}}

	t.Run("Increment", func(t *testing.T) {
		err := client.Increment("test.metric", 1, tags)
		assert.NoError(t, err)
		assert.True(t, mockHandler.incrementCalled)
		assert.Equal(t, "test.metric", mockHandler.lastMetric)
		assert.Equal(t, tags, mockHandler.lastTags)
	})

	t.Run("Gauge", func(t *testing.T) {
		err := client.Gauge("test.gauge", 99.9, tags)
		assert.NoError(t, err)
		assert.True(t, mockHandler.gaugeCalled)
		assert.Equal(t, "test.gauge", mockHandler.lastMetric)
		assert.Equal(t, tags, mockHandler.lastTags)
	})

	t.Run("Histogram", func(t *testing.T) {
		err := client.Histogram("test.histogram", 123.45, tags)
		assert.NoError(t, err)
		assert.True(t, mockHandler.histogramCalled)
		assert.Equal(t, "test.histogram", mockHandler.lastMetric)
		assert.Equal(t, tags, mockHandler.lastTags)
	})

	t.Run("Close", func(t *testing.T) {
		err := client.Close()
		assert.NoError(t, err)
		assert.True(t, mockHandler.closeCalled)
	})
}

func TestClientOperations_NoHandler(t *testing.T) {
	client := &Client{
		handler: nil, // Explicitly set handler to nil
	}
	tags := types.Tags{}

	assert.Equal(t, ErrHandlerNotFound, client.Increment("metric", 1, tags))
	assert.Equal(t, ErrHandlerNotFound, client.Gauge("metric", 1.0, tags))
	assert.Equal(t, ErrHandlerNotFound, client.Histogram("metric", 1.0, tags))
	assert.Equal(t, ErrHandlerNotFound, client.Close())
}
