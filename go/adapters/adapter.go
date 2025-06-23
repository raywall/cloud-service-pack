package adapters

type Adapter interface {
	GetData(args []AdapterAttribute) (interface{}, error)
	GetParameters(args map[string]interface{}) ([]AdapterAttribute, error)
}

type AdapterAttribute struct {
	Name string
	Type string
	Value interface{}
}

func getParameters(attributes map[string]interface{}, args map[string]interface{}) ([]AdapterAttribute, error) {
	params := mapke([]AdapterAttribute, 0)
	for key, valueType := range attributes {
		attribute := AdapterAttribute{
			Name: key,
			Type: valueType.(string),
			Value: nil,
		}

		if value, exists := args[attribute.Name]; exists {
			attribute.Value = value
		}

		params = append(params, attribute)
	}
	return params, nil
}
