package rules

import (
	"fmt"
	"strings"
)

// arrayOperation executa operações em arrays (MAX, MIN, AVERAGE, SUM, COUNT).
func arrayOperation(data map[string]interface{}, op, path string) (interface{}, error) {
    val, err := getValue(data, path)
    if err != nil {
        return nil, err
    }
    arr, ok := val.([]interface{})
    if !ok {
        return nil, fmt.Errorf("path %s is not an array", path)
    }
    if len(arr) == 0 {
        return 0.0, nil
    }
    switch strings.ToUpper(op) {
    case "COUNT":
        return float64(len(arr)), nil
    case "SUM", "AVERAGE", "MAX", "MIN":
        sum := 0.0
        min := float64(0)
        max := float64(0)
        for i, item := range arr {
            val, ok := item.(float64)
            if !ok {
                return nil, fmt.Errorf("invalid number in array: %v", item)
            }
            sum += val
            if i == 0 {
                min, max = val, val
            } else {
                if val < min {
                    min = val
                }
                if val > max {
                    max = val
                }
            }
        }
        switch strings.ToUpper(op) {
        case "SUM":
            return sum, nil
        case "AVERAGE":
            return sum / float64(len(arr)), nil
        case "MAX":
            return max, nil
        case "MIN":
            return min, nil
        }
    }
    return nil, fmt.Errorf("unsupported array operation: %s", op)
}

// getValue recupera um valor de um map[string]interface{} aninhado usando um caminho separado por pontos.
// Exemplo de caminho: "user.address.zipcode" ou "items[0].name"
func getValue(data map[string]interface{}, path string) (interface{}, error) {
    if strings.HasPrefix(path, "$.") {
        keys := strings.Split(strings.TrimPrefix(path, "$."), ".")
        current := data
        for i, key := range keys {
            if strings.Contains(key, "[") && strings.HasSuffix(key, "]") {
                arrayKey, index, subKey, err := parseArrayPath(key)
                if err != nil {
                    return nil, err
                }
                arr, ok := current[arrayKey].([]interface{})
                if !ok {
                    return nil, fmt.Errorf("path %s is not an array", arrayKey)
                }
                if index >= len(arr) {
                    return nil, fmt.Errorf("index %d out of bounds for %s", index, arrayKey)
                }
                current, ok = arr[index].(map[string]interface{})
                if !ok {
                    return nil, fmt.Errorf("array element at %s is not a map", path)
                }
                if subKey != "" {
                    if i == len(keys)-1 {
                        return current[subKey], nil
                    }
                    current, ok = current[subKey].(map[string]interface{})
                    if !ok {
                        return nil, fmt.Errorf("invalid subkey %s", subKey)
                    }
                } else if i == len(keys)-1 {
                    return arr[index], nil
                }
            } else {
                var ok bool
                current, ok = current[key].(map[string]interface{})
                if !ok {
                    if i == len(keys)-1 {
                        return data[key], nil
                    }
                    return nil, fmt.Errorf("invalid path: %s", path)
                }
            }
        }
        return current, nil
    }
 
	return parseLiteral(path)
}

// setValue define um valor em um map[string]interface{} aninhado usando um caminho separado por pontos.
// Cria mapas intermediários se eles não existirem.
func setValue(data map[string]interface{}, path string, value interface{}) error {
    if strings.HasPrefix(path, "$.") {
        keys := strings.Split(strings.TrimPrefix(path, "$."), ".")
        current := data
        for i, key := range keys[:len(keys)-1] {
            if strings.Contains(key, "[") && strings.HasSuffix(key, "]") {
                arrayKey, index, subKey, err := parseArrayPath(key)
                if err != nil {
                    return err
                }
                arr, ok := current[arrayKey].([]interface{})
                if !ok {
                    // Criar array se não existir
                    arr = make([]interface{}, index+1)
                    for j := range arr {
                        arr[j] = make(map[string]interface{})
                    }
                    current[arrayKey] = arr
                } else if index >= len(arr) {
                    // Expandir array se necessário
                    newArr := make([]interface{}, index+1)
                    copy(newArr, arr)
                    for j := len(arr); j <= index; j++ {
                        newArr[j] = make(map[string]interface{})
                    }
                    arr = newArr
                    current[arrayKey] = arr
                }
                current = arr[index].(map[string]interface{})
                if subKey != "" {
                    if i == len(keys)-2 {
                        current[subKey] = value
                    } else {
                        next, ok := current[subKey].(map[string]interface{})
                        if !ok {
                            next = make(map[string]interface{})
                            current[subKey] = next
                        }
                        current = next
                    }
                } else if i == len(keys)-2 {
                    arr[index] = value
                }
            } else {
                next, ok := current[key].(map[string]interface{})
                if !ok {
                    next = make(map[string]interface{})
                    current[key] = next
                }
                current = next
                if i == len(keys)-2 {
                    current[keys[len(keys)-1]] = value
                }
            }
        }
        return nil
    }
    return fmt.Errorf("invalid path: %s", path)
}