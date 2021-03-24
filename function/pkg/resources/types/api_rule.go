package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ApiRuleSpec struct {
	Gateway string  `yaml:"gateway"`
	Rules   []Rules `yaml:"rules"`
	Service Service `yaml:"service"`
}

type ApiRule struct {
	ApiVersion        string `json:"apiVersion"`
	Kind              string `json:"kind"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ApiRuleSpec `json:"spec"`
}

type Config struct {
	JwksUrls       []string `json:"jwks_urls"`
	TrustedIssuers []string `json:"trusted_issuers"`
}

type AccessStrategies struct {
	Config  Config `yaml:"config"`
	Handler string `yaml:"handler"`
}
type Rules struct {
	AccessStrategies []AccessStrategies `yaml:"accessStrategies"`
	Methods          []string           `yaml:"methods"`
	Path             string             `yaml:"path"`
}
type Service struct {
	Host string `yaml:"host"`
	Name string `yaml:"name"`
	Port string `yaml:"port"`
}
