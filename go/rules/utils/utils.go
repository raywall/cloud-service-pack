package utils

import (
	"encoding/json"
	"io/ioutil"

	"github.com/raywall/cloud-policy-serializer/pkg/json/schema"
	"github.com/raywall/cloud-policy-serializer/pkg/policy"

	"gopkg.in/yaml.v3"
)

type FilePath string

func (fp *FilePath) GetSchema() (*schema.Schema, error) {
	fileContent, err := ioutil.ReadFile(string(*fp))
	if err != nil {
		return nil, err
	}

	jsonData := schema.Schema{}
	err = json.Unmarshal(fileContent, &jsonData)
	if err != nil {
		return nil, err
	}
	return &jsonData, nil
}

func (fp *FilePath) GetPolicies() (*map[string]policy.PolicyDefinition, error) {
	fileContent, err := ioutil.ReadFile(string(*fp))
	if err != nil {
		return nil, err
	}

	itens := make(map[string][]string, 0)
	err = yaml.Unmarshal(fileContent, &itens)
	if err != nil {
		return nil, err
	}

	data := make(map[string]policy.PolicyDefinition, 0)
	for key, value := range itens {
		data[key] = policy.PolicyDefinition{
			Name:  key,
			Rules: value,
		}
	}
	return &data, nil
}
