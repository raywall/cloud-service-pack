package schema

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"time"
)

func validateFormat(value string, format string) []error {
	switch format {
	case "date-time":
		// tenta parsear no formato RFC3339
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			return []error{fmt.Errorf("formato inválido para date-time: %v", err)}
		}
	// pode adicionar outros formatos (email, uri, etc)
	default:
		// formatos desconhecidos são ignorados
	}
	return nil
}

func validateEnum(data interface{}, enumVals interface{}) []error {
	enumArr, ok := enumVals.([]interface{})
	if !ok {
		return []error{errors.New("enum inválido no schema")}
	}
	for _, v := range enumArr {
		if reflect.DeepEqual(v, data) {
			return nil
		}
	}
	return []error{fmt.Errorf("valor '%v' não está no enum permitido", data)}
}

func validateMinimum(data interface{}, min interface{}) []error {
	fmin, ok := toFloat(min)
	if !ok {
		return []error{errors.New("minimum inválido no schema")}
	}
	fval, ok := toFloat(data)
	if !ok {
		return []error{fmt.Errorf("valor '%v' não é numérico para minimum", data)}
	}
	if fval < fmin {
		return []error{fmt.Errorf("valor %v menor que minimum %v", fval, fmin)}
	}
	return nil
}

func validateMaximum(data interface{}, max interface{}) []error {
	fmax, ok := toFloat(max)
	if !ok {
		return []error{errors.New("maximum inválido no schema")}
	}
	fval, ok := toFloat(data)
	if !ok {
		return []error{fmt.Errorf("valor '%v' não é numérico para maximum", data)}
	}
	if fval > fmax {
		return []error{fmt.Errorf("valor %v maior que maximum %v", fval, fmax)}
	}
	return nil
}

func validatePattern(data interface{}, pattern string) []error {
	s, ok := data.(string)
	if !ok {
		return nil // pattern só faz sentido para string
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return []error{fmt.Errorf("pattern inválido no schema: %v", err)}
	}
	if !re.MatchString(s) {
		return []error{fmt.Errorf("string '%s' não corresponde ao pattern '%s'", s, pattern)}
	}
	return nil
}
