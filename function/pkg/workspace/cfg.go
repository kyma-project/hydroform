package workspace

import (
	"io"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"gopkg.in/yaml.v3"
)

var _ file = &Cfg{}

type SourceType int

const (
	SourceTypeInline SourceType = iota + 1
	SourceTypeGit
)

type Source interface {
	Type() SourceType
}

const CfgFilename = "config.yaml"

type Cfg struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels,omitempty"`

	Runtime types.Runtime `yaml:"runtime"`
	Source  Source        `yaml:"source"`

	Resources struct {
		Limits   ResourceList `yaml:"limits"`
		Requests ResourceList `yaml:"requests"`
	} `yaml:"resource,omitempty"`

	Triggers []struct {
		EventTypeVersion string `yaml:"eventTypeVersion"`
		Source           string `yaml:"source"`
		Type             string `yaml:"type"`
	} `yaml:"triggers,omitempty"`
}

type SourceInline struct {
	BaseDir           string `yaml:"baseDir"`
	SourceHandlerName string `yaml:"sourceHandlerName,omitempty"`
	DepsHandlerName   string `yaml:"depsHandlerName,omitempty"`
}

func (s SourceInline) Type() SourceType {
	return SourceTypeInline
}

type SourceGit struct {
	URL                   string `yaml:"url"`
	Repository            string `yaml:"repository"`
	Reference             string `yaml:"reference"`
	BaseDir               string `yaml:"baseDir"`
	CredentialsSecretName string `yaml:"credentialsSecretName"`
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
