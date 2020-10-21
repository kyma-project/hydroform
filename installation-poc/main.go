package main

import (
	"fmt"
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

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create internal client. Error: %v", err)
	}

	overridesProvider := overrides.New(kubeClient)

	componentsProvider := components.NewComponents(overridesProvider, resourcesPath)

	eng := engine.NewEngine(componentsProvider, resourcesPath)

	fmt.Println("Kyma installation")
	err = eng.Install()
	if err != nil {
		log.Fatalf("Kyma installation failed. Error: %v", err)
	}

	fmt.Println("Kyma installed")
	fmt.Println("Kyma uninstallation")

	err = eng.Uninstall()
	if err != nil {
		log.Fatalf("Kyma uninstallation failed. Error: %v", err)
	}

	fmt.Println("Kyma uninstalled")
}

func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}


