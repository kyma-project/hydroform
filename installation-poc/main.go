package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/installation"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatalf("Please set GOPATH")
	}

	resourcesPath := filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma", "resources")
	kubeconfigPath := "/Users/I517624/.kube/config"

	config, err := getClientConfig(kubeconfigPath)
	if err != nil {
		log.Fatalf("Unable to build kubernetes configuration. Error: %v", err)
	}

	//TODO: get prerequisites yaml
	prerequisitesContent := ""

	componentsContent, err := ioutil.ReadFile("pkg/test/data/installationCR.yaml")
	if err != nil {
		log.Fatalf("Failed to read installation CR file: %v", err)
	}

	overridesContent, err := ioutil.ReadFile("pkg/test/data/overrides.yaml")
	if err != nil {
		log.Fatalf("Failed to read overrides file: %v", err)
	}

	installer, err := installation.NewInstallation(prerequisitesContent, string(componentsContent), string(overridesContent), resourcesPath)
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
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
