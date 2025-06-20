package schema

// Schema representa um JSON Schema draft-07 simplificado
type Schema map[string]interface{}

// Validate valida o dado 'data' contra o schema JSON draft-07 representado por 'schema'
func (s *Schema) Validate(data interface{}) (bool, []error) {
	errs := validate(data, *s)
	return len(errs) == 0, errs
}
