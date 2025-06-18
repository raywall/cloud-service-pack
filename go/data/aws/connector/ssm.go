package connector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type SSMResource interface {
	GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}

// SSMCloudContext implementa CloudContext para SSM Parameter Store
type SSMCloudContext struct {
	svc SSMResource
}

func NewSSMContext(sess *session.Session) *SSMCloudContext {
	return &SSMCloudContext{
		svc: ssm.New(sess),
	}
}

// GetValue obtém o valor do parâmetro SSM
func (ctx *SSMCloudContext) GetValue(parameterName string, withDecryption bool) (interface{}, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(withDecryption),
	}

	result, err := ctx.svc.GetParameter(input)
	if err != nil {
		return nil, fmt.Errorf("error when obtaining SSM parameters: %w", err)
	}
	return *result.Parameter.Value, nil
}
