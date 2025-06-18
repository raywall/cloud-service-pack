package connector

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para Secrets Manager
type mockSecretsManagerClient struct {
	mock.Mock
}

type secretsClient struct {
	mockSecretsManager *mockSecretsManagerClient
	ctx                *SecretsManagerCloudContext
}

func (m *mockSecretsManagerClient) GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*secretsmanager.GetSecretValueOutput), args.Error(1)
}

var secret = secretsClient{}

func loadDefaultSecretsVariables() {
	mockSecretsManager := new(mockSecretsManagerClient)
	secret.mockSecretsManager = mockSecretsManager
	secret.ctx = &SecretsManagerCloudContext{
		svc: mockSecretsManager,
	}
}

func TestSecretsManagerCloudContext_GetValue(t *testing.T) {
	t.Run("Get text secret", func(t *testing.T) {
		loadDefaultSecretsVariables()

		// Preparar mock para Secrets Manager com segredo de texto
		secretValue := "test-secret-value"
		secret.mockSecretsManager.On("GetSecretValue", mock.Anything).Return(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(secretValue),
		}, nil)

		// Executar GetValue
		result, err := secret.ctx.GetValue("test-secret", "text")

		// Verificar resultados
		assert.NoError(t, err)
		assert.Equal(t, secretValue, result)
	})

	t.Run("Get JSON secret", func(t *testing.T) {
		loadDefaultSecretsVariables()

		// Preparar mock para Secrets Manager com segredo de texto
		secretJSON := `{"username": "admin", "password": "secret123"}`
		secret.mockSecretsManager.On("GetSecretValue", mock.Anything).Return(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(secretJSON),
		}, nil)

		// Executar GetValue
		result, err := secret.ctx.GetValue("test-secret", "json")

		// Verificar resultados
		assert.NoError(t, err)

		jsonResult := make(map[string]interface{})
		err = json.Unmarshal(result.([]byte), &jsonResult)

		assert.NoError(t, err)

		assert.Equal(t, "admin", jsonResult["username"])
		assert.Equal(t, "secret123", jsonResult["password"])
	})
}
