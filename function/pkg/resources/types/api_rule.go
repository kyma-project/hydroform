package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type APIRuleSpec struct {
	Gateway string  `json:"gateway"`
	Rules   []Rule  `json:"rules"`
	Service Service `json:"service"`
}

type APIRule struct {
	APIVersion        string `json:"apiVersion"`
	Kind              string `json:"kind"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              APIRuleSpec `json:"spec"`
}

type Config struct {
	JwksUrls       []string `json:"jwks_urls,omitempty"`
	TrustedIssuers []string `json:"trusted_issuers,omitempty"`
	RequiredScope  []string `json:"required_scope,omitempty"`
}

type AccessStrategie struct {
	Config  *Config `json:"config,omitempty"`
	Handler string  `json:"handler"`
}

type Rule struct {
	AccessStrategies []AccessStrategie `json:"accessStrategies"`
	Methods          []string          `json:"methods"`
	Path             string            `json:"path"`
}
type Service struct {
	Host string `json:"host"`
	Name string `json:"name"`
	Port int64  `json:"port"`
}

func (ar APIRule) IsReference(name string) bool {
	return ar.Spec.Service.Name == name
}
