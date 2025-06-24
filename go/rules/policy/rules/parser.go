package rules

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// parseArrayPath analisa caminhos de array (ex.: teste[0].nome).
func parseArrayPath(key string) (string, int, string, error) {
	parts := strings.SplitN(key, "[", 2)
	arrayKey := parts[0]
	rest := strings.TrimSuffix(parts[1], "]")
	if strings.Contains(rest, ".") {
		subParts := strings.SplitN(rest, ".", 2)
		index, err := strconv.Atoi(subParts[0])
		if err != nil {
			return "", 0, "", fmt.Errorf("invalid array index: %s", subParts[0])
		}
		return arrayKey, index, subParts[1], nil
	}
	index, err := strconv.Atoi(rest)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid array index: %s", rest)
	}
	return arrayKey, index, "", nil
}

// parseOperand converte uma string de operando (literal ou caminho $) em float64.
func parseOperand(operandStr string, data map[string]interface{}) (float64, error) {
	operandStr = strings.TrimSpace(operandStr)
	if strings.HasPrefix(operandStr, "$.") {
		val, err := getValue(data, operandStr)
		if err != nil {
			return 0, fmt.Errorf("falha ao obter valor do caminho do operando '%s': %v", operandStr, err)
		}
		num, ok := convertToFloat64(val)
		if !ok {
			return 0, fmt.Errorf("operando do caminho '%s' (valor: %v, tipo: %T) não é um número válido", operandStr, val, val)
		}
		return num, nil
	}
	// É um literal
	num, ok := convertToFloat64(operandStr)
	if !ok {
		return 0, fmt.Errorf("operando literal '%s' não é um número válido", operandStr)
	}
	return num, nil
}

func convertToFloat64(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false // Não pode converter nil para float64
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f, true
		}
	}
	return 0, false
}

// Funções auxiliares (mantidas do código anterior, com adição de matches)
func parseValue(val string, reference interface{}) (interface{}, error) {
	if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
		var arr []interface{}
		if err := json.Unmarshal([]byte(val), &arr); err != nil {
			return nil, fmt.Errorf("invalid array: %s", val)
		}
		return arr, nil
	}
	switch reference.(type) {
	case float64:
		return parseFloat(val)
	case string:
		return strings.Trim(val, `"'`), nil
	default:
		return val, nil
	}
}

func parseFloat(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}

func parseLiteral(val string) (interface{}, error) {
	if val == "null" {
		return nil, nil
	}
	if i, err := strconv.Atoi(val); err == nil {
		return float64(i), nil
	}
	if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f, nil
	}
	if b, err := strconv.ParseBool(val); err == nil {
		return b, nil
	}
	return strings.Trim(val, `"'`), nil
}
