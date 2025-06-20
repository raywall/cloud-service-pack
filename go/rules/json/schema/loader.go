package schema

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/raywall/cloud-policy-serializer/pkg/core/loader"
)

type (
	SchemaLoader interface {
		Load() (*Schema, error)
	}

	localLoader struct {
		loader *loader.LocalLoader
	}

	s3Loader struct {
		loader *loader.S3Loader
	}

	ssmLoader struct {
		loader *loader.SSMLoader
	}
)

func NewLoader(source string) (SchemaLoader, error) {
	ld, err := loader.NewLoader(source)
	if err != nil {
		return nil, err
	}

	switch v := ld.(type) {
	case *loader.LocalLoader:
		return &localLoader{
			loader: v,
		}, nil
	case *loader.S3Loader:
		return &s3Loader{
			loader: v,
		}, nil
	case *loader.SSMLoader:
		return &ssmLoader{
			loader: v,
		}, nil
	default:
		return nil, fmt.Errorf("tipo de loader não suportado para schema: %T", v)
	}
}

func (l *localLoader) Load() (*Schema, error) {
	jsonSchema := &Schema{}

	data, err := os.ReadFile(l.loader.Path)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if err = load(jsonSchema, data); err != nil {
		return nil, err
	}

	return jsonSchema, nil
}

func (l *ssmLoader) Load() (*Schema, error) {
	var (
		jsonSchema     = &Schema{}
		withDecryption = true
	)

	data, err := loader.GetParameter(l.loader.Client, l.loader.Path, withDecryption)
	if err != nil {
		return nil, err
	}
	if err = load(jsonSchema, data); err != nil {
		return nil, err
	}

	return jsonSchema, nil
}

func (l *s3Loader) Load() (*Schema, error) {
	jsonSchema := &Schema{}
	bucket, key := loader.ParseS3Path(l.loader.Path)

	data, err := loader.GetObject(l.loader.Client, bucket, key)
	if err != nil {
		return nil, err
	}
	if err = load(jsonSchema, data); err != nil {
		return nil, err
	}

	return jsonSchema, nil
}

// load carrega um JSON Schema draft-07
func load(schema *Schema, data []byte) error {
	if err := json.Unmarshal(data, schema); err != nil {
		return fmt.Errorf("unable to serialize ssm template file: %v", err)
	}

	return nil
}
