package loader

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type (
	LocalLoader struct {
		Path string
	}

	S3Loader struct {
		Path   string
		Client S3Client
	}

	SSMLoader struct {
		Path   string
		Client SSMClient
	}

	S3Client interface {
		GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	}

	SSMClient interface {
		GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
	}
)

func NewLoader(source string) (interface{}, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	switch {
	case strings.HasPrefix(source, "s3://"):
		return &S3Loader{
			Path:   source,
			Client: s3.NewFromConfig(cfg),
		}, nil
	case strings.HasPrefix(source, "ssm://"):
		return &SSMLoader{
			Path:   source,
			Client: ssm.NewFromConfig(cfg),
		}, nil
	default:
		return &LocalLoader{
			Path: source,
		}, nil
	}
}

func GetObject(client S3Client, bucket, key string) ([]byte, error) {
	output, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting object %s from bucket %s: %v", key, bucket, err)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	return data, nil
}

func GetParameter(client SSMClient, parameterPath string, withDecryption bool) ([]byte, error) {
	output, err := client.GetParameter(context.Background(), &ssm.GetParameterInput{
		Name:           &parameterPath,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting parameter %s: %v", parameterPath, err)
	}
	if output.Parameter == nil || output.Parameter.Value == nil {
		return nil, fmt.Errorf("invalid parameter: %v", err)
	}

	return []byte(*output.Parameter.Value), nil
}