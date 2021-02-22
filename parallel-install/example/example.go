package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/avast/retry-go"
	"k8s.io/client-go/dynamic"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller"
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

	log := logger.NewLogger(*verbose)

	if kubeconfigPath == nil || *kubeconfigPath == "" {
		log.Fatal("kubeconfig is required")
	}

	if version == nil || *version == "" {
		log.Fatal("version is required")
	}

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatal("Please set GOPATH")
	}

	restConfig, err := getClientConfig(*kubeconfigPath)
	if err != nil {
		log.Fatalf("Unable to build kubernetes configuration. Error: %v", err)
	}

	builder := &deployment.OverridesBuilder{}
	if err := builder.AddFile("./overrides.yaml"); err != nil {
		log.Error("Failed to Add overrides file. Exiting...")
		os.Exit(1)
	}

	installationCfg := &config.Config{
		WorkersCount:                  4,
		CancelTimeout:                 20 * time.Minute,
		QuitTimeout:                   25 * time.Minute,
		HelmTimeoutSeconds:            60 * 8,
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           log,
		HelmMaxRevisionHistory:        10,
		Profile:                       *profile,
		ComponentsListFile:            "./components.yaml",
		ResourcePath:                  fmt.Sprintf("%s/src/github.com/kyma-project/kyma/resources", goPath),
		InstallationResourcePath:      fmt.Sprintf("%s/src/github.com/kyma-project/kyma/installation/resources", goPath),
		Version:                       *version,
	}

	// used to receive progress updates of the install/uninstall process
	var progressCh chan deployment.ProcessUpdate
	if !(*verbose) {
		progressCh = make(chan deployment.ProcessUpdate)
		ctx := renderProgress(progressCh, log)
		defer func() {
			close(progressCh)
			<-ctx.Done()
		}()
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Error("Failed to create kube client. Exiting...")
		os.Exit(1)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		log.Error("Failed to create kube dynamic client. Exiting...")
		os.Exit(1)
	}

	commonRetryOpts := []retry.Option{
		retry.Delay(time.Duration(installationCfg.BackoffInitialIntervalSeconds) * time.Second),
		retry.Attempts(uint(installationCfg.BackoffMaxElapsedTimeSeconds / installationCfg.BackoffInitialIntervalSeconds)),
		retry.DelayType(retry.FixedDelay),
	}

	//Prepare cluster before Kyma installation
	preInstaller := preinstaller.NewPreInstaller(installationCfg, kubeClient, dynamicClient, commonRetryOpts, progressCh)

	err = preInstaller.InstallCRDs()
	if err != nil {
		log.Errorf("Failed to install CRDs: %s", err)
	}

	err = preInstaller.CreateNamespaces()
	if err != nil {
		log.Errorf("Failed to create namespaces: %s", err)
	}

	// TODO: remove
	return

	//Deploy Kyma
	deployer, err := deployment.NewDeployment(installationCfg, builder, kubeClient, progressCh)
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}

	err = deployer.StartKymaDeployment()
	if err != nil {
		log.Errorf("Failed to deploy Kyma: %v", err)
	} else {
		log.Info("Kyma deployed!")
	}

	kymaMeta, err := deployer.ReadKymaMetadata()
	if err != nil {
		log.Errorf("Failed to read Kyma metadata: %v", err)
	}

	log.Infof("Kyma version: %s", kymaMeta.Version)
	log.Infof("Kyma status: %s", kymaMeta.Status)

	//Delete Kyma
	deleter, err := deployment.NewDeletion(installationCfg, builder, kubeClient, progressCh)
	if err != nil {
		log.Fatalf("Failed to create deleter: %v", err)
	}
	err = deleter.StartKymaUninstallation()
	if err != nil {
		log.Fatalf("Failed to uninstall Kyma: %v", err)
	}
	log.Info("Kyma uninstalled!")
}

func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func renderProgress(progressCh chan deployment.ProcessUpdate, log logger.Interface) context.Context {
	context, cancel := context.WithCancel(context.Background())

	showCompStatus := func(comp components.KymaComponent) {
		if comp.Name != "" {
			log.Infof("Status of component '%s': %s", comp.Name, comp.Status)
		}
	}
	go func() {
		defer cancel()

		for update := range progressCh {
			switch update.Event {
			case deployment.ProcessStart:
				log.Infof("Starting installation phase '%s'", update.Phase)
			case deployment.ProcessRunning:
				showCompStatus(update.Component)
			case deployment.ProcessFinished:
				log.Infof("Finished installation phase '%s' successfully", update.Phase)
			default:
				//any failure case
				log.Infof("Process failed in phase '%s' with error state '%s':", update.Phase, update.Event)
				showCompStatus(update.Component)
			}
		}
	}()

	return context
}
