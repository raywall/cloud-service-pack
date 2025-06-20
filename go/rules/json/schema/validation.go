package schema

import (
	"errors"
	"fmt"
)

func validate(data interface{}, s Schema) []error {
	var errs []error

	// Validar o tipo principal
	if t, ok := s["type"]; ok {
		switch t := t.(type) {
		case string:
			errs = append(errs, validateType(data, t)...)
		case []interface{}: // tipo pode ser array de strings
			var valid bool
			var typeErrs []error
			for _, tt := range t {
				if ts, ok := tt.(string); ok {
					errsTmp := validateType(data, ts)
					if len(errsTmp) == 0 {
						valid = true
						break
					} else {
						typeErrs = append(typeErrs, errsTmp...)
					}
				}
			}
			if !valid {
				errs = append(errs, typeErrs...)
			}
		default:
			errs = append(errs, fmt.Errorf("tipo inválido no schema: %v", t))
		}
	}

	// Se for objeto, validar propriedades
	if typ, ok := s["type"]; ok && typ == "object" {
		obj, ok := data.(map[string]interface{})
		if !ok {
			errs = append(errs, errors.New("esperado objeto JSON"))
			return errs
		}
		errs = append(errs, validateObject(obj, s)...)
	}

	// Se for array, validar itens
	if typ, ok := s["type"]; ok && typ == "array" {
		arr, ok := data.([]interface{})
		if !ok {
			errs = append(errs, errors.New("esperado array JSON"))
			return errs
		}
		errs = append(errs, validateArray(arr, s)...)
	}

	// Validar formatos (exemplo: date-time)
	if format, ok := s["format"]; ok {
		if s, ok := data.(string); ok {
			errs = append(errs, validateFormat(s, format.(string))...)
		}
	}

	// Validar enum
	if enumVals, ok := s["enum"]; ok {
		errs = append(errs, validateEnum(data, enumVals)...)
	}

	// Validar mínimo e máximo para números
	if min, ok := s["minimum"]; ok {
		errs = append(errs, validateMinimum(data, min)...)
	}
	if max, ok := s["maximum"]; ok {
		errs = append(errs, validateMaximum(data, max)...)
	}

	// Validar pattern para strings
	if pattern, ok := s["pattern"]; ok {
		errs = append(errs, validatePattern(data, pattern.(string))...)
	}

	return errs
}

func validateType(data interface{}, t string) []error {
	switch t {
	case "string":
		if _, ok := data.(string); !ok {
			return []error{fmt.Errorf("esperado string, encontrado %T", data)}
		}
	case "number":
		if !isNumber(data) {
			return []error{fmt.Errorf("esperado number, encontrado %T", data)}
		}
	case "integer":
		if !isInteger(data) {
			return []error{fmt.Errorf("esperado integer, encontrado %T", data)}
		}
	case "boolean":
		if _, ok := data.(bool); !ok {
			return []error{fmt.Errorf("esperado boolean, encontrado %T", data)}
		}
	case "object":
		if _, ok := data.(map[string]interface{}); !ok {
			fmt.Println(data)
			return []error{fmt.Errorf("esperado object, encontrado %T", data)}
		}
	case "array":
		if _, ok := data.([]interface{}); !ok {
			return []error{fmt.Errorf("esperado array, encontrado %T", data)}
		}
	case "null":
		if data != nil {
			return []error{fmt.Errorf("esperado null, encontrado %T", data)}
		}
	default:
		return []error{fmt.Errorf("tipo desconhecido no schema: %s", t)}
	}
	return nil
}

func validateObject(obj map[string]interface{}, s Schema) []error {
	var errs []error

	// required
	if req, ok := s["required"]; ok {
		if reqArr, ok := req.([]interface{}); ok {
			for _, r := range reqArr {
				if key, ok := r.(string); ok {
					if _, exists := obj[key]; !exists {
						errs = append(errs, fmt.Errorf("campo obrigatório '%s' ausente", key))
					}
				}
			}
		}
	}

	// properties
	props, hasProps := s["properties"].(map[string]interface{})
	for key, val := range obj {
		if hasProps {
			if propSchemaRaw, ok := props[key]; ok {
				if propSchema, ok := propSchemaRaw.(map[string]interface{}); ok {
					errs = append(errs, validate(val, propSchema)...)
				} else {
					errs = append(errs, fmt.Errorf("schema inválido para propriedade '%s'", key))
				}
			} else {
				// Verifica additionalProperties
				if addProps, ok := s["additionalProperties"]; ok {
					switch v := addProps.(type) {
					case bool:
						if !v {
							errs = append(errs, fmt.Errorf("propriedade adicional '%s' não permitida", key))
						}
					case map[string]interface{}:
						errs = append(errs, validate(val, v)...)
					default:
						// assume permitido
					}
				}
				// else {
				// 	// Por padrão additionalProperties é permitido

				// }
			}
		}
	}

	return errs
}

func validateArray(arr []interface{}, s Schema) []error {
	var errs []error

	// items pode ser schema ou array de schemas
	if items, ok := s["items"]; ok {
		switch items := items.(type) {
		case map[string]interface{}:
			for i, item := range arr {
				errs = append(errs, prefixErrors(fmt.Sprintf("items[%d]", i), validate(item, items))...)
			}
		case []interface{}:
			for i, item := range arr {
				if i < len(items) {
					if itemSchema, ok := items[i].(map[string]interface{}); ok {
						errs = append(errs, prefixErrors(fmt.Sprintf("items[%d]", i), validate(item, itemSchema))...)
					}
				} else {
					// Pode validar additionalItems se definido
					if addItems, ok := s["additionalItems"]; ok {
						switch addItems := addItems.(type) {
						case bool:
							if !addItems {
								errs = append(errs, fmt.Errorf("item adicional no índice %d não permitido", i))
							}
						case map[string]interface{}:
							errs = append(errs, prefixErrors(fmt.Sprintf("additionalItems[%d]", i), validate(item, addItems))...)
						}
					}
				}
			}
		}
	}

	// minItems
	if minItems, ok := s["minItems"]; ok {
		if min, err := toInt(minItems); err == nil {
			if len(arr) < min {
				errs = append(errs, fmt.Errorf("array menor que minItems (%d)", min))
			}
		}
	}

	// maxItems
	if maxItems, ok := s["maxItems"]; ok {
		if max, err := toInt(maxItems); err == nil {
			if len(arr) > max {
				errs = append(errs, fmt.Errorf("array maior que maxItems (%d)", max))
			}
		}
	}

	// uniqueItems (não implementado, pode ser adicionado)

	return errs
}
