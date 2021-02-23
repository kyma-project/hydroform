package preinstaller

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"io/ioutil"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// PreInstaller performs CRDs installation.
type PreInstaller struct {
	cfg            config.Config
	kubeClient     kubernetes.Interface
	dynamicClient  dynamic.Interface
	retryOptions   []retry.Option
}

func NewPreInstaller(cfg config.Config, dynamicClient dynamic.Interface, retryOptions []retry.Option) *PreInstaller {
	return &PreInstaller{
		cfg:            cfg,
		dynamicClient:  dynamicClient,
		retryOptions:   retryOptions,
	}
}

func (i *PreInstaller) InstallCRDs() error {
	resource := newCrdPreInstallerResource(i.cfg.Log, i.dynamicClient, i.retryOptions)
	err := i.apply(*resource)
	if err != nil {
		return err
	}

	return nil
}

func (i *PreInstaller) CreateNamespaces() error {
	resource := newNamespacePreInstallerResource(i.cfg.Log, i.dynamicClient, i.retryOptions)
	err := i.apply(*resource)
	if err != nil {
		return err
	}

	return nil
}

func (i *PreInstaller) apply(resourceType resourceType) error {
	installationResourcePath := i.cfg.InstallationResourcePath
	path := fmt.Sprintf("%s/%s", installationResourcePath, resourceType.name)

	components, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	if len(components) == 0 {
		i.cfg.Log.Warn("There were no components detected for installation. Skipping.")
		return nil
	}

	for _, component := range components {
		componentName := component.Name()
		i.cfg.Log.Infof("Processing component: %s", componentName)
		pathToComponent := fmt.Sprintf("%s/%s", path, componentName)
		resources, err := ioutil.ReadDir(pathToComponent)
		if err != nil {
			return err
		}

		if len(resources) == 0 {
			i.cfg.Log.Warnf("There were no resources detected for component: ", componentName)
			return nil
		}

		for _, resource := range resources {
			resourceName := resource.Name()
			i.cfg.Log.Infof("Processing file: %s", resourceName)
			pathToResource := fmt.Sprintf("%s/%s", pathToComponent, resourceName)
			resourceData, err := ioutil.ReadFile(pathToResource)
			if err != nil {
				return err
			}

			err = resourceType.applier.Apply(string(resourceData))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
