package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ApiRuleSpec struct {
	Gateway string  `json:"gateway"`
	Rules   []Rules `json:"rules"`
	Service Service `json:"service"`
}

type ApiRule struct {
	ApiVersion        string `json:"apiVersion"`
	Kind              string `json:"kind"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ApiRuleSpec `json:"spec"`
}

type Config struct {
	JwksUrls       []string `json:"jwks_urls,omitempty"`
	TrustedIssuers []string `json:"trusted_issuers,omitempty"`
}

type AccessStrategies struct {
	Config  Config `json:"config,omitempty"`
	Handler string `json:"handler"`
}
type Rules struct {
	AccessStrategies []AccessStrategies `json:"accessStrategies"`
	Methods          []string           `json:"methods"`
	Path             string             `json:"path"`
}
type Service struct {
	Host string `json:"host"`
	Name string `json:"name"`
	Port int64  `json:"port"`
}
