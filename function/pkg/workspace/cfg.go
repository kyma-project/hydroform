package workspace

import (
	"io"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"gopkg.in/yaml.v3"
)

var _ file = &Cfg{}

type SourceType string

const (
	SourceTypeInline SourceType = "inline"
	SourceTypeGit    SourceType = "git"
)

const CfgFilename = "config.yaml"

type EventFilterProperty struct {
	Property string `yaml:"property"`
	Type     string `yaml:"type,omitempty"`
	Value    string `yaml:"value"`
}

type EventFilter struct {
	EventSource EventFilterProperty `yaml:"eventSource"`
	EventType   EventFilterProperty `yaml:"eventType"`
}

type Filter struct {
	Dialect string        `yaml:"dialect,omitempty"`
	Filters []EventFilter `yaml:"filters"`
}

type Subscription struct {
	Name     string `yaml:"name"`
	Protocol string `yaml:"protocol"`
	Filter   Filter `yaml:"filter"`
}

type Resources struct {
	Limits   ResourceList `yaml:"limits,omitempty"`
	Requests ResourceList `yaml:"requests,omitempty"`
}

type EnvVar struct {
	Name      string        `yaml:"name"`
	Value     string        `yaml:"value,omitempty"`
	ValueFrom *EnvVarSource `yaml:"valueFrom,omitempty"`
}

type EnvVarSource struct {
	ConfigMapKeyRef *ConfigMapKeySelector `yaml:"configMapKeyRef,omitempty"`
	SecretKeyRef    *SecretKeySelector    `yaml:"secretKeyRef,omitempty"`
}

type ConfigMapKeySelector struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type SecretKeySelector struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type ApiRule struct {
	Name    string  `yaml:"name,omitempty"`
	Gateway string  `yaml:"gateway,omitempty"`
	Service Service `yaml:"service,omitempty"`
	Rules   []Rule  `yaml:"rules,omitempty"`
}

type Service struct {
	Host string `yaml:"host"`
	Port int64  `yaml:"port"`
}

type Rule struct {
	Path             string            `yaml:"path"`
	Methods          []string          `yaml:"methods"`
	AccessStrategies []AccessStrategie `yaml:"accessStrategies"`
}

type AccessStrategie struct {
	Config  AccessStrategieConfig `json:"config,omitempty"`
	Handler string                `json:"handler"`
}

type AccessStrategieConfig struct {
	JwksUrls       []string `yaml:"jwksUrls"`
	TrustedIssuers []string `yaml:"trustedIssuers"`
	RequiredScope  []string `yaml:"requiredScope"`
}

type Cfg struct {
	Name          string            `yaml:"name"`
	Namespace     string            `yaml:"namespace"`
	Labels        map[string]string `yaml:"labels,omitempty"`
	Runtime       types.Runtime     `yaml:"runtime"`
	Source        Source            `yaml:"source"`
	Resources     Resources         `yaml:"resource,omitempty"`
	Subscriptions []Subscription    `yaml:"subscriptions,omitempty"`
	Env           []EnvVar          `yaml:"env,omitempty"`
	ApiRules      []ApiRule         `yaml:"apiRules,omitempty"`
}

type Source struct {
	Type         SourceType `yaml:"sourceType"`
	SourceInline `yaml:",inline"`
	SourceGit    `yaml:",inline"`
}

type SourceInline struct {
	SourcePath        string `yaml:"sourcePath,omitempty"`
	SourceHandlerName string `yaml:"sourceHandlerName,omitempty"`
	DepsHandlerName   string `yaml:"depsHandlerName,omitempty"`
}

func (s SourceInline) Type() SourceType {
	return SourceTypeInline
}

type SourceGit struct {
	URL                   string `yaml:"url,omitempty"`
	Repository            string `yaml:"repository,omitempty"`
	Reference             string `yaml:"reference,omitempty"`
	BaseDir               string `yaml:"baseDir,omitempty"`
	CredentialsSecretName string `yaml:"credentialsSecretName,omitempty"`
}

func (s SourceGit) Type() SourceType {
	return SourceTypeGit
}

type ResourceList = map[ResourceName]interface{}

type ResourceName = string

const (
	ResourceNameCPU    ResourceName = "cpu"
	ResourceNameMemory ResourceName = "memory"
)

func (cfg Cfg) write(writer io.Writer, _ interface{}) error {
	return yaml.NewEncoder(writer).Encode(&cfg)
}

func (cfg Cfg) fileName() string {
	return CfgFilename
}
