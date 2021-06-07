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
// CRDS are labeled with: LABEL_KEY_ORIGIN=LABEL_VALUE_KYMA, which come from constants,
// in order to distinguish them among other resources not managed by Kyma.
// As a result, on basis of the label they are marked for deletion during Kyma uninstallation.

package deployment

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"k8s.io/client-go/dynamic"
)

// inputConfig defines configuration values for the preInstaller.
type inputConfig struct {
	InstallationResourcePath string                  //Path to the installation resources.
	Log                      logger.Interface        //Logger to be used.
	KubeconfigSource         config.KubeconfigSource //KubeconfigSource to be used.
	RetryOptions             []retry.Option          //RetryOptions for networking operations.
}

// preInstaller prepares k8s cluster for Kyma installation.
type preInstaller struct {
	applier       ResourceApplier
	parser        ResourceParser
	cfg           inputConfig
	dynamicClient dynamic.Interface
}

// file consists of a path to the file that was a part of preInstaller installation
// and a component fileName that it belongs to.
type file struct {
	component string
	path      string
}

// output contains lists of Installed and not Installed files during preInstaller installation.
type output struct {
	// Installed files during preInstaller installation.
	Installed []file
	// NotInstalled files during preInstaller installation.
	NotInstalled []file
}

type resourceInfoInput struct {
	dirSuffix                string
	resourceType             string
	installationResourcePath string
	label                    string
}

type resourceInfoResult struct {
	component    string
	fileName     string
	path         string
	resourceType string
	label        string
}

// newPreInstaller creates a new instance of preInstaller.
func newPreInstaller(cfg inputConfig) (*preInstaller, error) {
	restConfig, err := config.RestConfig(cfg.KubeconfigSource)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	manager, err := NewDefaultResourceManager(cfg.KubeconfigSource, cfg.Log, cfg.RetryOptions)
	if err != nil {
		cfg.Log.Fatalf("Failed to create Kyma default resource manager: %v", err)
	}

	applier := NewGenericResourceApplier(cfg.Log, manager)
	parser := &GenericResourceParser{}

	return &preInstaller{
		applier:       applier,
		parser:        parser,
		cfg:           cfg,
		dynamicClient: dynamicClient,
	}, nil
}

// InstallCRDs on a k8s cluster.
// Returns output containing results of installation.
func (i *preInstaller) InstallCRDs() error {
	input := resourceInfoInput{
		resourceType:             "CustomResourceDefinition",
		dirSuffix:                "crds",
		installationResourcePath: i.cfg.InstallationResourcePath,
		label:                    config.LABEL_KEY_ORIGIN,
	}

	i.cfg.Log.Info("Kyma CRDs installation")
	output, err := i.install(input)
	if err != nil || len(output.NotInstalled) > 0 {
		return errors.Wrap(err, "Failed to install CRDs")
	}

	return nil
}

// CreateNamespaces in a k8s cluster.
// Returns output containing results of installation.
func (i *preInstaller) CreateNamespaces() error {
	input := resourceInfoInput{
		resourceType:             "Namespace",
		dirSuffix:                "namespaces",
		installationResourcePath: i.cfg.InstallationResourcePath,
		label:                    "",
	}

	i.cfg.Log.Info("Kyma Namespaces creation")
	output, err := i.install(input)
	if err != nil || len(output.NotInstalled) > 0 {
		return errors.Wrap(err, "Failed to create namespaces")
	}

	return nil
}

func (i *preInstaller) install(input resourceInfoInput) (o output, err error) {
	resources, err := i.findResourcesIn(input)
	if err != nil {
		return output{}, err
	}

	return i.apply(resources)
}

func (i *preInstaller) findResourcesIn(input resourceInfoInput) (results []resourceInfoResult, err error) {
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
			i.cfg.Log.Warnf("There were no resources detected for component: %s", componentName)
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
				label:        input.label,
			}

			results = append(results, resourceInfoResult)
		}
	}

	return results, nil
}

func (i *preInstaller) apply(resources []resourceInfoResult) (o output, err error) {
	for _, resource := range resources {
		file := file{
			component: resource.component,
			path:      resource.path,
		}

		parsedResource, err := i.parser.ParseFile(file.path)
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

		addLabel(parsedResource, resource.label, config.LABEL_VALUE_KYMA)

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

func addLabel(obj *unstructured.Unstructured, label string, value string) {
	if len(label) < 1 {
		return
	}

	labels := obj.GetLabels()
	if labels == nil {
		newLabels := map[string]string{
			label: value,
		}

		obj.SetLabels(newLabels)
	} else {
		labels[label] = value
		obj.SetLabels(labels)
	}
}
