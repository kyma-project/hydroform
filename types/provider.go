package types

type Provider struct {
	Type                 ProviderType
	ProjectName          string
	CredentialsFilePath  string
	CustomConfigurations map[string]interface{}
}

type ProviderType string

const (
	GCP      ProviderType = "gcp"
	Azure    ProviderType = "azure"
	AWS      ProviderType = "aws"
	Gardener ProviderType = "gardener"
)
