package unstructured

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"path"

	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ReadFile = func(filename string) ([]byte, error)

const functionApiVersion = "serverless.kyma-project.io/v1alpha1"

var errUnsupportedSource = fmt.Errorf("unsupported source")

func NewFunction(cfg workspace.Cfg) (unstructured.Unstructured, error) {
	switch cfg.Source.Type() {
	case workspace.SourceTypeInline:
		return newFunction(cfg, ioutil.ReadFile)
	case workspace.SourceTypeGit:
		return newGitFunction(cfg)
	default:
		return unstructured.Unstructured{}, errUnsupportedSource
	}
}

func decorators(cfg workspace.Cfg) []Decorate {
	return []Decorate{
		withFunction(cfg.Name, cfg.Namespace),
		withLabels(cfg.Labels),
		withRuntime(cfg.Runtime),
		withLimits(cfg.Resources.Limits),
		withRequests(cfg.Resources.Requests),
	}
}

func newGitFunction(cfg workspace.Cfg) (out unstructured.Unstructured, err error) {
	var source workspace.SourceGit
	if err = mapstructure.Decode(cfg.Source, &source); err != nil {
		return
	}

	decorators := append(decorators(cfg),
		withRepository(source.Reference),
	)
	err = decorate(&out, decorators)

	return
}

func newFunction(cfg workspace.Cfg, readFile ReadFile) (out unstructured.Unstructured, err error) {
	var source workspace.SourceInline
	if err = mapstructure.Decode(cfg.Source, &source); err != nil {
		return
	}

	// get default handler names
	sourceHandlerName, depsHandlerName, found := workspace.InlineFileNames(cfg.Runtime)
	if !found {
		return unstructured.Unstructured{}, fmt.Errorf("'%s' invalid runtime", cfg.Runtime)
	}

	// apply source handler name overrides
	if source.SourceHandlerName != "" {
		sourceHandlerName = source.SourceHandlerName
	}

	// apply deps handler name overrides
	if source.DepsHandlerName != "" {
		depsHandlerName = source.DepsHandlerName
	}

	decorators := decorators(cfg)

	// read sources and dependencies
	for _, item := range []struct {
		property property
		filename string
	}{
		{property: propertySource, filename: sourceHandlerName},
		{property: propertyDeps, filename: depsHandlerName},
	} {
		filePath := path.Join(source.BaseDir, item.filename)
		data, err := readFile(filePath)
		if err != nil {
			return unstructured.Unstructured{}, err
		}
		if len(data) == 0 {
			continue
		}
		decorators = append(decorators, decorateWithField(string(data), "spec", string(item.property)))
	}

	err = decorate(&out, decorators)
	return
}

type property string

const (
	propertySource property = "source"
	propertyDeps   property = "deps"
)
