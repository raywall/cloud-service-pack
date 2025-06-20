package rules

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	conditionPayload = map[string]interface{}{
		"valor":        150.00,
		"limiteMaximo": 500.00,
		"moeda":        "BRL",
		"idade":        21,
		"tipo":         "servico",
		"cliente": map[string]interface{}{
			"tipo": "premium",
		},
		"endereco": map[string]interface{}{
			"cep":    "01234-567",
			"cidade": "São Paulo",
			"estado": "SP",
		},
		"transacoes": []interface{}{
			map[string]interface{}{"id": "t1", "valor": 50.00},
			map[string]interface{}{"id": "t2", "valor": 75.00},
		},
		"limites": map[string]interface{}{
			"maxTransacoes": 10,
			"valorTotal":    1000.00,
		},
	}
)

func TestPolicyIfCondition(t *testing.T) {
	all_rules := []struct {
		name     string
		rule     string
		expected interface{}
	}{
		{name: "if menor", rule: `IF $.valor < 500 THEN SET $.result = 1`, expected: true},
		{name: "if menor ou igual", rule: `IF $.valor <= 150 THEN SET $.result = 1`, expected: true},
		{name: "if igual", rule: `IF $.valor == 150 THEN SET $.result = 1`, expected: true},
		{name: "if maior ou igual", rule: `IF $.valor >= 150 THEN SET $.result = 1`, expected: true},
		{name: "if maior", rule: `IF $.valor > 100 THEN SET $.result = 1`, expected: true},
		{name: "if diferente", rule: `IF $.valor != 200 THEN SET $.result = 1`, expected: true},
	}

	t.Run("", func(t *testing.T) {
		for _, cenario := range all_rules {
			value := NewRule(cenario.rule).IfCondition(conditionPayload)
			assert.Equal(t, value.Passed, cenario.expected, fmt.Sprintf("O valor está incorreto (%s)", cenario.name))
		}
	})
}

func TestPolicyOrCondition(t *testing.T) {
	all_rules := []struct {
		name     string
		rule     string
		executed bool
		passed   bool
	}{
		{name: "equal strings", rule: `$.moeda == "BRL" OR $.moeda == "USD"`, executed: true, passed: true},
		{name: "equal string and equal int", rule: `$.moeda == "EUR" OR $.idade == 21`, executed: true, passed: true},
		{name: "diferent string and equal int", rule: `$.moeda != "EUR" OR $.idade == 25`, executed: true, passed: true},
		{name: "equal string and equal int", rule: `$.moeda == "EUR" OR $.idade == 25`, executed: true, passed: false},
	}

	t.Run("", func(t *testing.T) {
		for _, cenario := range all_rules {
			value := NewRule(cenario.rule).OrCondition(conditionPayload)

			assert.Equal(t, value.Executed, cenario.executed, "%v (executed): expected %v, got %v", cenario.name, cenario.executed, value.Executed)
			assert.Equal(t, value.Passed, cenario.passed, "%v (passed): expected %v, got %v", cenario.name, cenario.passed, value.Passed)
		}
	})
}
