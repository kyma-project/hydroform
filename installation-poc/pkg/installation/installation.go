package installation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/components"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/config"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/engine"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Installation struct {
	// Map component > namespace
	Prerequisites [][]string
	// Content of the Installation CR YAML file
	ComponentsYaml string
	// Content of the Helm overrides YAML file
	OverridesYaml string
	ResourcesPath string
	Cfg           config.Config
}

type Installer interface {
	//This method will block until installation is finished or an error or timeout occurs.
	//If the installation is not finished in configured config.Config.QuitTimeoutSeconds, the method returns with an error. Some worker goroutines may still be active.
	StartKymaInstallation(kubeconfig *rest.Config) error
	//This method will block until uninstallation is finished or an error or timeout occurs.
	//If the uninstallation is not finished in configured config.Config.QuitTimeoutSeconds, the method returns with an error. Some worker goroutines may still be active.
	StartKymaUninstallation(kubeconfig *rest.Config) error
}

func NewInstallation(prerequisites [][]string, componentsYaml string, overridesYaml string, resourcesPath string, cfg config.Config) (*Installation, error) {
	if resourcesPath == "" {
		return nil, fmt.Errorf("Unable to create Installation. Resource path is required.")
	}
	if componentsYaml == "" {
		return nil, fmt.Errorf("Unable to create Installation. Components YAML file content is required.")
	}

	return &Installation{
		Prerequisites:  prerequisites,
		ComponentsYaml: componentsYaml,
		OverridesYaml:  overridesYaml,
		ResourcesPath:  resourcesPath,
		Cfg:            cfg,
	}, nil
}

func (i *Installation) StartKymaInstallation(kubeconfig *rest.Config) error {
	kubeClient, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("Unable to create internal client. Error: %v", err)
	}

	overridesProvider, err := overrides.New(kubeClient, i.OverridesYaml)
	if err != nil {
		return fmt.Errorf("Unable to create overrides provider. Error: %v", err)
	}

	prerequisitesProvider := components.NewPrerequisitesProvider(overridesProvider, i.ResourcesPath, i.Prerequisites, i.Cfg)
	componentsProvider := components.NewComponentsProvider(overridesProvider, i.ResourcesPath, i.ComponentsYaml, i.Cfg)

	engineCfg := engine.Config{i.Cfg.WorkersCount}
	eng := engine.NewEngine(overridesProvider, prerequisitesProvider, componentsProvider, i.ResourcesPath, engineCfg)

	fmt.Println("Kyma installation")
	cancelCtx, cancel := context.WithCancel(context.Background())
	statusChan, err := eng.Install(cancelCtx)
	if err != nil {
		return fmt.Errorf("Kyma installation failed. Error: %v", err)
	}

	var statusMap = map[string]string{}
	var errCount int = 0
	var cancelTimeout time.Duration = time.Duration(i.Cfg.CancelTimeoutSeconds) * time.Second
	var quitTimeout time.Duration = time.Duration(i.Cfg.QuitTimeoutSeconds) * time.Second
	var timeoutOccured bool = false

	//Await completion
	for {
		select {
		case cmp, ok := <-statusChan:
			if ok {
				//Received a status update
				if cmp.Status == components.StatusError {
					errCount++
				}
				statusMap[cmp.Name] = cmp.Status
			} else {
				//statusChan is closed
				if errCount > 0 {
					logStatuses(statusMap)
					return fmt.Errorf("Kyma installation failed due to errors in %d component(s)", errCount)
				}
				if timeoutOccured {
					logStatuses(statusMap)
					return fmt.Errorf("Kyma installation failed due to the timeout")
				}
				return nil
			}
		case <-time.After(cancelTimeout):
			timeoutOccured = true
			log.Printf("Timeout occurred after %v minutes. Cancelling installation", cancelTimeout.Minutes())
			cancel()
		case <-time.After(quitTimeout):
			log.Print("Installation doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Kyma installation failed due to the timeout")
		}
	}
}

func (i *Installation) StartKymaUninstallation(kubeconfig *rest.Config) error {
	kubeClient, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Printf("Unable to create internal client. Error: %v", err)
		return err
	}

	overridesProvider, err := overrides.New(kubeClient, i.OverridesYaml)
	if err != nil {
		log.Printf("Unable to create overrides provider. Error: %v", err)
		return err
	}

	prerequisitesProvider := components.NewPrerequisitesProvider(overridesProvider, i.ResourcesPath, i.Prerequisites, i.Cfg)
	componentsProvider := components.NewComponentsProvider(overridesProvider, i.ResourcesPath, i.ComponentsYaml, i.Cfg)

	engineCfg := engine.Config{i.Cfg.WorkersCount}
	eng := engine.NewEngine(overridesProvider, prerequisitesProvider, componentsProvider, i.ResourcesPath, engineCfg)

	log.Println("Kyma uninstallation started")

	cancelCtx, cancel := context.WithCancel(context.Background())
	statusChan, err := eng.Uninstall(cancelCtx)
	if err != nil {
		return err
	}

	var statusMap = map[string]string{}
	var errCount int = 0
	var cancelTimeout time.Duration = time.Duration(i.Cfg.CancelTimeoutSeconds) * time.Second
	var quitTimeout time.Duration = time.Duration(i.Cfg.QuitTimeoutSeconds) * time.Second
	var timeoutOccured bool = false

	//Await completion
	for {
		select {
		case cmp, ok := <-statusChan:
			if ok {
				if cmp.Status == components.StatusError {
					errCount++
				}
				statusMap[cmp.Name] = cmp.Status
			} else {
				if errCount > 0 {
					logStatuses(statusMap)
					return fmt.Errorf("Kyma uninstallation failed due to errors in %d component(s)", errCount)
				}
				if timeoutOccured {
					logStatuses(statusMap)
					return fmt.Errorf("Kyma uninstallation failed due to the timeout")
				}
				return nil
			}
		case <-time.After(cancelTimeout):
			timeoutOccured = true
			log.Printf("Timeout occurred after %v minutes. Cancelling uninstallation", cancelTimeout.Minutes())
			cancel()
		case <-time.After(quitTimeout):
			log.Print("Uninstallation doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Kyma uninstallation failed due to the timeout")
		}
	}
}

func logStatuses(statusMap map[string]string) {
	log.Print("Components processed so far:")
	for k, v := range statusMap {
		log.Printf("Component: %s, Status: %s", k, v)
	}
}
