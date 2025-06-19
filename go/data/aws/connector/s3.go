package connector

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
func (ctx *S3CloudContext) GetValue(bucketName, keyName string) (*string, error) {
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

	return &content, nil
}
