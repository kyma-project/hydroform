package workspace

import (
	"io"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"gopkg.in/yaml.v3"
)

var _ file = &Cfg{}

type Source = interface{}

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
	BaseDir        string `yaml:"baseDir"`
	SourceFileName string `yaml:"sourceFileName,omitempty"`
	DepsFileName   string `yaml:"depsFileName,omitempty"`
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
