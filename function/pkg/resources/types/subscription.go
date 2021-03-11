package types

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Subscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SubscriptionSpec `json:"spec"`
}

type SubscriptionSpec struct {
	Filter           Filter           `json:"filter"`
	ID               string           `json:"id,omitempty"`
	Protocol         string           `json:"protocol"`
	ProtocolSettings ProtocolSettings `json:"protocolsettings"`
	Sink             string           `json:"sink"`
}

type Filter struct {
	Dialect string      `json:"dialect,omitempty"`
	Filters []BebFilter `json:"filters"`
}

type BebFilter struct {
	EventSource Event `json:"eventSource"`
	EventType   Event `json:"eventType"`
}

type Event struct {
	Property string `json:"property"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value"`
}

type ProtocolSettings struct {
	ContentMode     string      `json:"contentMode"`
	ExemptHandshake bool        `json:"exemptHandshake"`
	Qos             string      `json:"qos"`
	WebhookAuth     WebhookAuth `json:"webhookAuth"`
}

type WebhookAuth struct {
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	GrantType    string   `json:"grantType"`
	Scope        []string `json:"scope"`
	TokenURL     string   `json:"tokenUrl"`
	Type         string   `json:"type"`
}

func (s Subscription) IsReference(name, namespace string) bool {
	expectedSinkName := fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace)
	return expectedSinkName == s.Spec.Sink
}


