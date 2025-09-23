package config

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// ConfigBuilder - Construtor de configurações
type ConfigBuilder struct {
	ssmClient *ssm.Client
}

// New cria uma nova instância do ConfigBuilder
func New(ssmClient *ssm.Client) *ConfigBuilder {
	return &ConfigBuilder{
		ssmClient: ssmClient,
	}
}

// BuildOptions opções para construção da configuração
type BuildOptions struct {
	Prefixes    []string
	StripPrefix bool
	JSONOutput  bool
}

// BuildConfigFromPrefixes constrói a configuração a partir dos prefixos
func (b *ConfigBuilder) BuildConfigFromPrefixes(ctx context.Context, opts BuildOptions) ([]byte, error) {
	configMap := make(map[string]interface{})

	for _, prefix := range opts.Prefixes {
		params, err := b.getParametersByPath(ctx, prefix)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar parâmetros do prefixo %s: %w", prefix, err)
		}

		prefixConfig := b.buildConfigForPrefix(params, prefix, opts.StripPrefix)
		b.mergeMaps(configMap, prefixConfig)
	}

	if opts.JSONOutput {
		return json.MarshalIndent(configMap, "", "  ")
	}
	return json.Marshal(configMap)
}

// buildConfigForPrefix constrói a configuração para um prefixo específico
func (b *ConfigBuilder) buildConfigForPrefix(params []types.Parameter, basePath string, stripPrefix bool) map[string]interface{} {
	if len(params) == 0 {
		return make(map[string]interface{})
	}

	// Agrupa parâmetros por seu caminho relativo ao basePath
	grouped := b.groupParametersByParentPath(params, basePath, stripPrefix)
	return b.buildStructureFromGroups(grouped, basePath)
}

// groupParametersByParentPath agrupa parâmetros pelo caminho do pai imediato
func (b *ConfigBuilder) groupParametersByParentPath(params []types.Parameter, basePath string, stripPrefix bool) map[string][]types.Parameter {
	groups := make(map[string][]types.Parameter)

	for _, param := range params {
		relativePath := b.getRelativePath(*param.Name, basePath, stripPrefix)
		parentPath := b.getParentPath(relativePath)
		groups[parentPath] = append(groups[parentPath], param)
	}

	return groups
}

// buildStructureFromGroups constrói a estrutura JSON a partir dos grupos
func (b *ConfigBuilder) buildStructureFromGroups(groups map[string][]types.Parameter, basePath string) map[string]interface{} {
	result := make(map[string]interface{})

	for parentPath, params := range groups {
		if parentPath == "." {
			// Parâmetros diretamente sob o basePath
			b.handleRootLevelParams(result, params, basePath)
		} else {
			// Parâmetros em subníveis
			b.handleNestedParams(result, parentPath, params)
		}
	}

	return result
}

// handleRootLevelParams lida com parâmetros no nível raiz do prefixo
func (b *ConfigBuilder) handleRootLevelParams(result map[string]interface{}, params []types.Parameter, basePath string) {
	if len(params) == 1 {
		// Apenas um parâmetro - objeto simples
		param := params[0]
		key := b.getLastPathSegment(*param.Name)
		value := b.parseParameterValue(*param.Value)
		result[key] = value
	} else {
		// Múltiplos parâmetros - array
		key := b.getLastPathSegment(basePath)
		array := b.buildArrayFromParams(params)
		result[key] = array
	}
}

// handleNestedParams lida com parâmetros aninhados
func (b *ConfigBuilder) handleNestedParams(result map[string]interface{}, parentPath string, params []types.Parameter) {
	pathParts := strings.Split(parentPath, "/")
	current := result

	// Navega até o nó pai
	for i, part := range pathParts {
		if i == len(pathParts)-1 {
			// Última parte - decide se é array ou objeto
			if len(params) > 1 {
				// Múltiplos parâmetros - array
				array := b.buildArrayFromParams(params)
				current[part] = array
			} else {
				// Apenas um parâmetro - objeto
				param := params[0]
				value := b.parseParameterValue(*param.Value)
				current[part] = value
			}
		} else {
			// Partes intermediárias - cria estrutura hierárquica
			if current[part] == nil {
				current[part] = make(map[string]interface{})
			}
			current = current[part].(map[string]interface{})
		}
	}
}

// buildArrayFromParams constrói um array a partir dos parâmetros
func (b *ConfigBuilder) buildArrayFromParams(params []types.Parameter) []interface{} {
	var array []interface{}
	for _, param := range params {
		value := b.parseParameterValue(*param.Value)
		array = append(array, value)
	}
	return array
}

// getRelativePath retorna o caminho relativo ao basePath
func (b *ConfigBuilder) getRelativePath(fullPath, basePath string, stripPrefix bool) string {
	if stripPrefix {
		basePath = strings.TrimSuffix(basePath, "/")
		relative := strings.TrimPrefix(fullPath, basePath)
		return strings.Trim(relative, "/")
	}
	return fullPath
}

// getParentPath retorna o caminho pai (remove o último segmento)
func (b *ConfigBuilder) getParentPath(path string) string {
	if path == "" {
		return "."
	}

	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		return "."
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

// getLastPathSegment retorna o último segmento do path
func (b *ConfigBuilder) getLastPathSegment(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// parseParameterValue parse o valor do parâmetro
func (b *ConfigBuilder) parseParameterValue(value string) interface{} {
	var result interface{}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return value
	}
	return result
}

// mergeMaps faz merge de dois maps recursivamente
func (b *ConfigBuilder) mergeMaps(dest, src map[string]interface{}) {
	for key, srcValue := range src {
		if destValue, exists := dest[key]; exists {
			if destMap, ok := destValue.(map[string]interface{}); ok {
				if srcMap, ok := srcValue.(map[string]interface{}); ok {
					b.mergeMaps(destMap, srcMap)
					continue
				}
			}
			if destArray, ok := destValue.([]interface{}); ok {
				if srcArray, ok := srcValue.([]interface{}); ok {
					dest[key] = append(destArray, srcArray...)
					continue
				}
			}
		}
		dest[key] = srcValue
	}
}

// getParametersByPath recupera parâmetros recursivamente
func (b *ConfigBuilder) getParametersByPath(ctx context.Context, path string) ([]types.Parameter, error) {
	var allParams []types.Parameter
	var nextToken *string

	for {
		input := &ssm.GetParametersByPathInput{
			Path:           aws.String(path),
			Recursive:      aws.Bool(true),
			WithDecryption: aws.Bool(true),
			NextToken:      nextToken,
		}

		result, err := b.ssmClient.GetParametersByPath(ctx, input)
		if err != nil {
			return nil, err
		}

		allParams = append(allParams, result.Parameters...)
		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken
	}

	sort.Slice(allParams, func(i, j int) bool {
		return *allParams[i].Name < *allParams[j].Name
	})

	return allParams, nil
}

// Métodos de conveniência
func (b *ConfigBuilder) BuildConnectorsConfig(ctx context.Context, appPrefix string) ([]byte, error) {
	opts := BuildOptions{
		Prefixes:    []string{path.Join(appPrefix, "connectors")},
		StripPrefix: true,
		JSONOutput:  true,
	}
	return b.BuildConfigFromPrefixes(ctx, opts)
}

func (b *ConfigBuilder) BuildSchemaConfig(ctx context.Context, appPrefix string) ([]byte, error) {
	opts := BuildOptions{
		Prefixes:    []string{path.Join(appPrefix, "schema")},
		StripPrefix: true,
		JSONOutput:  true,
	}
	return b.BuildConfigFromPrefixes(ctx, opts)
}

func (b *ConfigBuilder) BuildConfig(ctx context.Context, appPrefix string) ([]byte, error) {
	opts := BuildOptions{
		Prefixes:    []string{appPrefix},
		StripPrefix: true,
		JSONOutput:  true,
	}
	return b.BuildConfigFromPrefixes(ctx, opts)
}

func (b *ConfigBuilder) BuildCompleteConfig(ctx context.Context, appPrefix string) ([]byte, error) {
	opts := BuildOptions{
		Prefixes: []string{
			path.Join(appPrefix, "connectors"),
			path.Join(appPrefix, "schema"),
		},
		StripPrefix: true,
		JSONOutput:  true,
	}
	return b.BuildConfigFromPrefixes(ctx, opts)
}
