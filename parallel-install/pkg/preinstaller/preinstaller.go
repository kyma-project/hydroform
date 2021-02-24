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
	resource := resourceType{
		name: "crds",
	}

	output, err := i.apply(resource)
	if err != nil {
		return Output{}, err
	}

	return output, nil
}

// CreateNamespaces in a k8s cluster.
// Returns Output containing results of installation.
func (i *PreInstaller) CreateNamespaces() (Output, error) {
	resource := resourceType{
		name: "namespaces",
	}

	output, err := i.apply(resource)
	if err != nil {
		return Output{}, err
	}

	return output, nil
}

func (i *PreInstaller) apply(resourceType resourceType) (o Output, err error) {
	installationResourcePath := i.cfg.InstallationResourcePath
	path := fmt.Sprintf("%s/%s", installationResourcePath, resourceType.name)

	rawComponentsDir, err := ioutil.ReadDir(path)
	if err != nil {
		return o, err
	}

	components := findOnlyDirectoriesAmong(rawComponentsDir)
	if components == nil || len(components) == 0 {
		i.cfg.Log.Warn("There were no components detected for installation. Skipping.")
		return o, nil
	}

	for _, component := range components {
		componentName := component.Name()
		i.cfg.Log.Infof("Processing component: %s", componentName)
		pathToComponent := fmt.Sprintf("%s/%s", path, componentName)
		resources, err := ioutil.ReadDir(pathToComponent)
		if err != nil {
			return o, err // TODO: fail-fast or continue?
		}

		if len(resources) == 0 {
			i.cfg.Log.Warnf("There were no resources detected for component: ", componentName)
			break
		}

		for _, resource := range resources {
			resourceName := resource.Name()
			i.cfg.Log.Infof("Processing file: %s", resourceName)
			pathToResource := fmt.Sprintf("%s/%s", pathToComponent, resourceName)
			file := File{
				component: componentName,
				path:      pathToResource,
			}

			resourceData, err := ioutil.ReadFile(pathToResource)
			if err != nil {
				o.notInstalled = append(o.notInstalled, file)
				return o, err // TODO: fail-fast or continue?
			}

			err = i.applier.Apply(string(resourceData))
			if err != nil {
				o.notInstalled = append(o.notInstalled, file)
				i.cfg.Log.Warnf("Error occured when processing file %s : %s", resourceName, err)
			} else {
				o.installed = append(o.installed, file)
			}
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
