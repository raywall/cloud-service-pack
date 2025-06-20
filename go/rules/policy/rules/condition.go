package rules

import (
	"fmt"
	"regexp"
	"strings"
)

func ifCondition(trimmedRule string, data map[string]interface{}) RuleExecutionResult {
	if strings.HasPrefix(trimmedRule, "IF ") {
		parts := regexp.MustCompile(`^IF\s+(.+?)\s+THEN\s+(.+)$`).FindStringSubmatch(trimmedRule)
		if len(parts) != 3 {
			return RuleExecutionResult{
				Executed: true,
				Passed:   false,
				Details:  "",
				Err:      fmt.Errorf("regra IF...THEN inválida: %s", trimmedRule),
			}
		}
		conditionStr, actionStr := strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])
		conditionMet, condDetails, errCond := EvaluateRule(conditionStr, data)
		if errCond != nil {
			return RuleExecutionResult{
				Executed: true,
				Passed:   false,
				Details:  fmt.Sprintf("Erro condição IF ('%s'): %s", conditionStr, condDetails),
				Err:      errCond,
			}
		}
		if conditionMet {
			actionPassed, actionDetails, actionErr := EvaluateRule(actionStr, data) // Ação pode ser SET com EXP
			if actionErr != nil {
				return RuleExecutionResult{
					Executed: true,
					Passed:   false,
					Details:  fmt.Sprintf("IF (%s) -> true, erro ação THEN ('%s'): %s", condDetails, actionStr, actionDetails),
					Err:      actionErr,
				}
			}
			return RuleExecutionResult{
				Executed: true,
				Passed:   actionPassed,
				Details:  fmt.Sprintf("IF (%s) -> true, THEN (%s) -> resultado ação: %t", condDetails, actionDetails, actionPassed),
				Err:      nil,
			}
		}
		return RuleExecutionResult{
			Executed: true,
			Passed:   true,
			Details:  fmt.Sprintf("IF (%s) -> false, ação ignorada: %s", condDetails, actionStr),
			Err:      nil,
		}
	}

	return RuleExecutionResult{
		Executed: false,
	}
}

func orCondition(trimmedRule string, data map[string]interface{}) RuleExecutionResult {
	if strings.Contains(trimmedRule, " OR ") && !isOperatorProtected(trimmedRule, " OR ") {
		parts := strings.SplitN(trimmedRule, " OR ", 2)
		if len(parts) == 2 {
			leftRule, rightRule := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			leftPassed, leftDetails, leftErr := EvaluateRule(leftRule, data)
			if leftErr != nil {
				return RuleExecutionResult{
					Executed: true,
					Passed:   false,
					Details:  fmt.Sprintf("Erro LHS OR ('%s'): %s", leftRule, leftDetails),
					Err:      leftErr,
				}
			}
			if leftPassed {
				return RuleExecutionResult{
					Executed: true,
					Passed:   true,
					Details:  fmt.Sprintf("(%s) OR ('%s' não avaliada) -> true", leftDetails, rightRule),
					Err:      nil,
				}
			}
			rightPassed, rightDetails, rightErr := EvaluateRule(rightRule, data)
			if rightErr != nil {
				return RuleExecutionResult{
					Executed: true,
					Passed:   false,
					Details:  fmt.Sprintf("Erro RHS OR ('%s'): %s", rightRule, rightDetails),
					Err:      rightErr,
				}
			}
			return RuleExecutionResult{
				Executed: true,
				Passed:   rightPassed,
				Details:  fmt.Sprintf("(%s) OR (%s) -> %t", leftDetails, rightDetails, rightPassed),
				Err:      nil,
			}
		}
	}

	return RuleExecutionResult{
		Executed: false,
	}
}
