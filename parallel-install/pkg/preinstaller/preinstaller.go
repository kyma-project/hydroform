package preinstaller

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"io/ioutil"
	"k8s.io/client-go/dynamic"
	"os"
)

// PreInstaller prepares k8s cluster for Kyma installation.
type PreInstaller struct {
	applier       ResourceApplier
	cfg           config.Config
	dynamicClient dynamic.Interface
	retryOptions  []retry.Option
}

// File consists of a path to the file that was a part of PreInstaller installation
// and a component name that it belongs to.
type File struct {
	component string
	path      string
}

// Output contains lists of installed and not installed files during PreInstaller installation.
type Output struct {
	installed    []File
	notInstalled []File
}

type resourceInfoInput struct {
	dirSuffix                string
	resourceType             string
	installationResourcePath string
}

type resourceInfoResult struct {
	component    string
	name         string
	path         string
	resourceType string
}

// NewPreInstaller creates a new instance of PreInstaller.
func NewPreInstaller(applier ResourceApplier, cfg config.Config, dynamicClient dynamic.Interface, retryOptions []retry.Option) *PreInstaller {
	return &PreInstaller{
		applier:       applier,
		cfg:           cfg,
		dynamicClient: dynamicClient,
		retryOptions:  retryOptions,
	}
}

// InstallCRDs on a k8s cluster.
// Returns Output containing results of installation.
func (i *PreInstaller) InstallCRDs() (Output, error) {
	input := resourceInfoInput{
		resourceType:             "CRD",
		dirSuffix:                "crds",
		installationResourcePath: i.cfg.InstallationResourcePath,
	}

	output, err := i.install(input)
	if err != nil {
		return Output{}, err
	}

	return output, nil
}

// CreateNamespaces in a k8s cluster.
// Returns Output containing results of installation.
func (i *PreInstaller) CreateNamespaces() (Output, error) {
	input := resourceInfoInput{
		resourceType:             "Namespace",
		dirSuffix:                "namespaces",
		installationResourcePath: i.cfg.InstallationResourcePath,
	}

	output, err := i.install(input)
	if err != nil {
		return Output{}, err
	}

	return output, nil
}

func (i *PreInstaller) install(input resourceInfoInput) (o Output, err error) {
	resources, err := i.findResourcesIn(input)
	if err != nil {
		return Output{}, err
	}

	return i.apply(resources)
}

func (i *PreInstaller) findResourcesIn(input resourceInfoInput) (results []resourceInfoResult, err error) {
	installationResourcePath := input.installationResourcePath
	path := fmt.Sprintf("%s/%s", installationResourcePath, input.dirSuffix)
	rawComponentsDir, err := ioutil.ReadDir(path)
	if err != nil {
		return results, err
	}

	components := findOnlyDirectoriesAmong(rawComponentsDir)

	if components == nil || len(components) == 0 {
		i.cfg.Log.Warn("There were no components detected for installation. Skipping.")
		return results, nil
	}

	for _, component := range components {
		componentName := component.Name()
		pathToComponent := fmt.Sprintf("%s/%s", path, componentName)
		resources, err := ioutil.ReadDir(pathToComponent)
		if err != nil {
			return results, err
		}

		if len(resources) == 0 {
			i.cfg.Log.Warnf("There were no resources detected for component: ", componentName)
			break
		}

		for _, resource := range resources {
			resourceName := resource.Name()
			pathToResource := fmt.Sprintf("%s/%s", pathToComponent, resourceName)
			resourceInfoResult := resourceInfoResult{
				component:    componentName,
				name:         resourceName,
				path:         pathToResource,
				resourceType: input.resourceType,
			}

			results = append(results, resourceInfoResult)
		}
	}

	return results, nil
}

func (i *PreInstaller) apply(resources []resourceInfoResult) (o Output, err error) {
	for _, resource := range resources {
		file := File{
			component: resource.component,
			path:      resource.path,
		}

		i.cfg.Log.Info(fmt.Sprintf("Processing %s file: %s of component: %s", resource.resourceType, resource.name, resource.component))
		err = i.applier.Apply(file.path)
		if err != nil {
			i.cfg.Log.Warn(fmt.Sprintf("Error occurred when processing file %s of component %s : %s", resource.name, resource.component, err.Error()))
			o.notInstalled = append(o.notInstalled, file)
		} else {
			o.installed = append(o.installed, file)
		}
	}

	return o, nil
}

func findOnlyDirectoriesAmong(input []os.FileInfo) (o []os.FileInfo) {
	for _, item := range input {
		if item.IsDir() {
			o = append(o, item)
		}
	}

	return o
}
