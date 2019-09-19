package types

// Provider holds the details for the provider in use.
type Provider struct {
	Type                 ProviderType           `json:"type"`
	ProjectName          string                 `json:"projectName"`
	CredentialsFilePath  string                 `json:"credentialsFilePath"`
	CustomConfigurations map[string]interface{} `json:"customConfigurations"`
}

// ProviderType shows the provider type.
type ProviderType string

const (
	// GCP is a possible value for ProviderType, standing for the Google Cloud Platform.
	GCP ProviderType = "gcp"
	// Azure is a possible value for ProviderType, standing for the Microsoft Azure Cloud Computing Platform.
	Azure ProviderType = "azure"
	// AWS is a possible value for ProviderType, standing for the Amazon Web Services.
	AWS ProviderType = "aws"
	// Gardener is a possible value for ProviderType, standing for the Gardener platform.
	Gardener ProviderType = "gardener"
)
