package types

type (
	CloudProviderType string
	ServiceType       string
)

const (
	AWS   CloudProviderType = "aws"
	GCP   CloudProviderType = "gcp"
	AZURE CloudProviderType = "azure"
)

const (
	S3             ServiceType = "s3"
	SSM            ServiceType = "ssm"
	SecretsManager ServiceType = "secrets"
)

type Config struct {
	Provider     CloudProviderType `json:"provider"`
	Service      ServiceType       `json:"service"`
	Name         string            `json:"name"`
	Attribute    interface{}       `json:"argument"`
	DefaultValue string            `json:"defaultValue"`
}
