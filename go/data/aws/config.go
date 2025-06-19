package aws

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/raywall/cloud-service-pack/go/data/types"
)

type (
	// SSMConfig representa a configuração inline do AWS SSM
	SSMConfig types.Config

	// SecretConfig representa a configuração inline do AWS Secrets Manager
	SecretConfig types.Config

	// S3Config representa a configuração inline do AWS S3
	S3Config types.Config
)

// Regex para validação
var (
	// SSM: aws:ssm:<parameter>[:true|false][:base64]
	ssmRegex = regexp.MustCompile(fmt.Sprintf(`^(%s::%s::){1}([a-zA-Z0-9_/.-]+){1}(?:(::(true|false)){1})?(?:(::[A-Za-z0-9+/=]+))?$`, types.AWS, types.SSM))
	// Secrets: aws:secrets:<secret>[:json|text][:base64]
	secretRegex = regexp.MustCompile(fmt.Sprintf(`^(%s::%s::){1}([a-zA-Z0-9_/.-]+){1}(?:(::(json|text)){1})?(?:(::[A-Za-z0-9+/=]+))?$`, types.AWS, types.SecretsManager))
	// S3: aws:s3:<bucket>[:file][:base64]
	s3Regex = regexp.MustCompile(fmt.Sprintf(`^(%s::%s::){1}([a-zA-Z0-9_-]+){1}(?:(::[a-zA-Z0-9_/.-]+){1})?(?:(::[A-Za-z0-9+/=]+))?$`, types.AWS, types.S3))
)

// Extract para SSMConfig implementa a interface parser
func (s *SSMConfig) Extract() (*types.Config, error) {
	if !ssmRegex.MatchString(s.Name) {
		return nil, fmt.Errorf("formato de configuração AWS SSM inválido: %s", s.Name)
	}

	parts := strings.Split(s.Name, "::")
	if len(parts) < 3 {
		return nil, fmt.Errorf("formato de configuração AWS SSM inválido: número de partes incorreto")
	}

	*s = SSMConfig{
		Provider:  types.AWS,
		Service:   types.SSM,
		Name:      parts[2],
		Attribute: false,
	}

	if len(parts) > 3 {
		s.Attribute = parts[3] == "true"
	}

	if len(parts) > 4 {
		s.DefaultValue = parts[4]
	}

	config := types.Config(*s)
	return &config, nil
}

// Extract para SecretConfig implementa a interface ConfigExtractor
func (s *SecretConfig) Extract() (*types.Config, error) {
	if !secretRegex.MatchString(s.Name) {
		return nil, fmt.Errorf("formato de configuração AWS Secret inválido: %s", s.Name)
	}

	parts := strings.Split(s.Name, "::")
	if len(parts) < 3 {
		return nil, fmt.Errorf("formato de configuração AWS Secret inválido: número de partes incorreto")
	}

	*s = SecretConfig{
		Provider:  types.AWS,
		Service:   types.SecretsManager,
		Name:      parts[2],
		Attribute: "text",
	}

	if len(parts) > 3 {
		s.Attribute = parts[3]
	}

	if len(parts) > 4 {
		s.DefaultValue = parts[4]
	}

	config := types.Config(*s)
	return &config, nil
}

// Extract para S3Config implementa a interface ConfigExtractor
func (s *S3Config) Extract() (*types.Config, error) {
	if !s3Regex.MatchString(s.Name) {
		return nil, fmt.Errorf("formato de configuração AWS S3 inválido: %s", s.Name)
	}

	parts := strings.Split(s.Name, "::")
	if len(parts) < 4 {
		return nil, fmt.Errorf("formato de configuração AWS S3 inválido: número de partes incorreto")
	}

	*s = S3Config{
		Provider:  types.AWS,
		Service:   types.S3,
		Name:      parts[2],
		Attribute: parts[3],
	}

	if len(parts) > 4 {
		s.DefaultValue = parts[4]
	}

	config := types.Config(*s)
	return &config, nil
}

// ParseConfig é a função para processar configurações AWS
func ParseConfig(configString string) (*types.Config, error) {
	var p parser

	switch {
	case strings.HasPrefix(configString, fmt.Sprintf("%s::%s::", types.AWS, types.SSM)):
		p = &SSMConfig{Name: configString}

	case strings.HasPrefix(configString, fmt.Sprintf("%s::%s::", types.AWS, types.SecretsManager)):
		p = &SecretConfig{Name: configString}

	case strings.HasPrefix(configString, fmt.Sprintf("%s::%s::", types.AWS, types.S3)):
		p = &S3Config{Name: configString}

	default:
		return nil, fmt.Errorf("tipo de configuração AWS desconhecido: %s", configString)
	}

	return p.Extract()
}
