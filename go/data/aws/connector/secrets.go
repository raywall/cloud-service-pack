package connector

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type SecretsManagerResource interface {
	GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
}

// SecretsManagerCloudContext implementa CloudContext para Secrets Manager
type SecretsManagerCloudContext struct {
	svc SecretsManagerResource
}

func NewSecretsManagerContext(sess *session.Session) *SecretsManagerCloudContext {
	return &SecretsManagerCloudContext{
		svc: secretsmanager.New(sess),
	}
}

// GetValue obtém e processa o segredo do Secrets Manager
func (ctx *SecretsManagerCloudContext) GetValue(secretName, secretType string) (*string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := ctx.svc.GetSecretValue(input)
	if err != nil {
		return nil, fmt.Errorf("error when obtaining secret: %w", err)
	}

	var secretValue string
	if result.SecretString != nil {
		secretValue = *result.SecretString
	} else {
		return nil, errors.New("binary secret is not supported")
	}

	return &secretValue, nil
}
