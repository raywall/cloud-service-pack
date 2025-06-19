package local

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/raywall/cloud-service-pack/go/data/types"
)

type (
	// FileConfig representa a configuração inline de um arquivo local
	FileConfig types.Config

	// EnvironmentConfig representa a configuração inline de uma variável de ambiente local
	EnvironmentConfig types.Config
)

var (
	fileRegex = regexp.MustCompile(fmt.Sprintf(`^%s::%s::.*$`, types.LOCAL, types.FILE))
	envRegex  = regexp.MustCompile(fmt.Sprintf(`^%s::%s::.*$`, types.LOCAL, types.ENV))
)

// Extract para FileConfig implementa a interface parser
func (s *FileConfig) Extract() (*types.Config, error) {
	if !fileRegex.MatchString(s.Name) {
		return nil, fmt.Errorf("formato de configuração de arquivo inválido: %s", s.Name)
	}

	parts := strings.Split(s.Name, "::")
	if len(parts) < 3 {
		return nil, fmt.Errorf("formato de configuração de arquivo inválido: número de partes incorreto")
	}

	*s = FileConfig{
		Provider: types.LOCAL,
		Service:  types.FILE,
		Name:     parts[2],
	}

	if len(parts) > 3 {
		s.DefaultValue = parts[3]
	}

	config := types.Config(*s)
	return &config, nil
}

// Extract para EnvironmentConfig implementa a interface parser
func (s *EnvironmentConfig) Extract() (*types.Config, error) {
	if !envRegex.MatchString(s.Name) {
		return nil, fmt.Errorf("formato de configuração de variável de ambiente inválida: %s", s.Name)
	}

	parts := strings.Split(s.Name, "::")
	if len(parts) < 3 {
		return nil, fmt.Errorf("formato de configuração de variável de ambiente inválida: número de partes incorreto")
	}

	*s = EnvironmentConfig{
		Provider: types.LOCAL,
		Service:  types.FILE,
		Name:     parts[2],
	}

	if len(parts) > 3 {
		s.DefaultValue = parts[3]
	}

	config := types.Config(*s)
	return &config, nil
}

// ParseConfig é a função para processar configurações AWS
func ParseConfig(configString string) (*types.Config, error) {
	var p parser

	switch {
	case strings.HasPrefix(configString, fmt.Sprintf("%s::%s::", types.LOCAL, types.FILE)):
		p = &FileConfig{Name: configString}

	case strings.HasPrefix(configString, fmt.Sprintf("%s::%s::", types.LOCAL, types.ENV)):
		p = &EnvironmentConfig{Name: configString}

	default:
		return nil, fmt.Errorf("tipo de configuração AWS desconhecido: %s", configString)
	}

	return p.Extract()
}
