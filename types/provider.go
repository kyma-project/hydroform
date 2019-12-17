package types

// Provider specifies the provider-related information Hydroform needs to perform its tasks.
type Provider struct {
	// Type specifies the cloud provider to use.
	Type ProviderType `json:"type"`
	// ProjectName specifies the project the cluster will be created in.
	// In the case of Azure, it represents the resource group.
	ProjectName string `json:"projectName"`
	// CredentialsFilePath specifies the path to credentials used to access the cloud provider.
	CredentialsFilePath string `json:"credentialsFilePath"`
	// CustomConfigurations is a list of custom properties relevant for the chosen provider.
	CustomConfigurations map[string]interface{} `json:"customConfigurations"`
}

// ProviderType lists available cloud providers.
type ProviderType string

const (
	// GCP stands for the Google Cloud Platform.
	GCP ProviderType = "gcp"
	// Azure stands for the Microsoft Azure Cloud Computing Platform.
	Azure ProviderType = "azure"
	// AWS stands for Amazon Web Services.
	AWS ProviderType = "aws"
	// Gardener stands for the Gardener platform.
	Gardener ProviderType = "gardener"
)
