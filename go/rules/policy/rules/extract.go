package rules

import (
	"strings"
)

func extractExpression(valueStr string) (string, bool) {
	valueStr = strings.TrimSpace(valueStr)
	if strings.HasPrefix(valueStr, "EXP(") && strings.HasSuffix(valueStr, ")") {
		return valueStr[4 : len(valueStr)-1], true
	}
	return "", false
}
