package types

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

type Tag struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type Tags []Tag

// ToStringArray convert an array of DatadogTag into a string array
func (tags *Tags) ToStringArray() []string {
	result := make([]string, 0)

	for _, tag := range *tags {
		result = append(result, fmt.Sprintf("%s:%v", tag.Name, tag.Value))
	}

	return result
}

// ToAttributes converts OtelTags to OpenTelemetry attributes
func (t *Tags) ToAttributes() []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(*t))
	for _, tag := range *t {
		attrs = append(attrs, tag.ToAttribute())
	}
	return attrs
}

// ToAttribute converts a single OtelTag to an OpenTelemetry attribute
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
		return attribute.String(t.Name, fmt.Sprintf("%v", v))
	}
}
