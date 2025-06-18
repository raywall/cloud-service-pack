package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/raywall/cloud-service-pack/go/data/aws"
	"github.com/raywall/cloud-service-pack/go/data/aws/connector"
	"github.com/raywall/cloud-service-pack/go/data/types"
)

// ParseConfig é a função principal para processar configurações de qualquer provedor cloud
func ParseConfig(inlineConfig string) (*types.Config, error) {
	switch {
	case strings.HasPrefix(inlineConfig, fmt.Sprintf("%s:", types.AWS)):
		return aws.ParseConfig(inlineConfig)
	default:
		return nil, fmt.Errorf("cloud desconhecida: %s", inlineConfig)
	}
}

// GetValue é a função responsável por recuperar as configurações nos serviços de qualquer provedor cloud
func GetValue(config *types.Config, sess *session.Session) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	switch {
	case config.Provider == types.AWS && config.Service == types.S3:
		ctx := connector.NewS3Context(sess)
		obj, err := ctx.GetValue(config.Name, config.Attribute.(string))
		if err != nil {
			return nil, err
		}

		data = obj.(map[string]interface{})
		return data, nil

	case config.Provider == types.AWS && config.Service == types.SSM:
		ctx := connector.NewSSMContext(sess)
		param, err := ctx.GetValue(config.Name, config.Attribute.(bool))
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(param.(string)), &data); err != nil {
			data["data"] = param.(string)
		}
		return data, nil

	case config.Provider == types.AWS && config.Service == types.SecretsManager:
		ctx := connector.NewSecretsManagerContext(sess)
		secret, err := ctx.GetValue(config.Name, config.Attribute.(string))
		if err != nil {
			return nil, err
		}

		if config.Attribute.(string) == "json" {
			data = secret.(map[string]interface{})
		} else {
			data["data"] = secret.(string)
		}

		return data, nil

	default:
		return nil, fmt.Errorf("cloud e serviço desconhecidos: %s - %s", config.Provider, config.Service)
	}
}
