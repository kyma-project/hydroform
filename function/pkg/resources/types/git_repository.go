package types

type GitRepository struct {
	Spec GitRepositorySpec `json:"spec,omitempty"`
}

type GitRepositorySpec struct {
	URL  string          `json:"url"`
	Auth *RepositoryAuth `json:"auth,omitempty"`
}

type RepositoryAuth struct {
	Type       string `json:"type"`
	SecretName string `json:"secretName"`
}
