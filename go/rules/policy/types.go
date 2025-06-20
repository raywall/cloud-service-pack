package policy

import (
	"github.com/raywall/cloud-policy-serializer/pkg/policy/rules"
)

// PolicyDefinition representa uma única política com suas regras.
// As regras são strings na linguagem de política customizada.
type PolicyDefinition struct {
	Name  string
	Rules []string
}

// PolicyExecutionResult armazena o resultado da execução de uma política.
type PolicyExecutionResult struct {
	PolicyName  string
	Passed      bool
	Error       error // Mensagem de erro se a avaliação falhou ou a regra não foi cumprida
	RuleResults []rules.RuleExecutionResult
}
