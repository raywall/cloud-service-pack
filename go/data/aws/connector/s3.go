package connector

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/yaml.v3"
)

type S3Resource interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

// S3CloudContext implements CloudContext para S3
type S3CloudContext struct {
	svc S3Resource
}

func NewS3Context(sess *session.Session) *S3CloudContext {
	return &S3CloudContext{
		svc: s3.New(sess),
	}
}

// GetValue obtém o conteúdo do arquivo S3 e o converte para o formato apropriado
func (ctx *S3CloudContext) GetValue(bucketName, keyName string) (interface{}, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	}

	result, err := ctx.svc.GetObject(input)
	if err != nil {
		return nil, fmt.Errorf("error when obtaining S3 object: %w", err)
	}
	defer result.Body.Close()

	// Ler o conteúdo do arquivo
	bodyBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading file content: %w", err)
	}
	content := string(bodyBytes)

	// Determinar o tipo de arquivo a processar de acordo
	if strings.HasSuffix(keyName, ".json") {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
			return nil, fmt.Errorf("error when analyzing JSON: %w", err)
		}
		return jsonData, nil

	} else if strings.HasSuffix(keyName, ".yaml") || strings.HasSuffix(keyName, ".yml") {
		var yamlData map[string]interface{}
		if err := yaml.Unmarshal(bodyBytes, &yamlData); err != nil {
			return nil, fmt.Errorf("error when analyzing yaml: %w", err)
		}
		return yamlData, nil

	} else if strings.HasSuffix(keyName, ".csv") {
		reader := csv.NewReader(strings.NewReader(content))
		records, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("error when analyzing CSV: %w", err)
		}

		// Converter CSV para mapa
		if len(records) < 2 {
			return records, nil // Retorna registros crus se não houver cabeçalho
		}

		headers := records[0]
		result := make([]map[string]string, 0, len(records)-1)

		for i := 1; i < len(records); i++ {
			row := make(map[string]string)
			for j := 0; j < len(headers) && j < len(records[i]); j++ {
				row[headers[j]] = records[i][j]
			}
			result = append(result, row)
		}
		return result, nil

	} else {
		// Assumir que é um arquivo de teste simples
		return content, nil
	}
}
