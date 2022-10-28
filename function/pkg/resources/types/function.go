package types

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FunctionSpec struct {
	Source               Source                       `json:"source"`
	Deps                 string                       `json:"deps,omitempty"`
	Runtime              Runtime                      `json:"runtime,omitempty"`
	RuntimeImageOverride string                       `json:"runtimeImageOverride,omitempty"`
	Resources            *corev1.ResourceRequirements `json:"resources,omitempty"`
	Labels               map[string]string            `json:"labels,omitempty"`
	Env                  []corev1.EnvVar              `json:"env,omitempty"`
}

type Source struct {
	GitRepository *GitRepositorySource `json:"gitRepository,omitempty"`
	Inline        *InlineSource        `json:"inline,omitempty"`
}

type GitRepositorySource struct {
	URL        string          `json:"url"`
	Auth       *RepositoryAuth `json:"auth,omitempty"`
	Repository `json:",inline"`
}

type InlineSource struct {
	Source       string `json:"source"`
	Dependencies string `json:"dependencies,omitempty"`
}

type RepositoryAuth struct {
	Type       RepositoryAuthType `json:"type"`
	SecretName string             `json:"secretName"`
}

type RepositoryAuthType string

const (
	RepositoryAuthBasic  RepositoryAuthType = "basic"
	RepositoryAuthSSHKey RepositoryAuthType = "key"
)

func (s FunctionSpec) toMap(l corev1.ResourceList) map[string]interface{} {
	length := len(l)
	if length == 0 {
		return nil
	}

	result := make(map[string]interface{}, length)
	for name, quantity := range l {
		result[name.String()] = quantity.String()
	}

	return result
}

func (s FunctionSpec) ResourceLimits() map[string]interface{} {
	return s.toMap(s.Resources.Limits)
}

func (s FunctionSpec) ResourceRequests() map[string]interface{} {
	return s.toMap(s.Resources.Requests)
}

type Function struct {
	APIVersion        string `json:"apiVersion"`
	Kind              string
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FunctionSpec `json:"spec,omitempty"`
}
