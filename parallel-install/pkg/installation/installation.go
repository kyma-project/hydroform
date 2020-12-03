//Package installation provides a top-level API to control Kyma installation and uninstallation.
package installation

import (
	"context"
	"fmt"
	"log"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	prereq "github.com/kyma-incubator/hydroform/parallel-install/pkg/prerequisites"
)

type Installation struct {
	// Slice of pairs: [component, namespace]
	Prerequisites [][]string
	// Content of the Installation CR YAML file
	ComponentsYaml string
	// Content of the Helm overrides YAML files
	OverridesYamls []string
	// Root dir in a local filesystem with subdirectories containing components' Helm charts
	ResourcesPath string
	Cfg           config.Config
}

type Installer interface {
	//StartKymaInstallation installs Kyma on the cluster.
	//This method will block until installation is finished or an error or timeout occurs.
	//If the installation is not finished in configured config.Config.QuitTimeout,
	//the method returns with an error. Some worker goroutines may still run in the background.
	StartKymaInstallation(kubeClient kubernetes.Interface) error
	//StartKymaUninstallation uninstalls Kyma from the cluster.
	//This method will block until uninstallation is finished or an error or timeout occurs.
	//If the uninstallation is not finished in configured config.Config.QuitTimeout,
	//the method returns with an error. Some worker goroutines may still run in the background.
	StartKymaUninstallation(kubeClient kubernetes.Interface) error
}

//NewInstallation should be used to create Installation instances
//
//prerequisites is a slice of pairs: [component-name, namespace]
//
//componentsYaml is a string containing Installation CR in yaml format.
//
//overridesYamls contains data in yaml format.
//See overrides.New for details about overrides contract.
//
//resourcesPath is a local filesystem path where components' charts are located
func NewInstallation(prerequisites [][]string, componentsYaml string, overridesYamls []string, resourcesPath string, cfg config.Config) (*Installation, error) {
	if resourcesPath == "" {
		return nil, fmt.Errorf("Unable to create Installation. Resource path is required.")
	}
	if componentsYaml == "" {
		return nil, fmt.Errorf("Unable to create Installation. Components YAML file content is required.")
	}

	return &Installation{
		Prerequisites:  prerequisites,
		ComponentsYaml: componentsYaml,
		OverridesYamls: overridesYamls,
		ResourcesPath:  resourcesPath,
		Cfg:            cfg,
	}, nil
}

//StartKymaInstallation implements Installer.StartKymaInstallation contract
func (i *Installation) StartKymaInstallation(kubeClient kubernetes.Interface) error {
	overridesProvider, prerequisitesProvider, engine, err := i.getConfig(kubeClient)
	if err != nil {
		return err
	}
	return i.startKymaInstallation(kubeClient, prerequisitesProvider, overridesProvider, engine)
}

//StartKymaUninstallation implements Installer.StartKymaUninstallation contract
func (i *Installation) StartKymaUninstallation(kubeClient kubernetes.Interface) error {
	_, prerequisitesProvider, engine, err := i.getConfig(kubeClient)
	if err != nil {
		return err
	}
	return i.startKymaUninstallation(kubeClient, prerequisitesProvider, engine)
}

func (i *Installation) startKymaInstallation(kubeClient kubernetes.Interface, prerequisitesProvider components.Provider, overridesProvider overrides.OverridesProvider, eng *engine.Engine) error {
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	i.Cfg.Log("Kyma prerequisites installation")

	prerequisites, err := prerequisitesProvider.GetComponents()
	if err != nil {
		return err
	}
	err = overridesProvider.ReadOverridesFromCluster()
	if err != nil {
		return fmt.Errorf("error while reading overrides: %v", err)
	}

	cancelTimeout := i.Cfg.CancelTimeout
	quitTimeout := i.Cfg.QuitTimeout

	startTime := time.Now()
	err = i.installPrerequisites(cancelCtx, cancel, kubeClient, prerequisites, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}
	endTime := time.Now()

	i.Cfg.Log("Kyma installation")

	cancelTimeout = calculateDuration(startTime, endTime, i.Cfg.CancelTimeout)
	quitTimeout = calculateDuration(startTime, endTime, i.Cfg.QuitTimeout)

	err = i.installComponents(cancelCtx, cancel, eng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (i *Installation) startKymaUninstallation(kubeClient kubernetes.Interface, prerequisitesProvider components.Provider, eng *engine.Engine) error {
	i.Cfg.Log("Kyma uninstallation started")

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancelTimeout := i.Cfg.CancelTimeout
	quitTimeout := i.Cfg.QuitTimeout

	startTime := time.Now()
	err := i.uninstallComponents(cancelCtx, cancel, eng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}
	endTime := time.Now()

	log.Print("Kyma prerequisites uninstallation")

	cancelTimeout = calculateDuration(startTime, endTime, i.Cfg.CancelTimeout)
	quitTimeout = calculateDuration(startTime, endTime, i.Cfg.QuitTimeout)

	prerequisites, err := prerequisitesProvider.GetComponents()
	if err != nil {
		return err
	}

	err = i.uninstallPrerequisites(cancelCtx, cancel, kubeClient, prerequisites, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (i *Installation) logStatuses(statusMap map[string]string) {
	i.Cfg.Log("Components processed so far:")
	for k, v := range statusMap {
		i.Cfg.Log("Component: %s, Status: %s", k, v)
	}
}

func (i *Installation) installPrerequisites(ctx context.Context, cancelFunc context.CancelFunc, kubeClient kubernetes.Interface, p []components.Component, cancelTimeout time.Duration, quitTimeout time.Duration) error {

	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	timeoutOccurred := false

	prereqStatusChan := prereq.InstallPrerequisites(ctx, kubeClient, p)

Prerequisites:
	for {
		select {
		case prerequisiteErr, ok := <-prereqStatusChan:
			if ok {
				if prerequisiteErr != nil {
					return fmt.Errorf("Kyma installation failed due to an error: %s", prerequisiteErr)
				}
			} else {
				if timeoutOccurred {
					return fmt.Errorf("Kyma prerequisites installation failed due to the timeout")
				}
				break Prerequisites
			}
		case <-cancelTimeoutChan:
			timeoutOccurred = true
			i.Cfg.Log("Timeout reached. Cancelling installation")
			cancelFunc()
		case <-quitTimeoutChan:
			i.Cfg.Log("Installation doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma prerequisites installation failed due to the timeout")
		}
	}
	return nil
}

func (i *Installation) uninstallPrerequisites(ctx context.Context, cancelFunc context.CancelFunc, kubeClient kubernetes.Interface, p []components.Component, cancelTimeout time.Duration, quitTimeout time.Duration) error {

	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	timeoutOccurred := false

	prereqStatusChan := prereq.UninstallPrerequisites(ctx, kubeClient, p)

Prerequisites:
	for {
		select {
		case prerequisiteErr, ok := <-prereqStatusChan:
			if ok {
				if prerequisiteErr != nil {
					config.Log("Failed to uninstall a prerequisite: %s", prerequisiteErr)
				}
			} else {
				if timeoutOccurred {
					return fmt.Errorf("Kyma prerequisites uninstallation failed due to the timeout")
				}
				break Prerequisites
			}
		case <-cancelTimeoutChan:
			timeoutOccurred = true
			i.Cfg.Log("Timeout reached. Cancelling uninstallation")
			cancelFunc()
		case <-quitTimeoutChan:
			i.Cfg.Log("Uninstallation doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma prerequisites uninstallation failed due to the timeout")
		}
	}
	return nil
}

func (i *Installation) installComponents(ctx context.Context, cancelFunc context.CancelFunc, eng *engine.Engine, cancelTimeout time.Duration, quitTimeout time.Duration) error {
	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	timeoutOccurred := false
	statusMap := map[string]string{}
	errCount := 0

	statusChan, err := eng.Install(ctx)
	if err != nil {
		return fmt.Errorf("Kyma installation failed. Error: %v", err)
	}

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
					i.logStatuses(statusMap)
					return fmt.Errorf("Kyma installation failed due to errors in %d component(s)", errCount)
				}
				if timeoutOccurred {
					i.logStatuses(statusMap)
					return fmt.Errorf("Kyma installation failed due to the timeout")
				}
				return nil
			}
		case <-cancelTimeoutChan:
			timeoutOccurred = true
			i.Cfg.Log("Timeout occurred after %v minutes. Cancelling installation", cancelTimeout.Minutes())
			cancelFunc()
		case <-quitTimeoutChan:
			i.Cfg.Log("Installation doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma installation failed due to the timeout")
		}
	}
}

func (i *Installation) uninstallComponents(ctx context.Context, cancelFunc context.CancelFunc, eng *engine.Engine, cancelTimeout time.Duration, quitTimeout time.Duration) error {
	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	var statusMap = map[string]string{}
	var errCount int = 0
	var timeoutOccured bool = false

	statusChan, err := eng.Uninstall(ctx)
	if err != nil {
		return err
	}
	//Await completion
Loop:
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
					i.logStatuses(statusMap)
					return fmt.Errorf("Kyma uninstallation failed due to errors in %d component(s)", errCount)
				}
				if timeoutOccured {
					i.logStatuses(statusMap)
					return fmt.Errorf("Kyma uninstallation failed due to the timeout")
				}
				break Loop
			}
		case <-cancelTimeoutChan:
			timeoutOccured = true
			i.Cfg.Log("Timeout occurred after %v minutes. Cancelling uninstallation", cancelTimeout.Minutes())
			cancelFunc()
		case <-quitTimeoutChan:
			i.Cfg.Log("Uninstallation doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma uninstallation failed due to the timeout")
		}
	}
	return nil
}

func (i *Installation) getConfig(kubeClient kubernetes.Interface) (overrides.OverridesProvider, components.Provider, *engine.Engine, error) {
	overridesProvider, err := overrides.New(kubeClient, i.OverridesYamls, i.Cfg.Log)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to create overrides provider. Exiting...")
	}

	prerequisitesProvider := components.NewPrerequisitesProvider(overridesProvider, i.ResourcesPath, i.Prerequisites, i.Cfg)
	componentsProvider := components.NewComponentsProvider(overridesProvider, i.ResourcesPath, i.ComponentsYaml, i.Cfg)

	engineCfg := engine.Config{WorkersCount: i.Cfg.WorkersCount}
	eng := engine.NewEngine(overridesProvider, componentsProvider, engineCfg)

	return overridesProvider, prerequisitesProvider, eng, nil
}

func calculateDuration(start time.Time, end time.Time, duration time.Duration) time.Duration {
	elapsedTime := end.Sub(start)
	return duration - elapsedTime
}
