package types

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SubscriptionV1alpha1 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SubscriptionSpecV1alpha1 `json:"spec"`
}

type SubscriptionSpecV1alpha1 struct {
	Filter           Filter            `json:"filter"`
	ID               string            `json:"id,omitempty"`
	Protocol         string            `json:"protocol"`
	ProtocolSettings *ProtocolSettings `json:"protocolsettings"`
	Sink             string            `json:"sink"`
}

type Filter struct {
	Dialect string        `json:"dialect,omitempty"`
	Filters []EventFilter `json:"filters"`
}

type EventFilter struct {
	EventSource EventFilterProperty `json:"eventSource"`
	EventType   EventFilterProperty `json:"eventType"`
}

type EventFilterProperty struct {
	Property string `json:"property"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value"`
}

type ProtocolSettings struct {
	ContentMode     string       `json:"contentMode,omitempty"`
	ExemptHandshake bool         `json:"exemptHandshake,omitempty"`
	Qos             string       `json:"qos,omitempty"`
	WebhookAuth     *WebhookAuth `json:"webhookAuth,omitempty"`
}

type WebhookAuth struct {
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	GrantType    string   `json:"grantType"`
	Scope        []string `json:"scope,omitempty"`
	TokenURL     string   `json:"tokenUrl"`
	Type         string   `json:"type"`
}

func (s SubscriptionV1alpha1) IsReference(name, namespace string) bool {
	expectedSinkName := fmt.Sprintf("http://%s.%s.svc.cluster.local", name, namespace)
	return expectedSinkName == s.Spec.Sink
}
