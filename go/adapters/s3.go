package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Adapter interface {
	Adapter
}

type s3Adapter struct {
	client *s3.Client
	bucket string
	attr   map[string]interface{}
}

func NewS3Adapter(region, bucket, accessKeyId, secretAccessKey string) S3Adapter {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     accessKeyId,
				SecretAccessKey: secretAccessKey,
			}, nil
		})),
	)
	if err != nil {
		panic(fmt.Errorf("failed to load AWS config: %v", err))
	}

	return &s3Adapter{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}
}

func (s *s3Adapter) GetData(args []AdapterAttribute) (interface{}, error) {
	key := args[0].Name

	result, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s from S3: %v", key, err)
	}
	defer result.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(result.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode S3 object %s: %v", key, err)
	}
	return data, nil
}

func (r *s3Adapter) GetParameters(args map[string]interface{}) ([]AdapterAttribute, error) {
	return getParameters(r.attr, args)
}
