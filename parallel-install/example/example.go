package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller"
)

var log *logger.Logger

//main provides an example of how to integrate the parallel-install library with your code.
func main() {
	kubeconfigPath := flag.String("kubeconfig", "", "Path to the Kubeconfig file")
	kubeconfigContent := flag.String("kubeconfigcontent", "", "Raw content of the Kubeconfig file")
	profile := flag.String("profile", "", "Deployment profile")
	version := flag.String("version", "latest", "Kyma version")
	verbose := flag.Bool("verbose", false, "Verbose mode")

	flag.Parse()

	log = logger.NewLogger(*verbose)

	if (kubeconfigPath == nil || *kubeconfigPath == "") &&
		(kubeconfigContent == nil || *kubeconfigContent == "") {
		log.Fatal("either kubeconfig or kubeconfigcontent property has to be set")
	}

	if version == nil || *version == "" {
		log.Fatal("version is required")
	}

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatal("Please set GOPATH")
	}

	builder := &deployment.OverridesBuilder{}
	if err := builder.AddFile("./overrides.yaml"); err != nil {
		log.Error("Failed to add overrides file. Exiting...")
		os.Exit(1)
	}
	newKymaOverrides := make(map[string]interface{})
	newKymaOverrides["isBEBEnabled"] = true
	if err := builder.AddOverrides("global", newKymaOverrides); err != nil {
		log.Error("Failed to add overrides isBEBEnabled. Exiting...")
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
		KubeconfigSource: config.KubeconfigSource{
			Path:    *kubeconfigPath,
			Content: *kubeconfigContent,
		},
		Version: *version,
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
		KubeconfigSource:         installationCfg.KubeconfigSource,
	}

	resourceParser := &preinstaller.GenericResourceParser{}
	resourceManager, err := preinstaller.NewDefaultResourceManager(installationCfg.KubeconfigSource, preInstallerCfg.Log, commonRetryOpts)
	if err != nil {
		log.Fatalf("Failed to create Kyma default resource manager: %v", err)
	}

	resourceApplier := preinstaller.NewGenericResourceApplier(installationCfg.Log, resourceManager)
	preInstaller, err := preinstaller.NewPreInstaller(resourceApplier, resourceParser, preInstallerCfg, commonRetryOpts)
	if err != nil {
		log.Fatalf("Failed to create Kyma pre-installer: %v", err)
	}

	result, err := preInstaller.InstallCRDs()
	if err != nil || len(result.NotInstalled) > 0 {
		log.Fatalf("Failed to install CRDs: %s", err)
	}

	result, err = preInstaller.CreateNamespaces()
	if err != nil || len(result.NotInstalled) > 0 {
		log.Fatalf("Failed to create namespaces: %s", err)
	}

	//Deploy Kyma
	deployer, err := deployment.NewDeployment(installationCfg, builder, callbackUpdate)
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}

	err = deployer.StartKymaDeployment()
	if err != nil {
		log.Errorf("Failed to deploy Kyma: %v", err)
	} else {
		log.Info("Kyma deployed!")
	}

	metadataProvider, err := helm.NewKymaMetadataProvider(installationCfg.KubeconfigSource)
	if err != nil {
		log.Fatalf("Failed to create Kyma metadata provider: %v", err)
	}

	versionSet, err := metadataProvider.Versions()
	if err == nil {
		log.Infof("Found %d Kyma version: %s", versionSet.Count(), strings.Join(versionSet.Names(), ", "))
	} else {
		log.Errorf("Failed to deploy Kyma: %v", err)
	}

	//Delete Kyma
	deleter, err := deployment.NewDeletion(installationCfg, builder, callbackUpdate)
	if err != nil {
		log.Fatalf("Failed to create deleter: %v", err)
	}
	err = deleter.StartKymaUninstallation()
	if err != nil {
		log.Fatalf("Failed to uninstall Kyma: %v", err)
	}
	log.Info("Kyma uninstalled!")
}

func callbackUpdate(update deployment.ProcessUpdate) {

	showCompStatus := func(comp components.KymaComponent) {
		if comp.Name != "" {
			log.Infof("Status of component '%s': %s", comp.Name, comp.Status)
		}
	}

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
