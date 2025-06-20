package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	rulePayload = map[string]interface{}{
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

func TestPolicySetCondition(t *testing.T) {
	all_rules := []struct {
		name      string
		rule      string
		expected  interface{}
		executed  bool
		passed    bool
		attribute string
	}{
		{name: "set float value", rule: `SET $.desconto = 15.0`, expected: 15.0, executed: true, passed: true, attribute: `desconto`},
		{name: "set int value", rule: `SET $.idade = 22`, expected: 22.0, executed: true, passed: true, attribute: `idade`},
	}

	t.Run("", func(t *testing.T) {
		for _, cenario := range all_rules {
			value := NewRule(cenario.rule).SetValue(rulePayload)

			assert.Equal(t, rulePayload[cenario.attribute], cenario.expected, "%v: expected %v, got %v", cenario.name, cenario.expected, rulePayload[cenario.attribute])
			assert.Equal(t, value.Executed, cenario.executed, "%v: expected %v, got %v", cenario.name, cenario.expected, value.Executed)
			assert.Equal(t, value.Passed, cenario.passed, "%v: expected %v, got %v", cenario.name, cenario.passed, value.Passed)
		}
	})
}
