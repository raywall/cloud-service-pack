package types

type (
	ServiceType       string
	CloudProviderType string
)

const (
	AWS   CloudProviderType = "aws"
	GCP   CloudProviderType = "gcp"
	AZURE CloudProviderType = "azure"
	LOCAL CloudProviderType = "local"
)

const (
	S3             ServiceType = "s3"
	SSM            ServiceType = "ssm"
	SecretsManager ServiceType = "secrets"
	ENV            ServiceType = "env"
	FILE           ServiceType = "file"
)

type Config struct {
	Provider     CloudProviderType `json:"provider"`
	Service      ServiceType       `json:"service"`
	Name         string            `json:"name"`
	Attribute    interface{}       `json:"argument"`
	DefaultValue string            `json:"defaultValue"`
}
