package rules

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	simpleConditionPayload = map[string]interface{}{
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

func TestPolicySimpleCondition(t *testing.T) {
	all_rules := []struct {
		name     string
		rule     string
		expected bool
	}{
		{name: "maior ou igual", rule: `$.idade >= 18`, expected: true},
		{name: "igual", rule: `$.tipo == "adulto"`, expected: false},
		{name: "maior", rule: `$.valor > 0`, expected: true},
		{name: "menor ou igual", rule: `$.valor <= $.limiteMaximo`, expected: true},
		{name: "diferente", rule: `$.moeda != "BRL"`, expected: false},
		{name: "diferente de nulo", rule: `$.endereco.cep != null`, expected: true},
	}

	t.Run("", func(t *testing.T) {
		for _, cenario := range all_rules {
			actual, _, err := EvaluateRule(cenario.rule, simpleConditionPayload)

			assert.NoError(t, err, "Não deveria haver erros")
			assert.Equal(t, actual, cenario.expected, fmt.Sprintf("O resultado está incorreto (%s)", cenario.name))
		}
	})
}
