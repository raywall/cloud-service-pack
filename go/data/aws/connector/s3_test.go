package connector

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/raywall/cloud-service-pack/go/data/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v3"
)

type mockS3Client struct {
	mock.Mock
}

type bucketClient struct {
	mockS3 *mockS3Client
	ctx    *S3CloudContext
}

func (m *mockS3Client) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

var bc = bucketClient{}

func loadDefaultBucketVariables() {
	mockS3 := new(mockS3Client)

	bc.mockS3 = mockS3
	bc.ctx = &S3CloudContext{
		svc: mockS3,
	}
}

func TestRetrieverS3ConfigValue(t *testing.T) {
	t.Run("Recuperar dados de um objeto JSON", func(t *testing.T) {
		loadDefaultBucketVariables()

		// Preparar mock para S3 com arquivo JSON
		jsonContent := `{"name": "test", "value": 123}`
		bc.mockS3.On("GetObject", mock.Anything).Return(&s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewReader([]byte(jsonContent))),
		}, nil)

		// Executar GetValue
		result, err := bc.ctx.GetValue("test-bucket", "test-file.json")

		// Verificar resultados
		assert.NoError(t, err)

		jsonResult := make(map[string]interface{})
		_ = json.Unmarshal([]byte(*result), &jsonResult)

		assert.Equal(t, "test", jsonResult["name"])
		assert.Equal(t, float64(123), jsonResult["value"])
	})

	t.Run("Recuperar dados de um objeto YAML", func(t *testing.T) {
		loadDefaultBucketVariables()

		// Preparar mock para S3 com arquivo YAML
		yamlContent := `name: test
value: 123
list:
  - item1
  - item2`

		bc.mockS3.On("GetObject", mock.Anything).Return(&s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewReader([]byte(yamlContent))),
		}, nil)

		// Executar GetValue
		result, err := bc.ctx.GetValue("test-bucket", "test-file.yaml")

		// Verificar resultados
		assert.NoError(t, err)

		yamlResult := make(map[string]interface{})
		_ = yaml.Unmarshal([]byte(*result), &yamlResult)

		assert.Equal(t, "test", yamlResult["name"])
		assert.Equal(t, 123, yamlResult["value"])
	})

	t.Run("Recuperar dados de um objeto CSV", func(t *testing.T) {
		loadDefaultBucketVariables()

		// Preparar mock para S3 com arquivo CSV
		csvContent := `id,name,age
1,Alice,30
2,Bob,25`

		bc.mockS3.On("GetObject", mock.Anything).Return(&s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewReader([]byte(csvContent))),
		}, nil)

		// Executar GetValue
		result, err := bc.ctx.GetValue("test-bucket", "test-file.csv")

		// Verificar resultados
		assert.NoError(t, err)

		data, _ := types.ParseStringToCSV(*result)
		csvResult, ok := data.([]map[string]interface{})

		assert.True(t, ok)
		assert.Len(t, csvResult, 2)
		assert.Equal(t, "Alice", csvResult[0]["name"])
		assert.Equal(t, "30", csvResult[0]["age"])
		assert.Equal(t, "Bob", csvResult[1]["name"])
		assert.Equal(t, "25", csvResult[1]["age"])
	})

	t.Run("Recuperar dados de um arquivo de texto", func(t *testing.T) {
		loadDefaultBucketVariables()

		// Preparar mock para S3 com arquivo de texto
		textContent := "This is a plain text file."
		bc.mockS3.On("GetObject", mock.Anything).Return(&s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewReader([]byte(textContent))),
		}, nil)

		// Executar GetValue
		result, err := bc.ctx.GetValue("test-bucket", "test-file.txt")

		// Verificar resultados
		assert.NoError(t, err)
		assert.Equal(t, textContent, *result)
	})
}
