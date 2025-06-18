package data

import (
	"testing"

	"github.com/raywall/cloud-service-pack/go/data/types"
	"github.com/stretchr/testify/assert"
)

func TestRetrieverInlineConfig(t *testing.T) {
	t.Run("Configuração inline do serviço SSM da AWS", func(t *testing.T) {
		inlineConfig := "aws:ssm:/path/to/parameter:true:SGVsbG8="

		expected := &types.Config{
			Provider:     types.AWS,
			Service:      types.SSM,
			Name:         "/path/to/parameter",
			Attribute:    true,
			DefaultValue: "SGVsbG8=",
		}

		actual, err := ParseConfig(inlineConfig)

		assert.NoError(t, err)
		assert.Equal(t, *expected, *actual)
	})

	t.Run("Configuração inline do serviço SecretsManager da AWS", func(t *testing.T) {
		inlineConfig := "aws:secrets:/path/to/secret:json:SGVsbG8gd29ybGQ="

		expected := &types.Config{
			Provider:     types.AWS,
			Service:      types.SecretsManager,
			Name:         "/path/to/secret",
			Attribute:    "json",
			DefaultValue: "SGVsbG8gd29ybGQ=",
		}

		actual, err := ParseConfig(inlineConfig)

		assert.NoError(t, err)
		assert.Equal(t, *expected, *actual)
	})

	t.Run("Configuração inline do serviço S3 da AWS", func(t *testing.T) {
		inlineConfig := "aws:s3:my-bucket:file.txt:SGVsbG8="

		expected := &types.Config{
			Provider:     types.AWS,
			Service:      types.S3,
			Name:         "my-bucket",
			Attribute:    "file.txt",
			DefaultValue: "SGVsbG8=",
		}

		actual, err := ParseConfig(inlineConfig)

		assert.NoError(t, err)
		assert.Equal(t, *expected, *actual)
	})

	t.Run("Configuração inline inválida do serviço SSM da AWS", func(t *testing.T) {
		inlineConfig := "aws:ssm:invalid::SGVsbG8="

		_, err := ParseConfig(inlineConfig)
		assert.Error(t, err)
	})
}
