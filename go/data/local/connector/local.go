package connector

import (
	"fmt"
	"os"
)

type LocalContext struct{}

func NewLocalContext() *LocalContext {
	return &LocalContext{}
}

func (ctx *LocalContext) GetValue(local, name string) (*string, error) {
	switch local {
	case "file":
		content, err := os.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("falha ao recuperar os dados do arquivo %s: %v", name, err)
		}
		value := string(content)
		return &value, nil

	case "env":
		if value, ok := os.LookupEnv(name); ok {
			return &value, nil
		}

	default:
		return nil, fmt.Errorf("local desconhecido: %s", local)
	}

	return nil, nil
}
