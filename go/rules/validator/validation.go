package validator

import (
	"fmt"
	"reflect"

	"github.com/raywall/cloud-policy-serializer/pkg/json/schema"
)

// --- Validação de Schema (Simplificada) ---
// validateDataAgainstSchema verifica se os dados estão em conformidade com um schema simplificado.
// Formato do Schema: {"fieldName": {"type": "string|number|bool|object|array", "required": true/false, "properties": {schema para sub-objeto}, "items": {schema para itens de array}}}
func validateDataAgainstSchema(data map[string]interface{}, sch *schema.Schema, currentPathPrefix string) []error {
	var validationErrors []error

	for key, schemaDefInterface := range *sch {
		schemaDef, ok := schemaDefInterface.(map[string]interface{})
		if !ok {
			validationErrors = append(validationErrors, fmt.Errorf("definição de schema inválida para chave '%s' no caminho '%s'", key, currentPathPrefix))
			continue
		}

		fullPath := key
		if currentPathPrefix != "" {
			fullPath = currentPathPrefix + "." + key
		}

		value, existsInData := data[key]

		required, _ := schemaDef["required"].(bool)
		if required && !existsInData {
			validationErrors = append(validationErrors, fmt.Errorf("campo obrigatório ausente '%s'", fullPath))
			continue
		}

		if !existsInData { // Não obrigatório e não presente, pular verificações adicionais
			continue
		}

		expectedType, typeSpecified := schemaDef["type"].(string)
		if typeSpecified {
			actualKind := reflect.TypeOf(value).Kind()
			validType := false
			switch expectedType {
			case "string":
				if actualKind == reflect.String {
					validType = true
				}
			case "number": // Números JSON podem ser float64 ao desserializar
				if actualKind == reflect.Float64 || actualKind == reflect.Int || actualKind == reflect.Int32 || actualKind == reflect.Int64 {
					validType = true
				}
			case "boolean":
				if actualKind == reflect.Bool {
					validType = true
				}
			case "object":
				if subMap, ok := value.(map[string]interface{}); ok {
					validType = true
					if properties, pExists := schemaDef["properties"].(schema.Schema); pExists {
						validationErrors = append(validationErrors, validateDataAgainstSchema(subMap, &properties, fullPath)...)
					}
				}
			case "array":
				if arr, ok := value.([]interface{}); ok {
					validType = true
					if itemSchemaDef, iExists := schemaDef["items"].(schema.Schema); iExists { // Schema para itens do array
						for i, item := range arr {
							itemPath := fmt.Sprintf("%s[%d]", fullPath, i)
							// Valida cada item recursivamente. Assume que itens são objetos se 'properties' estiver definido no itemSchema.
							// Para simplificar, se 'items' tem 'type', valida esse tipo básico.
							if itemMap, isMap := item.(map[string]interface{}); isMap {
								validationErrors = append(validationErrors, validateDataAgainstSchema(itemMap, &itemSchemaDef, itemPath)...)
							} else {
								itemExpectedType, itemTypeSpecified := itemSchemaDef["type"].(string)
								if itemTypeSpecified {
									itemActualKind := reflect.TypeOf(item).Kind()
									itemValid := false
									switch itemExpectedType {
									case "string":
										if itemActualKind == reflect.String {
											itemValid = true
										}
									case "number":
										if itemActualKind == reflect.Float64 || itemActualKind == reflect.Int {
											itemValid = true
										}
									case "boolean":
										if itemActualKind == reflect.Bool {
											itemValid = true
										}
									default:
										validationErrors = append(validationErrors, fmt.Errorf("tipo de item não suportado '%s' no schema para array '%s'", itemExpectedType, fullPath))
									}
									if !itemValid {
										validationErrors = append(validationErrors, fmt.Errorf("tipo inválido para item em %s: esperado %s, obtido %v", itemPath, itemExpectedType, itemActualKind))
									}
								} else {
									// Se 'items' não especifica um 'type' nem 'properties', é um schema malformado para validação de itens.
									validationErrors = append(validationErrors, fmt.Errorf("schema de item malformado para array '%s'", fullPath))
								}
							}
						}
					}
				}
			default:
				validationErrors = append(validationErrors, fmt.Errorf("tipo não suportado '%s' no schema para chave '%s'", expectedType, fullPath))
			}
			if !validType {
				validationErrors = append(validationErrors, fmt.Errorf("tipo inválido para campo '%s': esperado %s, obtido %v", fullPath, expectedType, actualKind))
			}
		}
	}
	return validationErrors
}
