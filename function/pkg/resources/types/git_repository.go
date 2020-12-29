package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type GitRepository struct {
	ApiVersion        string
	Kind              string
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GitRepositorySpec `json:"spec,omitempty"`
}

type GitRepositorySpec struct {
	URL  string          `json:"url"`
	Auth *RepositoryAuth `json:"auth,omitempty"`
}

type RepositoryAuth struct {
	Type       string `json:"type"`
	SecretName string `json:"secretName"`
}
