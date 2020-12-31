//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/metadata"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	prereq "github.com/kyma-incubator/hydroform/parallel-install/pkg/prerequisites"
)

type Deployment struct {
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
	//StartKymaDeployment deploys Kyma on the cluster.
	//This method will block until deployment is finished or an error or timeout occurs.
	//If the deployment is not finished in configured config.Config.QuitTimeout,
	//the method returns with an error. Some worker goroutines may still run in the background.
	StartKymaDeployment(kubeClient kubernetes.Interface) error
	//StartKymaUninstallation uninstalls Kyma from the cluster.
	//This method will block until uninstallation is finished or an error or timeout occurs.
	//If the uninstallation is not finished in configured config.Config.QuitTimeout,
	//the method returns with an error. Some worker goroutines may still run in the background.
	StartKymaUninstallation(kubeClient kubernetes.Interface) error
}

//NewDeployment should be used to create Deployment instances.
//
//prerequisites is a slice of pairs: [component-name, namespace]
//
//componentsYaml is a string containing an Deployment CR in the YAML format.
//
//overridesYamls contains data in the YAML format.
//See overrides.New for details about the overrides contract.
//
//resourcesPath is a local filesystem path where components' charts are located.
func NewDeployment(prerequisites [][]string, componentsYaml string, overridesYamls []string, resourcesPath string, cfg config.Config) (*Deployment, error) {
	if resourcesPath == "" {
		return nil, fmt.Errorf("Unable to create Deployment. Resource path is required.")
	}
	if componentsYaml == "" {
		return nil, fmt.Errorf("Unable to create Deployment. Components YAML file content is required.")
	}

	return &Deployment{
		Prerequisites:  prerequisites,
		ComponentsYaml: componentsYaml,
		OverridesYamls: overridesYamls,
		ResourcesPath:  resourcesPath,
		Cfg:            cfg,
	}, nil
}

//StartKymaDeployment implements the Installer.StartKymaDeployment contract.
func (i *Deployment) StartKymaDeployment(kubeClient kubernetes.Interface) error {
	metadataProvider := metadata.New(kubeClient)
	//TODO refactoring needed: something like metadatProvider.WriteKymaProgress
	meta := &metadata.KymaMetadata{
		Version: i.Cfg.Version,
		Profile: i.Cfg.Profile,
		Status: "Deployment in progress",
	}

	err := metadataProvider.WriteKKymaMetadata(meta)
	if err != nil {
		return err
	}

	overridesProvider, prerequisitesProvider, engine, err := i.getConfig(kubeClient)
	if err != nil {
		return err
	}

	err = i.startKymaDeployment(kubeClient, prerequisitesProvider, overridesProvider, engine)
	if err != nil {
		//TODO refactoring needed: something like metadatProvider.WriteKymaError
		meta := &metadata.KymaMetadata{
			Version: i.Cfg.Version,
			Profile: i.Cfg.Profile,
			Status: "Deployment error",
			Reason: err.Error(),
		}

		err := metadataProvider.WriteKKymaMetadata(meta)
		if err != nil {
			return err
		}
	}

	//TODO refactoring needed: something like metadatProvider.WriteKymaDeployed
	meta = &metadata.KymaMetadata{
		Version: i.Cfg.Version,
		Profile: i.Cfg.Profile,
		Status: "Deployed",
	}

	err = metadataProvider.WriteKKymaMetadata(meta)
	if err != nil {
		return err
	}

	return nil
}

//StartKymaUninstallation implements the Installer.StartKymaUninstallation contract.
func (i *Deployment) StartKymaUninstallation(kubeClient kubernetes.Interface) error {
	_, prerequisitesProvider, engine, err := i.getConfig(kubeClient)
	if err != nil {
		return err
	}

	metadataProvider := metadata.New(kubeClient)
	//TODO refactoring needed: something like metadatProvider.WriteKymaProgress
	meta := &metadata.KymaMetadata{
		Version: i.Cfg.Version,
		Profile: i.Cfg.Profile,
		Status: "Uninstallation in progress",
	}

	err = metadataProvider.WriteKKymaMetadata(meta)
	if err != nil {
		return err
	}

	err = i.startKymaUninstallation(kubeClient, prerequisitesProvider, engine)
	if err!=nil{
		//TODO refactoring needed: something like metadatProvider.WriteKymaError
		meta := &metadata.KymaMetadata{
			Version: i.Cfg.Version,
			Profile: i.Cfg.Profile,
			Status: "Uninstallation error",
			Reason: err.Error(),
		}

		err := metadataProvider.WriteKKymaMetadata(meta)
		if err != nil {
			return err
		}
	}

	err = metadataProvider.DeleteKymaMetadata()
	if err != nil {
		return err
	}

	return nil
}

func (i *Deployment) startKymaDeployment(kubeClient kubernetes.Interface, prerequisitesProvider components.Provider, overridesProvider overrides.OverridesProvider, eng *engine.Engine) error {
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	i.Cfg.Log("Kyma prerequisites deployment")

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
	err = i.deployPrerequisites(cancelCtx, cancel, kubeClient, prerequisites, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}
	endTime := time.Now()

	i.Cfg.Log("Kyma deployment")

	cancelTimeout = calculateDuration(startTime, endTime, i.Cfg.CancelTimeout)
	quitTimeout = calculateDuration(startTime, endTime, i.Cfg.QuitTimeout)

	err = i.deployComponents(cancelCtx, cancel, eng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (i *Deployment) startKymaUninstallation(kubeClient kubernetes.Interface, prerequisitesProvider components.Provider, eng *engine.Engine) error {
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

func (i *Deployment) logStatuses(statusMap map[string]string) {
	i.Cfg.Log("Components processed so far:")
	for k, v := range statusMap {
		i.Cfg.Log("Component: %s, Status: %s", k, v)
	}
}

func (i *Deployment) deployPrerequisites(ctx context.Context, cancelFunc context.CancelFunc, kubeClient kubernetes.Interface, p []components.KymaComponent, cancelTimeout time.Duration, quitTimeout time.Duration) error {

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
					return fmt.Errorf("Kyma deployment failed due to an error: %s", prerequisiteErr)
				}
			} else {
				if timeoutOccurred {
					return fmt.Errorf("Kyma prerequisites deployment failed due to the timeout")
				}
				break Prerequisites
			}
		case <-cancelTimeoutChan:
			timeoutOccurred = true
			i.Cfg.Log("Timeout reached. Cancelling deployment")
			cancelFunc()
		case <-quitTimeoutChan:
			i.Cfg.Log("Deployment doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma prerequisites deployment failed due to the timeout")
		}
	}
	return nil
}

func (i *Deployment) uninstallPrerequisites(ctx context.Context, cancelFunc context.CancelFunc, kubeClient kubernetes.Interface, p []components.KymaComponent, cancelTimeout time.Duration, quitTimeout time.Duration) error {

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

func (i *Deployment) deployComponents(ctx context.Context, cancelFunc context.CancelFunc, eng *engine.Engine, cancelTimeout time.Duration, quitTimeout time.Duration) error {
	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	timeoutOccurred := false
	statusMap := map[string]string{}
	errCount := 0

	statusChan, err := eng.Deploy(ctx)
	if err != nil {
		return fmt.Errorf("Kyma deployment failed. Error: %v", err)
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
					return fmt.Errorf("Kyma deployment failed due to errors in %d component(s)", errCount)
				}
				if timeoutOccurred {
					i.logStatuses(statusMap)
					return fmt.Errorf("Kyma deployment failed due to the timeout")
				}
				return nil
			}
		case <-cancelTimeoutChan:
			timeoutOccurred = true
			i.Cfg.Log("Timeout occurred after %v minutes. Cancelling deployment", cancelTimeout.Minutes())
			cancelFunc()
		case <-quitTimeoutChan:
			i.Cfg.Log("Deployment doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma deployment failed due to the timeout")
		}
	}
}

func (i *Deployment) uninstallComponents(ctx context.Context, cancelFunc context.CancelFunc, eng *engine.Engine, cancelTimeout time.Duration, quitTimeout time.Duration) error {
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

func (i *Deployment) getConfig(kubeClient kubernetes.Interface) (overrides.OverridesProvider, components.Provider, *engine.Engine, error) {
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
