package main

import (
	"flag"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"log"

	"os"
	"path/filepath"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/installation"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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
		CancelTimeoutSeconds:          60 * 20,
		QuitTimeoutSeconds:            60 * 25,
		HelmTimeoutSeconds:            60 * 8,
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           log.Printf,
	}

	installer, err := installation.NewInstallation(prerequisitesContent,
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
		//return fmt.Errorf("Unable to create internal client. Error: %v", err)
	}

	overridesProvider, err := overrides.New(kubeClient, installer.OverridesYamls, installer.Cfg.Log)
	if err != nil {
		//return fmt.Errorf("Unable to create overrides provider. Error: %v", err)
	}

	prerequisitesProvider := components.NewPrerequisitesProvider(overridesProvider, installer.ResourcesPath, installer.Prerequisites, installer.Cfg)
	componentsProvider := components.NewComponentsProvider(overridesProvider, installer.ResourcesPath, installer.ComponentsYaml, installer.Cfg)

	engineCfg := engine.Config{WorkersCount: installer.Cfg.WorkersCount}
	eng := engine.NewEngine(overridesProvider, componentsProvider, installer.ResourcesPath, engineCfg)

	err = installer.StartKymaInstallation(*prerequisitesProvider, overridesProvider, eng)
	if err != nil {
		log.Printf("Failed to install Kyma: %v", err)
	} else {
		log.Println("Kyma installed!")
	}

	err = installer.StartKymaUninstallation(*prerequisitesProvider, overridesProvider, eng)
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
