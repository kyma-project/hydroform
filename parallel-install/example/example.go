package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
)

const (
	actionDeployAndDelete = "deploy+delete"
	actionTemplate        = "template"
)

var log *logger.Logger

//main provides an example of how to integrate the parallel-install library with your code.
func main() {
	kubeconfigPath := flag.String("kubeconfig", "", "Path to the Kubeconfig file")
	kubeconfigContent := flag.String("kubeconfigcontent", "", "Raw content of the Kubeconfig file")
	profile := flag.String("profile", "", "Deployment profile")
	version := flag.String("version", "latest", "Kyma version")
	verbose := flag.Bool("verbose", false, "Verbose mode")
	action := flag.String("action", actionDeployAndDelete,
		fmt.Sprintf("Which action to apply. Supported are: %s (default: %s)",
			strings.Join([]string{actionDeployAndDelete, actionTemplate}, ", "), actionDeployAndDelete))

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

	builder := &overrides.Builder{}
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

	cfg := &config.Config{
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

	switch *action {
	case actionDeployAndDelete:
		deployAndDelete(cfg, builder)
	case actionTemplate:
		template(cfg, builder)
	default:
		log.Errorf("Action '%s' is not supported", *action)
	}
}

func deployAndDelete(cfg *config.Config, builder *overrides.Builder) {
	commonRetryOpts := []retry.Option{
		retry.Delay(time.Duration(cfg.BackoffInitialIntervalSeconds) * time.Second),
		retry.Attempts(uint(cfg.BackoffMaxElapsedTimeSeconds / cfg.BackoffInitialIntervalSeconds)),
		retry.DelayType(retry.FixedDelay),
	}

	//Deploy Kyma
	deployer, err := deployment.NewDeployment(cfg, builder, callbackUpdate)
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}

	err = deployer.StartKymaDeployment()
	if err != nil {
		log.Errorf("Failed to deploy Kyma: %v", err)
	} else {
		log.Info("Kyma deployed!")
	}

	metadataProvider, err := helm.NewKymaMetadataProvider(cfg.KubeconfigSource)
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
	deleter, err := deployment.NewDeletion(cfg, builder, callbackUpdate, commonRetryOpts)
	if err != nil {
		log.Fatalf("Failed to create deleter: %v", err)
	}
	err = deleter.StartKymaUninstallation()
	if err != nil {
		log.Fatalf("Failed to uninstall Kyma: %v", err)
	}

	log.Info("Kyma uninstalled!")
}

func template(cfg *config.Config, builder *overrides.Builder) {
	//Deploy Kyma
	templating, err := deployment.NewTemplating(cfg, builder)
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}
	manifests, err := templating.Render()
	if err != nil {
		log.Fatalf("Failed to render Helm charts: %v", err)
	}
	for _, manifest := range manifests {
		var filename string
		if manifest.Type == components.CRD {
			filename = path.Join("template", manifest.Name)
		} else {
			filename = path.Join("template", fmt.Sprintf("%s.yaml", manifest.Name))
		}

		if err := ioutil.WriteFile(filename, []byte(manifest.Manifest), 0600); err != nil {
			log.Errorf("Failed to write manifest '%s'", manifest.Name)
		}
	}
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
