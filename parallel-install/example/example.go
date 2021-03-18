package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"k8s.io/client-go/dynamic"

	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
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

	compList, err := config.NewComponentList("./components.yaml")
	if err != nil {
		log.Fatalf("Cannot read component list: %s", err)
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
		ComponentList:                 compList,
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
	preInstallerCfg := preinstaller.Config{
		InstallationResourcePath: installationCfg.InstallationResourcePath,
		Log:                      installationCfg.Log,
	}

	resourceManager := preinstaller.NewDefaultResourceManager(dynamicClient, preInstallerCfg.Log, commonRetryOpts)
	resourceApplier := preinstaller.NewGenericResourceApplier(installationCfg.Log, resourceManager)
	preInstaller := preinstaller.NewPreInstaller(resourceApplier, preInstallerCfg, dynamicClient, commonRetryOpts)

	result, err := preInstaller.InstallCRDs()
	if err != nil || len(result.NotInstalled) > 0 {
		log.Fatalf("Failed to install CRDs: %s", err)
	}

	result, err = preInstaller.CreateNamespaces()
	if err != nil || len(result.NotInstalled) > 0 {
		log.Fatalf("Failed to create namespaces: %s", err)
	}

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

	metadataProvider := helm.NewKymaMetadataProvider(kubeClient)
	versionSet, err := metadataProvider.Versions()
	if err == nil {
		log.Infof("Found %d Kyma version: %s", versionSet.Count(), strings.Join(versionSet.Names(), ", "))
	} else {
		log.Errorf("Failed to deploy Kyma: %v", err)
	}

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
