package main

import (
	"flag"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//main provides an example of how to integrate the parallel-install library with your code.
func main() {
	kubeconfigPath := flag.String("kubeconfig", "", "Path to the Kubeconfig file")
	flag.Parse()

	if kubeconfigPath == nil || *kubeconfigPath == "" {
		log.Fatalf("kubeconfig is required")
	}

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatalf("Please set GOPATH")
	}

	resourcesPath := filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma", "resources")

	restConfig, err := getClientConfig(*kubeconfigPath)
	if err != nil {
		log.Fatalf("Unable to build kubernetes configuration. Error: %v", err)
	}

	prerequisitesContent := [][]string{
		{"cluster-essentials", "kyma-system"},
		{"istio", "istio-system"},
		{"xip-patch", "kyma-installer"},
	}

	componentsContent, err := ioutil.ReadFile("../pkg/test/data/installationCR.yaml")
	if err != nil {
		log.Fatalf("Failed to read installation CR file: %v", err)
	}

	overridesContent, err := ioutil.ReadFile("../pkg/test/data/overrides.yaml")
	if err != nil {
		log.Fatalf("Failed to read overrides file: %v", err)
	}

	installationCfg := config.Config{
		WorkersCount:                  4,
		CancelTimeout:                 20 * time.Minute,
		QuitTimeout:                   25 * time.Minute,
		HelmTimeoutSeconds:            60 * 8,
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           log.Printf,
		HelmMaxRevisionHistory:        10,
	}

	installer, err := deployment.NewDeployment(prerequisitesContent,
		string(componentsContent),
		[]string{string(overridesContent)},
		resourcesPath,
		installationCfg)
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}

	config.SetupLogger(log.Printf)

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Printf("Failed to create kube client. Exiting...")
		os.Exit(1)
	}

	err = installer.StartKymaDeployment(kubeClient)
	if err != nil {
		log.Printf("Failed to deploy Kyma: %v", err)
	} else {
		log.Println("Kyma deployed!")
	}

	err = installer.StartKymaUninstallation(kubeClient)
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
