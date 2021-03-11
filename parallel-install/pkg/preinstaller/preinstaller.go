// Package `preinstaller` implements the logic related to preparing a k8s cluster for Kyma installation.
// It installs provided resources (or upgrades if necessary).
//
// The code in the package uses the user-provided function for logging and installation resources path.
// Resources should be organized in the following way:
// <provided-path>
//	crds
//		component-fileName-1
//			file-1
//			file-2
//			...
//			file-n
//		component-fileName-2
//			...
//		...
//		component-fileName-n
// namespaces
// ...
// Installing CRDs resources requires a folder named `crds`.
// Installing Namespace resources requires a folder named `namespaces`.
// For now only these two resources types are supported.

package preinstaller

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"io/ioutil"
	"k8s.io/client-go/dynamic"
	"os"
)

// Config defines configuration values for the PreInstaller.
type Config struct {
	InstallationResourcePath string           //Path to the installation resources.
	Log                      logger.Interface //Logger to be used
}

// PreInstaller prepares k8s cluster for Kyma installation.
type PreInstaller struct {
	applier       ResourceApplier
	parser        ResourceParser
	cfg           Config
	dynamicClient dynamic.Interface
	retryOptions  []retry.Option
}

// File consists of a path to the file that was a part of PreInstaller installation
// and a component fileName that it belongs to.
type File struct {
	component string
	path      string
}

// Output contains lists of Installed and not Installed files during PreInstaller installation.
type Output struct {
	// Installed files during PreInstaller installation.
	Installed []File
	// NotInstalled files during PreInstaller installation.
	NotInstalled []File
}

type resourceInfoInput struct {
	dirSuffix                string
	resourceType             string
	installationResourcePath string
}

type resourceInfoResult struct {
	component    string
	fileName     string
	path         string
	resourceType string
}

// NewPreInstaller creates a new instance of PreInstaller.
func NewPreInstaller(applier ResourceApplier, parser ResourceParser, cfg Config, dynamicClient dynamic.Interface, retryOptions []retry.Option) *PreInstaller {
	return &PreInstaller{
		applier:       applier,
		parser:        parser,
		cfg:           cfg,
		dynamicClient: dynamicClient,
		retryOptions:  retryOptions,
	}
}

// InstallCRDs on a k8s cluster.
// Returns Output containing results of installation.
func (i *PreInstaller) InstallCRDs() (Output, error) {
	input := resourceInfoInput{
		resourceType:             "CustomResourceDefinition",
		dirSuffix:                "crds",
		installationResourcePath: i.cfg.InstallationResourcePath,
	}

	i.cfg.Log.Info("Kyma CRDs installation")
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

	i.cfg.Log.Info("Kyma Namespaces creation")
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
				fileName:     resourceName,
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

		parsedResource, err := i.parser.ParseUnstructuredResourceFrom(file.path)
		if err != nil {
			i.cfg.Log.Warnf("Error occurred when processing resource %s of component %s : %s", resource.fileName, resource.component, err.Error())
			o.NotInstalled = append(o.NotInstalled, file)
			continue
		}

		if parsedResource.GetKind() != resource.resourceType {
			i.cfg.Log.Warnf("Resource type does not match for resource %s of component %s : got %s but expected %s", resource.fileName, resource.component, parsedResource.GroupVersionKind().Kind, resource.resourceType)
			o.NotInstalled = append(o.NotInstalled, file)
			continue
		}

		i.cfg.Log.Infof("Processing %s file: %s of component: %s", resource.resourceType, resource.fileName, resource.component)
		err = i.applier.Apply(parsedResource)
		if err != nil {
			i.cfg.Log.Warnf("Error occurred when processing file %s of component %s : %s", resource.fileName, resource.component, err.Error())
			o.NotInstalled = append(o.NotInstalled, file)
			continue
		}

		o.Installed = append(o.Installed, file)
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
