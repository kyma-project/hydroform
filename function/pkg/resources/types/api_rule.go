package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type APIRule struct {
	APIVersion        string `json:"apiVersion"`
	Kind              string `json:"kind"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              APIRuleSpec `json:"spec"`
}

type APIRuleSpec struct {
	Gateway string  `json:"gateway"`
	Host    string  `json:"host"`
	Service Service `json:"service"`
	Rules   []Rule  `json:"rules"`
}

type Service struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Port      int64  `json:"port"`
}

type Rule struct {
	Path             string            `json:"path"`
	Methods          []string          `json:"methods"`
	AccessStrategies []AccessStrategie `json:"accessStrategies"`
}

type AccessStrategie struct {
	Config  *Config `json:"config,omitempty"`
	Handler string  `json:"handler"`
}

type Config struct {
	JwksUrls       []string `json:"jwks_urls,omitempty"`
	TrustedIssuers []string `json:"trusted_issuers,omitempty"`
	RequiredScope  []string `json:"required_scope,omitempty"`
}

func (ar APIRule) IsReference(name string) bool {
	return ar.Spec.Service.Name == name
}
