package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/components"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/engine"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var resourcesPath string

func main() {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatalf("Please set GOPATH")
	}

	resourcesPath = filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma", "resources")
	kubeconfigPath := "/Users/i304607/Downloads/mst.yml"

	config, err := getClientConfig(kubeconfigPath)
	if err != nil {
		log.Fatalf("Unable to build kubernetes configuration. Error: %v", err)
	}

	componentsContent, err := ioutil.ReadFile("pkg/test/data/installationCR.yaml")
	if err != nil {
		log.Fatalf("Failed to read installation CR file: %v", err)
	}

	overridesContent, err := ioutil.ReadFile("pkg/test/data/overrides.yaml")
	if err != nil {
		log.Fatalf("Failed to read overrides file: %v", err)
	}

	installer := Installation{
		ResourcesPath:  resourcesPath,
		ComponentsYaml: string(componentsContent),
		OverridesYaml:  string(overridesContent),
	}

	err = installer.StartKymaInstallation(config)
	if err != nil {
		log.Fatalf("Failed to install Kyma: %v", err)
	}
	log.Println("Kyma installed!")

	err = installer.StartKymaUninstallation(config)
	if err != nil {
		log.Fatalf("Failed to uninstall Kyma: %v", err)
	}
	log.Println("Kyma uninstalled!")
}

func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

type Installation struct {
	// Content of the Installation CR YAML file
	ComponentsYaml string
	// Content of the Helm overrides YAML file
	OverridesYaml string
	ResourcesPath string
}

type Installer interface {
	StartKymaInstallation(kubeconfig *rest.Config) error
	StartKymaUninstallation(kubeconfig *rest.Config) error
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

	componentsProvider := components.NewComponents(overridesProvider, i.ResourcesPath, i.ComponentsYaml)

	eng := engine.NewEngine(overridesProvider, componentsProvider, i.ResourcesPath)

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

	componentsProvider := components.NewComponents(overridesProvider, i.ResourcesPath, i.ComponentsYaml)

	eng := engine.NewEngine(overridesProvider, componentsProvider, i.ResourcesPath)

	fmt.Println("Kyma uninstallation")

	err = eng.Uninstall()
	if err != nil {
		log.Fatalf("Kyma uninstallation failed. Error: %v", err)
	}

	return nil
}
