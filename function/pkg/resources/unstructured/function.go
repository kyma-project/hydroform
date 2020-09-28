package unstructured

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ReadFile = func(filename string) ([]byte, error)

const (
	functionApiVersion = "serverless.kyma-project.io/v1alpha1"
)

func NewFunction(cfg workspace.Cfg) (unstructured.Unstructured, error) {
	return newFunction(cfg, ioutil.ReadFile)
}

var (
	//FIXME extend later
	errInvalidSourceType = fmt.Errorf("invalid source type")
)

//FIXME readfile powinno być callbackiem
func newFunction(cfg workspace.Cfg, readFile ReadFile) (out unstructured.Unstructured, err error) {
	source, ok := cfg.Source.(workspace.SourceInline)
	if !ok {
		return unstructured.Unstructured{}, errInvalidSourceType
	}

	decorators := []Decorate{
		withFunction(cfg.Name, cfg.Namespace),
		withLabels(cfg.Labels),
		withRuntime(cfg.Runtime),
		withLimits(cfg.Resources.Limits),
		withRequests(cfg.Resources.Requests),
	}

	for _, item := range []struct {
		property property
		filename string
	}{
		{property: propertySource, filename: source.SourceFileName},
		{property: propertyDeps, filename: source.DepsFileName},
	} {
		//FIXME to ma dawać callback
		filePath := path.Join(source.BaseDir, item.filename)
		data, err := readFile(filePath)
		if err != nil {
			return unstructured.Unstructured{}, err
		}
		if len(data) == 0 {
			continue
		}
		decorateWithField(string(data), "spec", string(item.property))
	}

	err = decorate(&out, decorators)
	return
}

type property string

const (
	propertySource property = "source"
	propertyDeps   property = "deps"
)

var (
	runtimeMappings = map[types.Runtime]map[property]workspace.FileName{
		types.Nodejs12: {
			propertySource: workspace.FileNameHandlerJs,
			propertyDeps:   workspace.FileNamePackageJSON,
		},
		types.Nodejs10: {
			propertySource: workspace.FileNameHandlerJs,
			propertyDeps:   workspace.FileNamePackageJSON,
		},
		types.Python38: {
			propertySource: workspace.FileNameHandlerPy,
			propertyDeps:   workspace.FileNameRequirementsTxt,
		},
	}
)
