package rules

import (
	"reflect"
	"strings"
	"fmt"
)

func compareEquals(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	aNum, aIsNum := convertToFloat64(a)
	bNum, bIsNum := convertToFloat64(b)
	if aIsNum && bIsNum {
		return aNum == bNum
	}

	// Se não são ambos números, compara os tipos e depois os valores
	// Isso é importante para "string" == "string" e bool == bool
	// mas também para evitar que 1 (int) seja igual a "1" (string) sem conversão explícita
	if reflect.TypeOf(a) == reflect.TypeOf(b) {
		return reflect.DeepEqual(a, b) // DeepEqual lida com strings, bools, etc.
	}

	// Se os tipos são diferentes e não são ambos convertíveis para float64, considera-os diferentes.
	// Ex: 10 (número) != "10" (string)
	return false
}

// isOperatorProtected verifica se um operador (como " OR " ou "*") está dentro de aspas,
// o que significaria que faz parte de um valor literal de string e não um operador real.
// Esta é uma verificação simples e pode não cobrir todos os casos de escape.
func isOperatorProtected(rulePart string, operator string) bool {
	opIndex := strings.Index(rulePart, operator)
	if opIndex == -1 {
		return false
	}
	inSingleQuote, inDoubleQuote := false, false
	for i := 0; i < opIndex; i++ {
		char := rulePart[i]
		if char == '\'' {
			inSingleQuote = !inSingleQuote
		}
		if char == '"' {
			inDoubleQuote = !inDoubleQuote
		}
	}
	return inSingleQuote || inDoubleQuote
}

func compareNumbers(left, right interface{}, op string) (bool, error) {
    l, lok := left.(float64)
    r, rok := right.(float64)
    if !lok || !rok {
        return false, fmt.Errorf("invalid number comparison: %v %s %v", left, op, right)
    }
    switch op {
    case ">=":
        return l >= r, nil
    case ">":
        return l > r, nil
    case "<=":
        return l <= r, nil
    case "<":
        return l < r, nil
    default:
        return false, fmt.Errorf("unsupported number operator: %s", op)
    }
}

func inArray(val, arr interface{}) (bool, error) {
    arrVal, ok := arr.([]interface{})
    if !ok {
        return false, fmt.Errorf("invalid array for IN: %v", arr)
    }
    for _, item := range arrVal {
        if reflect.DeepEqual(val, item) {
            return true, nil
        }
    }
    return false, nil
}