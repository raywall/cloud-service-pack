package rules

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var functionPayload = map[string]interface{}{
	"valor":        150.00,
	"limiteMaximo": 500.00,
	"moeda":        "BRL",
	"idade":        21,
	"tipo":         "adulto",
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

func TestPolicyGetValueFunction(t *testing.T) {
	all_rules := []struct {
		name     string
		rule     string
		expected interface{}
	}{
		{name: "first order string attribute", rule: `$.tipo`, expected: "adulto"},
		{name: "second order string attribute", rule: `$.endereco.estado`, expected: "SP"},
		{name: "third order string attribute", rule: `$.transacoes[0].id`, expected: "t1"},
		{name: "first order number attribute", rule: `$.idade`, expected: 21},
		{name: "second order number attribute", rule: `$.limites.maxTransacoes`, expected: 10},
		{name: "third order number attribute", rule: `$.transacoes[0].valor`, expected: 50.0},
		{name: "first order object attribute", rule: `$.limites`, expected: functionPayload["limites"]},
		{name: "second order object attribute", rule: `$.transacoes[0]`, expected: (functionPayload["transacoes"].([]interface{}))[0]},
	}

	t.Run("", func(t *testing.T) {
		for _, cenario := range all_rules {
			actual, err := getValue(functionPayload, cenario.rule)

			assert.NoError(t, err, "Não deveria haver erros")
			assert.Equal(t, actual, cenario.expected, fmt.Sprintf("O resultado está incorreto (%s)", cenario.name))
		}
	})
}
