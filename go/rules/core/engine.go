package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/raywall/cloud-policy-serializer/pkg/json/schema"
	"github.com/raywall/cloud-policy-serializer/pkg/policy"
	"github.com/raywall/cloud-policy-serializer/pkg/policy/rules"
)

// ExecutePolicies executa as políticas especificadas contra os dados.
func (ec *EngineContext) ExecutePolicies(data map[string]interface{}, policyNames []string) ([]policy.PolicyExecutionResult, bool) {
	var results []policy.PolicyExecutionResult
	allPassedOverall := true

	for _, policyName := range policyNames {
		policyDef, exists := ec.Policies[policyName]
		if !exists {
			results = append(results, policy.PolicyExecutionResult{
				PolicyName: policyName,
				Passed:     false,
				Error:      fmt.Errorf("política '%s' não definida", policyName),
			})
			allPassedOverall = false
			continue
		}

		currentPolicyAllRulesPassed := true
		var currentPolicyFirstError error
		var ruleResultsForThisPolicy []rules.RuleExecutionResult

		for _, ruleStr := range policyDef.Rules {
			passedThisRule, details, errThisRule := rules.EvaluateRule(ruleStr, data)

			ruleExecRes := rules.RuleExecutionResult{
				Rule:    ruleStr,
				Passed:  passedThisRule,
				Details: details,
			}
			if errThisRule != nil {
				ruleExecRes.Details += " (Erro: " + errThisRule.Error() + ")"
			}
			ruleResultsForThisPolicy = append(ruleResultsForThisPolicy, ruleExecRes)

			isSetOrIf := strings.HasPrefix(strings.TrimSpace(ruleStr), "SET ") || strings.HasPrefix(strings.TrimSpace(ruleStr), "IF ")

			if errThisRule != nil {
				currentPolicyAllRulesPassed = false
				if currentPolicyFirstError == nil { // Pega o primeiro erro da política
					currentPolicyFirstError = fmt.Errorf("erro ao executar regra '%s': %v. Detalhes: %s", ruleStr, errThisRule, details)
				}
				// Não necessariamente para aqui, pode continuar para registrar outras falhas de regras se desejado,
				// mas a política já falhou. Para o comportamento de "parar na primeira falha da política":
				break
			}

			// Para regras de condição (não SET/IF), 'passedThisRule == false' significa falha na condição.
			if !isSetOrIf && !passedThisRule {
				currentPolicyAllRulesPassed = false
				if currentPolicyFirstError == nil {
					currentPolicyFirstError = fmt.Errorf("condição da regra não atendida: '%s'. Detalhes: %s", ruleStr, details)
				}
				// Parar na primeira condição falha dentro da política
				break
			}
			// Para SET/IF, 'passedThisRule == false' (e errThisRule != nil) indica um erro na execução da ação SET/IF.
			// Se errThisRule == nil, então a operação SET/IF foi bem-sucedida ou a condição IF foi falsa (o que não é uma falha).
		}

		results = append(results, policy.PolicyExecutionResult{
			PolicyName:  policyName,
			Passed:      currentPolicyAllRulesPassed,
			Error:       currentPolicyFirstError,
			RuleResults: ruleResultsForThisPolicy,
		})
		if !currentPolicyAllRulesPassed {
			allPassedOverall = false
		}
	}
	return results, allPassedOverall
}

// NewEngineContext cria um novo contexto de motor.
func NewEngineContext(reqSchema, respSchema *schema.Schema, policiesConfig map[string]policy.PolicyDefinition, inputType string) *EngineContext {
	return &EngineContext{
		RequestSchema:  reqSchema,
		ResponseSchema: respSchema,
		Policies:       policiesConfig,
		InputType:      inputType,
	}
}

// ProcessRequest lida com uma string de requisição raw.
func (ec *EngineContext) ProcessRequest(rawRequestBody []byte) (map[string]interface{}, error) {
	var req Request
	if err := json.Unmarshal(rawRequestBody, &req); err != nil {
		return nil, fmt.Errorf("falha ao desserializar requisição: %v", err)
	}

	// 2. Validar dados contra o schema da requisição
	if ec.RequestSchema != nil {
		_, validationErrors := ec.RequestSchema.Validate(req.Data)
		if len(validationErrors) > 0 {
			var errMsgs []string
			for _, vErr := range validationErrors {
				errMsgs = append(errMsgs, vErr.Error())
			}
			return nil, fmt.Errorf("validação do schema dos dados da requisição falhou: %s", strings.Join(errMsgs, "; "))
		}

		// validationErrors := validateDataAgainstSchema(req.Data, ec.RequestSchema, "")
		// if len(validationErrors) > 0 {
		// 	var errMsgs []string
		// 	for _, vErr := range validationErrors {
		// 		errMsgs = append(errMsgs, vErr.Error())
		// 	}
		// 	return nil, fmt.Errorf("validação do schema dos dados da requisição falhou: %s", strings.Join(errMsgs, "; "))
		// }
	}

	// 3. Executar políticas
	policyExecutionResults, allPoliciesPassed := ec.ExecutePolicies(req.Data, req.Policies)

	// 4. Lidar com falhas de política
	if !allPoliciesPassed {
		errorMessages := []string{"Execução de política(s) falhou:"}
		for _, res := range policyExecutionResults {
			status := "PASSOU"
			if !res.Passed {
				status = "FALHOU"
			}
			errMsg := ""
			if res.Error != nil {
				errMsg = fmt.Sprintf(" Erro: %s.", res.Error.Error())
			}

			var ruleDetailsStrings []string
			for _, rr := range res.RuleResults {
				ruleStatus := "OK"
				// Para regras de condição, !rr.Passed significa que a condição não foi atendida.
				// Para SET/IF, rr.Passed geralmente é true se a operação foi tentada; um erro real estaria em res.Error ou no details.
				isSetOrIf := strings.HasPrefix(strings.TrimSpace(rr.Rule), "SET ") || strings.HasPrefix(strings.TrimSpace(rr.Rule), "IF ")
				if !isSetOrIf && !rr.Passed {
					ruleStatus = "FALHA_CONDICAO"
				} else if strings.Contains(rr.Details, "Erro:") { // Se o detalhe da regra indica um erro de execução
					ruleStatus = "ERRO_EXECUCAO"
				}
				ruleDetailsStrings = append(ruleDetailsStrings, fmt.Sprintf("    - Regra: '%s', Status: %s, Detalhes: %s", rr.Rule, ruleStatus, rr.Details))
			}

			errorMessages = append(errorMessages, fmt.Sprintf("  Política '%s': %s.%s\n  Detalhes das Regras:\n%s", res.PolicyName, status, errMsg, strings.Join(ruleDetailsStrings, "\n")))
		}
		return nil, errors.New(strings.Join(errorMessages, "\n"))
	}

	// 5. Montar resposta baseada no schema de resposta (simplificado: retorna dados modificados)
	responsePayload := make(map[string]interface{})
	responsePayload["id"] = req.ID + "-response"
	responsePayload["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	responsePayload["status"] = "success"
	responsePayload["processedData"] = req.Data // Os dados após as políticas

	// TODO: Implementar transformação real do schema de resposta usando ec.ResponseSchema

	return responsePayload, nil
}
