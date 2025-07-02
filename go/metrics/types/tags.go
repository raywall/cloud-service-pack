package types

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

// Tag represents a single key-value pair used to add dimensions to a metric.
type Tag struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// Tags is a slice of Tag objects.
type Tags []Tag

// ToStringArray converts a slice of Tags into a string array formatted for Datadog ("key:value").
func (tags *Tags) ToStringArray() []string {
	result := make([]string, 0, len(*tags))
	for _, tag := range *tags {
		result = append(result, fmt.Sprintf("%s:%v", tag.Name, tag.Value))
	}
	return result
}

// ToAttributes converts a slice of Tags to an array of OpenTelemetry KeyValue attributes.
func (t *Tags) ToAttributes() []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(*t))
	for _, tag := range *t {
		attrs = append(attrs, tag.ToAttribute())
	}
	return attrs
}

// ToAttribute converts a single Tag to an OpenTelemetry KeyValue attribute,
// automatically detecting the value type.
func (t *Tag) ToAttribute() attribute.KeyValue {
	switch v := t.Value.(type) {
	case string:
		return attribute.String(t.Name, v)
	case int:
		return attribute.Int(t.Name, v)
	case int64:
		return attribute.Int64(t.Name, v)
	case float64:
		return attribute.Float64(t.Name, v)
	case bool:
		return attribute.Bool(t.Name, v)
	default:
		// Fallback for other types
		return attribute.String(t.Name, fmt.Sprintf("%v", v))
	}
}
