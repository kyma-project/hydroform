package installation

import (
	"fmt"
	"log"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/components"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/engine"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Installation struct {
	// Map component > namespace
	Prerequisites map[string]string
	// Content of the Installation CR YAML file
	ComponentsYaml              string
	// Content of the Helm overrides YAML file
	OverridesYaml string
	ResourcesPath string
}

type Installer interface {
	StartKymaInstallation(kubeconfig *rest.Config) error
	StartKymaUninstallation(kubeconfig *rest.Config) error
}

func NewInstallation(prerequisites map[string]string, componentsYaml string, overridesYaml string, resourcesPath string) (*Installation, error) {
	if resourcesPath == "" {
		return nil, fmt.Errorf("Unable to create Installation. Resource path is required.")
	}
	if componentsYaml == "" {
		return nil, fmt.Errorf("Unable to create Installation. Components YAML file content is required.")
	}

	return &Installation{
		Prerequisites:  prerequisites,
		ComponentsYaml: componentsYaml,
		OverridesYaml:  overridesYaml,
		ResourcesPath:  resourcesPath,
	}, nil
}

func (i *Installation) StartKymaInstallation(kubeconfig *rest.Config) error {
	kubeClient, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("Unable to create internal client. Error: %v", err)
	}

	overridesProvider, err := overrides.New(kubeClient, i.OverridesYaml)
	if err != nil {
		log.Fatalf("Unable to create overrides provider. Error: %v", err)
	}

	prerequisitesProvider := components.NewPrerequisitesProvider(overridesProvider, i.ResourcesPath, i.Prerequisites)
	componentsProvider := components.NewComponentsProvider(overridesProvider, i.ResourcesPath, i.ComponentsYaml)

	eng := engine.NewEngine(overridesProvider, prerequisitesProvider, componentsProvider, i.ResourcesPath)

	fmt.Println("Kyma installation")
	err = eng.Install()
	if err != nil {
		return fmt.Errorf("Kyma installation failed. Error: %v", err)
	}

	return nil
}

func (i *Installation) StartKymaUninstallation(kubeconfig *rest.Config) error {
	kubeClient, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Fatalf("Unable to create internal client. Error: %v", err)
	}

	overridesProvider, err := overrides.New(kubeClient, i.OverridesYaml)
	if err != nil {
		log.Fatalf("Unable to create overrides provider. Error: %v", err)
	}

	prerequisitesProvider := components.NewPrerequisitesProvider(overridesProvider, i.ResourcesPath, i.Prerequisites)
	componentsProvider := components.NewComponentsProvider(overridesProvider, i.ResourcesPath, i.ComponentsYaml)

	eng := engine.NewEngine(overridesProvider, prerequisitesProvider, componentsProvider, i.ResourcesPath)

	fmt.Println("Kyma uninstallation")

	err = eng.Uninstall()
	if err != nil {
		log.Fatalf("Kyma uninstallation failed. Error: %v", err)
	}

	return nil
}
