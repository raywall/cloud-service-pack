package data

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/raywall/cloud-service-pack/go/data/aws"
	ac "github.com/raywall/cloud-service-pack/go/data/aws/connector"
	"github.com/raywall/cloud-service-pack/go/data/local"
	lc "github.com/raywall/cloud-service-pack/go/data/local/connector"
	"github.com/raywall/cloud-service-pack/go/data/types"
)

type (
	String      string
	ContentType string
)

const (
	CSV  ContentType = "csv"
	JSON ContentType = "json"
	YAML ContentType = "yaml"
)

// IsConfig identifica se uma string representa uma configuração inline
//
// Ex.
// - aws::ssm::my-parameter-name				(aws parameter store)
// - aws::secrets::my-secret-name				(aws secretsmanager)
// - aws::s3::my-bucket-name::my-file-key 		(aws s3 bucket)
// - local::file::c:\my-folder\my-filename.json	(local file)
// - local::env::my-env-variable-name 			(environment variable)
func IsConfig(config string) bool {
	return regexp.
		MustCompile(`^((local::){1}(file|env){1}|(aws::){1}(s3|ssm|secrets){1}){1}(::).*$`).
		MatchString(config)
}

// ParseConfig é a função principal para processar configurações de qualquer provedor cloud
func ParseConfig(inlineConfig string) (*types.Config, error) {
	switch {
	case strings.HasPrefix(inlineConfig, fmt.Sprintf("%s::", types.AWS)):
		return aws.ParseConfig(inlineConfig)

	case strings.HasPrefix(inlineConfig, fmt.Sprintf("%s::", types.LOCAL)):
		return local.ParseConfig(inlineConfig)

	default:
		return nil, fmt.Errorf("cloud desconhecida: %s", inlineConfig)
	}
}

// GetValue é a função responsável por recuperar as configurações nos serviços de qualquer provedor cloud
func GetValue(config *types.Config, sess *session.Session) (String, error) {
	var value String

	switch {
	case config.Provider == types.AWS:
		switch config.Service {
		case types.S3:
			ctx := ac.NewS3Context(sess)
			obj, err := ctx.GetValue(config.Name, config.Attribute.(string))
			if err != nil {
				return value, err
			}
			value = String(*obj)

		case types.SSM:
			ctx := ac.NewSSMContext(sess)
			param, err := ctx.GetValue(config.Name, config.Attribute.(bool))
			if err != nil {
				return value, err
			}
			value = String(*param)

		case types.SecretsManager:
			ctx := ac.NewSecretsManagerContext(sess)
			secret, err := ctx.GetValue(config.Name, config.Attribute.(string))
			if err != nil {
				return value, err
			}
			value = String(*secret)
		}
	case config.Provider == types.LOCAL:
		ctx := lc.NewLocalContext()
		content, err := ctx.GetValue(string(config.Service), config.Name)
		if err != nil {
			return value, err
		}
		value = String(*content)
	default:
		return value, fmt.Errorf("cloud e serviço desconhecidos: %s - %s", config.Provider, config.Service)
	}

	return value, nil
}

// CastTo é o método que possibilita converter um objeto String em um mapa que consiga
// representar melhor um json, yaml ou csv
func (s *String) CastTo(contentType ContentType) (interface{}, error) {
	switch contentType {
	case CSV:
		data, err := types.ParseStringToCSV(string(*s))
		if err != nil {
			return nil, err
		}
		return data, nil

	case JSON:
		data, err := types.ParseStringToJSON(string(*s))
		if err != nil {
			return nil, err
		}
		return data, nil

	case YAML:
		data, err := types.ParseStringToYAML(string(*s))
		if err != nil {
			return nil, err
		}
		return data, nil

	default:
		return nil, fmt.Errorf("unrecognized content type: %s", contentType)
	}
}
