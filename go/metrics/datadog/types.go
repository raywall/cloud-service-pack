package datadog

import (
	"github.com/DataDog/datadog-go/v5/statsd"
)

// DatadogTag is a structure that represents a Datadog metric Tag
type DatadogTag struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// DatadogTags is a struct that represents an array of Datadog metric Tag
type DatadogTags []DatadogTag

// datadogClient is a struct that represents a Datadog client configuration
type datadogClient struct {
	client *statsd.Client
}
