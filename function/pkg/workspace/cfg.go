package workspace

import (
	"io"

	"github.com/kyma-project/hydroform/function/pkg/resources/types"
	"gopkg.in/yaml.v3"
)

var _ File = &Cfg{}

type SourceType string

const (
	SourceTypeInline SourceType = "inline"
	SourceTypeGit    SourceType = "git"
)

const CfgFilename = "config.yaml"

type EventType struct {
	Property string `yaml:"property" jsonschema:"default=type"`
	Type     string `yaml:"type" jsonschema:"default=exact"`
	Value    string `yaml:"value"`
}

type EventSource struct {
	Property string `yaml:"property" jsonschema:"default=source"`
	Type     string `yaml:"type" jsonschema:"default=exact"`
	Value    string `yaml:"value" jsonschema:"default=\"\""`
}

type EventFilter struct {
	EventSource EventSource `yaml:"eventSource"`
	EventType   EventType   `yaml:"eventType"`
}

type Filter struct {
	Dialect string        `yaml:"dialect,omitempty"`
	Filters []EventFilter `yaml:"filters"`
}

type Subscription struct {
	Name     string `yaml:"name,omitempty"`
	Protocol string `yaml:"protocol" jsonschema:"default=\"\""`
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

type APIRule struct {
	Name    string  `yaml:"name,omitempty"`
	Gateway string  `yaml:"gateway,omitempty"`
	Service Service `yaml:"service"`
	Rules   []Rule  `yaml:"rules"`
}

type Service struct {
	Host string `yaml:"host"`
	Port int64  `yaml:"port,omitempty"`
}

type Rule struct {
	Path             string            `yaml:"path,omitempty"`
	Methods          []string          `yaml:"methods"`
	AccessStrategies []AccessStrategie `yaml:"accessStrategies"`
}

type AccessStrategie struct {
	Config  AccessStrategieConfig `yaml:"config,omitempty"`
	Handler string                `yaml:"handler" jsonschema:"enum=oauth2_introspection,enum=jwt,enum=noop,enum=allow,default=allow"`
}

type AccessStrategieConfig struct {
	JwksUrls       []string `yaml:"jwksUrls,omitempty"`
	TrustedIssuers []string `yaml:"trustedIssuers,omitempty"`
	RequiredScope  []string `yaml:"requiredScope,omitempty"`
}

type Cfg struct {
	Name                 string            `yaml:"name"`
	Namespace            string            `yaml:"namespace"`
	Labels               map[string]string `yaml:"labels,omitempty"`
	Runtime              types.Runtime     `yaml:"runtime" jsonschema:"enum=nodejs12,enum=nodejs14,enum=python39"`
	RuntimeImageOverride string            `yaml:"runtimeImageOverride,omitempty"`
	Source               Source            `yaml:"source"`
	Resources            Resources         `yaml:"resource,omitempty"`
	Subscriptions        []Subscription    `yaml:"subscriptions,omitempty"`
	Env                  []EnvVar          `yaml:"env,omitempty"`
	APIRules             []APIRule         `yaml:"apiRules,omitempty"`
}

type Source struct {
	Type         SourceType `yaml:"sourceType" jsonschema:"enum=inline,enum=git"`
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
	CredentialsType       string `yaml:"credentialsType,omitempty"`
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

func (cfg Cfg) Write(writer io.Writer, _ interface{}) error {
	return yaml.NewEncoder(writer).Encode(&cfg)
}

func (cfg Cfg) FileName() string {
	return CfgFilename
}
