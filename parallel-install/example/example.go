package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//main provides an example of how to integrate the parallel-install library with your code.
func main() {
	kubeconfigPath := flag.String("kubeconfig", "", "Path to the Kubeconfig file")
	profile := flag.String("profile", "", "Deployment profile")
	version := flag.String("version", "latest", "Kyma version")
	verbose := flag.Bool("verbose", false, "Verbose mode")

	flag.Parse()

	if kubeconfigPath == nil || *kubeconfigPath == "" {
		log.Fatalf("kubeconfig is required")
	}

	if version == nil || *version == "" {
		log.Fatalf("version is required")
	}

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatalf("Please set GOPATH")
	}

	restConfig, err := getClientConfig(*kubeconfigPath)
	if err != nil {
		log.Fatalf("Unable to build kubernetes configuration. Error: %v", err)
	}

	overrides := deployment.Overrides{}
	overrides.AddFile("./overrides.yaml")

	installationCfg := config.Config{
		WorkersCount:                  4,
		CancelTimeout:                 20 * time.Minute,
		QuitTimeout:                   25 * time.Minute,
		HelmTimeoutSeconds:            60 * 8,
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           getLogFunc(*verbose),
		HelmMaxRevisionHistory:        10,
		Profile:                       *profile,
		ComponentsListFile:            "./components.yaml",
		ResourcePath:                  fmt.Sprintf("%s/src/github.com/kyma-project/kyma/resources", goPath),
		CrdPath:                       fmt.Sprintf("%s/src/github.com/kyma-project/kyma/resources/cluster-essentials/files", goPath),
		Version:                       *version,
	}

	// used to receive progress updates of the install/uninstall process
	var progressCh chan deployment.ProcessUpdate
	if !(*verbose) {
		progressCh = make(chan deployment.ProcessUpdate)
		ctx := renderProgress(progressCh)
		defer func() {
			close(progressCh)
			<-ctx.Done()
		}()
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Printf("Failed to create kube client. Exiting...")
		os.Exit(1)
	}

	//Deploy Kyma
	deployer, err := deployment.NewDeployment(installationCfg, overrides, kubeClient, progressCh)
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}

	err = deployer.StartKymaDeployment()
	if err != nil {
		log.Printf("Failed to deploy Kyma: %v", err)
	} else {
		log.Println("Kyma deployed!")
	}

	kymaMeta, err := deployer.ReadKymaMetadata()
	if err != nil {
		log.Printf("Failed to read Kyma metadata: %v", err)
	}

	log.Printf("Kyma version: %s", kymaMeta.Version)
	log.Printf("Kyma status: %s", kymaMeta.Status)

	//Delete Kyma
	deleter, err := deployment.NewDeletion(installationCfg, overrides, kubeClient, progressCh)
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}
	err = deleter.StartKymaUninstallation()
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

func getLogFunc(verbose bool) func(string, ...interface{}) {
	if verbose {
		return log.Printf
	}
	return func(msg string, args ...interface{}) {
		// do nothing
	}
}

func renderProgress(progressCh chan deployment.ProcessUpdate) context.Context {
	context, cancel := context.WithCancel(context.Background())

	showCompStatus := func(comp components.KymaComponent) {
		if comp.Name != "" {
			log.Printf("Status of component '%s': %s", comp.Name, comp.Status)
		}
	}
	go func() {
		defer cancel()

		for update := range progressCh {
			switch update.Event {
			case deployment.ProcessStart:
				log.Printf("Starting installation phase '%s'", update.Phase)
			case deployment.ProcessRunning:
				showCompStatus(update.Component)
			case deployment.ProcessFinished:
				log.Printf("Finished installation phase '%s' successfully", update.Phase)
			default:
				//any failure case
				log.Printf("Process failed in phase '%s' with error state '%s':", update.Phase, update.Event)
				showCompStatus(update.Component)
			}
		}
	}()

	return context
}
