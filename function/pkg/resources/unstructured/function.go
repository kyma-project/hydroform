package unstructured

import (
	"fmt"
	"io/ioutil"
	"path"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ReadFile = func(filename string) ([]byte, error)

const functionApiVersion = "serverless.kyma-project.io/v1alpha1"

var errUnsupportedSource = fmt.Errorf("unsupported source")

func NewFunction(cfg workspace.Cfg) (unstructured.Unstructured, error) {
	switch cfg.Source.Type {
	case workspace.SourceTypeInline:
		return newFunction(cfg, ioutil.ReadFile)
	case workspace.SourceTypeGit:
		return newGitFunction(cfg)
	default:
		return unstructured.Unstructured{}, errUnsupportedSource
	}
}

func newGitFunction(cfg workspace.Cfg) (out unstructured.Unstructured, err error) {
	repository := cfg.Name
	if cfg.Source.Repository != "" {
		repository = cfg.Source.Repository
	}

	f, err := prepareBaseFunction(cfg)
	if err != nil {
		return unstructured.Unstructured{}, err
	}

	f.Spec.Type = "git"
	f.Spec.Source = repository
	f.Spec.Reference = cfg.Source.Reference
	f.Spec.BaseDir = cfg.Source.BaseDir

	unstructuredFunction, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&f)
	out = unstructured.Unstructured{Object: unstructuredFunction}
	if err != nil {
		return unstructured.Unstructured{}, err
	}

	return
}

func newFunction(cfg workspace.Cfg, readFile ReadFile) (out unstructured.Unstructured, err error) {
	// get default handler names
	sourceHandlerName, depsHandlerName, found := workspace.InlineFileNames(cfg.Runtime)
	if !found {
		return unstructured.Unstructured{}, fmt.Errorf("'%s' invalid runtime", cfg.Runtime)
	}

	// apply source handler name overrides
	if cfg.Source.SourceHandlerName != "" {
		sourceHandlerName = cfg.Source.SourceHandlerName
	}

	// apply deps handler name overrides
	if cfg.Source.DepsHandlerName != "" {
		depsHandlerName = cfg.Source.DepsHandlerName
	}

	f, err := prepareInlineFunction(cfg, readFile, sourceHandlerName, depsHandlerName)
	if err != nil {
		return unstructured.Unstructured{}, err
	}

	unstructuredFunction, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&f)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	out = unstructured.Unstructured{Object: unstructuredFunction}
	return
}

func prepareInlineFunction(cfg workspace.Cfg, readFile ReadFile, sourceHandlerName workspace.SourceFileName, depsHandlerName workspace.DepsFileName) (types.Function, error) {
	specSource, err := prepareFunctionSource(cfg, readFile, sourceHandlerName)
	if err != nil {
		return types.Function{}, err
	}

	specDeps, err := prepareFunctionDeps(cfg, readFile, depsHandlerName)
	if err != nil {
		return types.Function{}, err
	}

	f, err := prepareBaseFunction(cfg)
	if err != nil {
		return types.Function{}, err
	}

	f.Spec.Source = string(specSource)
	f.Spec.Deps = string(specDeps)

	return f, nil
}

func prepareBaseFunction(cfg workspace.Cfg) (types.Function, error) {
	resources, err := prepareFunctionResources(cfg)
	if err != nil {
		return types.Function{}, err
	}

	envs := prepareEnvVars(cfg.Env)

	f := types.Function{
		ApiVersion: functionApiVersion,
		Kind:       "Function",
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
		},
		Spec: types.FunctionSpec{
			Runtime:    cfg.Runtime,
			Resources:  resources,
			Labels:     cfg.Labels,
			Repository: types.Repository{},
			Env:        envs,
		},
	}
	return f, nil
}

func prepareFunctionSource(cfg workspace.Cfg, readFile ReadFile, sourceHandlerName workspace.SourceFileName) ([]byte, error) {
	specSource, err := readFile(path.Join(cfg.Source.SourcePath, sourceHandlerName))
	if err != nil {
		return nil, err
	}

	return specSource, nil
}

func prepareFunctionDeps(cfg workspace.Cfg, readFile ReadFile, depsHandlerName workspace.DepsFileName) ([]byte, error) {
	specDeps, err := readFile(path.Join(cfg.Source.SourcePath, depsHandlerName))
	if err != nil {
		return nil, err
	}
	return specDeps, nil
}

func prepareFunctionResources(cfg workspace.Cfg) (v1.ResourceRequirements, error) {

	limitsCPU, err := resource.ParseQuantity(cfg.Resources.Limits[workspace.ResourceNameCPU].(string))
	if err != nil {
		return v1.ResourceRequirements{}, err
	}

	limitsMemory, err := resource.ParseQuantity(cfg.Resources.Limits[workspace.ResourceNameMemory].(string))
	if err != nil {
		return v1.ResourceRequirements{}, err
	}

	requestsCPU, err := resource.ParseQuantity(cfg.Resources.Requests[workspace.ResourceNameCPU].(string))
	if err != nil {
		return v1.ResourceRequirements{}, err
	}

	requestsMemory, err := resource.ParseQuantity(cfg.Resources.Requests[workspace.ResourceNameMemory].(string))
	if err != nil {
		return v1.ResourceRequirements{}, err
	}

	resources := v1.ResourceRequirements{
		Limits: map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU:    limitsCPU,
			v1.ResourceMemory: limitsMemory,
		},
		Requests: map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU:    requestsCPU,
			v1.ResourceMemory: requestsMemory,
		},
	}

	return resources, nil
}

func prepareEnvVars(envs []workspace.EnvVar) []v1.EnvVar {
	newEnvs := make([]v1.EnvVar, 0)

	for _, envVar := range envs {
		newEnv := v1.EnvVar{
			Name:  envVar.Name,
			Value: envVar.Value,
		}

		if envVar.ValueFrom != nil {
			newEnv.ValueFrom = &v1.EnvVarSource{}

			if envVar.ValueFrom.SecretKeyRef != nil {
				newEnv.ValueFrom.SecretKeyRef = &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: envVar.ValueFrom.SecretKeyRef.Name,
					},
					Key: envVar.ValueFrom.SecretKeyRef.Key,
				}
			}

			if envVar.ValueFrom.ConfigMapKeyRef != nil {
				newEnv.ValueFrom.ConfigMapKeyRef = &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: envVar.ValueFrom.ConfigMapKeyRef.Name,
					},
					Key: envVar.ValueFrom.ConfigMapKeyRef.Key,
				}
			}
		}

		newEnvs = append(newEnvs, newEnv)
	}
	return newEnvs
}
