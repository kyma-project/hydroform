package types

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SourceType string

type FunctionSpec struct {
	Source     string                      `json:"source"`
	Deps       string                      `json:"deps,omitempty"`
	Runtime    Runtime                     `json:"runtime,omitempty"`
	Resources  corev1.ResourceRequirements `json:"resources,omitempty"`
	Labels     map[string]string           `json:"labels,omitempty"`
	Type       SourceType                  `json:"type,omitempty"`
	Repository `json:",inline,omitempty"`
}

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
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FunctionSpec `json:"spec,omitempty"`
}
