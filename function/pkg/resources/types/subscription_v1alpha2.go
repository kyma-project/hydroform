package types

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SubscriptionV1alpha2 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SubscriptionSpecV1alpha2 `json:"spec"`
}

type SubscriptionSpecV1alpha2 struct {
	ID           string   `json:"id,omitempty"`
	Sink         string   `json:"sink"`
	TypeMatching string   `json:"typeMatching,omitempty"`
	EventSource  string   `json:"source"`
	Types        []string `json:"types"`
}

func (s SubscriptionV1alpha2) IsReference(name, namespace string) bool {
	expectedSinkName := fmt.Sprintf("http://%s.%s.svc.cluster.local", name, namespace)
	return expectedSinkName == s.Spec.Sink
}
