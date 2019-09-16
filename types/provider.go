package types

type Provider struct {
	Type                 ProviderType
	ProjectName          string
	CredentialsFilePath  string
	CustomConfigurations map[string]interface{}
}

type ProviderType string

const (
	GCP   ProviderType = "google"
	Azure ProviderType = "azure"
	AWS   ProviderType = "aws"
)
