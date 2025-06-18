package connector

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para SSM
type mockSSMClient struct {
	mock.Mock
}

type pClient struct {
	mockSSM *mockSSMClient
	ctx     *SSMCloudContext
}

func (m *mockSSMClient) GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*ssm.GetParameterOutput), args.Error(1)
}

var params = pClient{}

func loadDefaultParameterVariables() {
	mockSSM := new(mockSSMClient)
	params.mockSSM = mockSSM
	params.ctx = &SSMCloudContext{
		svc: mockSSM,
	}
}

func TestSSMCloudContext_GetValue(t *testing.T) {
	t.Run("Get parameter content", func(t *testing.T) {
		loadDefaultParameterVariables()

		// Preparar mock para SSM
		paramValue := "test-parameter-value"
		params.mockSSM.On("GetParameter", mock.Anything).Return(&ssm.GetParameterOutput{
			Parameter: &ssm.Parameter{
				Value: aws.String(paramValue),
			},
		}, nil)

		// executar GetValue
		result, err := params.ctx.GetValue("/test/param", true)

		// Verificar resultados
		assert.NoError(t, err)
		assert.Equal(t, paramValue, result)
	})
}
