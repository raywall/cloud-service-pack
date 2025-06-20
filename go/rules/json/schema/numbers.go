package schema

import (
	"encoding/json"
	"errors"
	"strconv"
)

func isNumber(data interface{}) bool {
	switch data.(type) {
	case float64:
		return true
	case float32:
		return true
	case int, int32, int64:
		return true
	default:
		return false
	}
}

func isInteger(data interface{}) bool {
	// JSON unmarshals numbers as float64
	f, ok := data.(float64)
	if !ok {
		return false
	}
	return f == float64(int64(f))
}

func toInt(v interface{}) (int, error) {
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	case int64:
		return int(n), nil
	case string:
		return strconv.Atoi(n)
	default:
		return 0, errors.New("não é número inteiro")
	}
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
