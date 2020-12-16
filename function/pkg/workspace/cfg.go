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

type Attributes = map[string]interface{}

type Trigger struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Source  string `yaml:"source"`
	Type    string `yaml:"type"`
}

func (t Trigger) Attributes() Attributes {
	return map[string]interface{}{
		"eventtypeversion": t.Version,
		"source":           t.Source,
		"type":             t.Type,
	}
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

func (e EnvVar) ConvertToMapStringInterface() map[string]interface{} {
	if e.ValueFrom != nil {
		if e.ValueFrom.ConfigMapKeyRef != nil {
			return map[string]interface{}{
				"name":  e.Name,
				"value": e.Value,
				"valueFrom": map[string]interface{}{
					"configMapKeyRef": map[string]interface{}{
						"name": e.ValueFrom.ConfigMapKeyRef.Name,
						"key":  e.ValueFrom.ConfigMapKeyRef.Key,
					},
				},
			}
		} else {
			return map[string]interface{}{
				"name":  e.Name,
				"value": e.Value,
				"valueFrom": map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": e.ValueFrom.SecretKeyRef.Name,
						"key":  e.ValueFrom.SecretKeyRef.Key,
					},
				},
			}
		}
	} else {
		return map[string]interface{}{
			"name":  e.Name,
			"value": e.Value,
		}
	}
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

type Cfg struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels,omitempty"`
	Runtime   types.Runtime     `yaml:"runtime"`
	Source    Source            `yaml:"source"`
	Resources Resources         `yaml:"resource,omitempty"`
	Triggers  []Trigger         `yaml:"triggers,omitempty"`
	Env       []EnvVar          `yaml:"env,omitempty"`
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
