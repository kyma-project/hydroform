package installation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	prereq "github.com/kyma-incubator/hydroform/parallel-install/pkg/prerequisites"
	"k8s.io/client-go/rest"
)

type Installation struct {
	// Map component > namespace
	Prerequisites [][]string
	// Content of the Installation CR YAML file
	ComponentsYaml string
	// Content of the Helm overrides YAML files
	OverridesYamls []string
	ResourcesPath  string
	Cfg            config.Config
}

type Installer interface {
	//This method will block until installation is finished or an error or timeout occurs.
	//If the installation is not finished in configured config.Config.QuitTimeoutSeconds, the method returns with an error. Some worker goroutines may still be active.
	StartKymaInstallation(kubeconfig *rest.Config) error
	//This method will block until uninstallation is finished or an error or timeout occurs.
	//If the uninstallation is not finished in configured config.Config.QuitTimeoutSeconds, the method returns with an error. Some worker goroutines may still be active.
	StartKymaUninstallation(kubeconfig *rest.Config) error
}

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

func (i *Installation) StartKymaInstallation(prerequisitesProvider components.PrerequisitesProvider, overridesProvider overrides.OverridesProvider, eng *engine.Engine) error {
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

	cancelTimeout := time.Duration(i.Cfg.CancelTimeoutSeconds) * time.Second
	quitTimeout := time.Duration(i.Cfg.QuitTimeoutSeconds) * time.Second
	startTime := time.Now()
	err = i.installPrerequisites(cancelCtx, cancel, prerequisites, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}
	endTime := time.Now()

	i.Cfg.Log("Kyma installation")

	cancelTimeout = calculateDuration(startTime, endTime, i.Cfg.CancelTimeoutSeconds)
	quitTimeout = calculateDuration(startTime, endTime, i.Cfg.QuitTimeoutSeconds)

	err = i.installComponents(cancelCtx, cancel, eng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (i *Installation) StartKymaUninstallation(prerequisitesProvider components.PrerequisitesProvider, overridesProvider overrides.OverridesProvider, eng *engine.Engine) error {
	i.Cfg.Log("Kyma uninstallation started")

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancelTimeout := time.Duration(i.Cfg.CancelTimeoutSeconds) * time.Second
	quitTimeout := time.Duration(i.Cfg.QuitTimeoutSeconds) * time.Second

	startTime := time.Now()
	err := i.uninstallComponents(cancelCtx, cancel, eng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}
	endTime := time.Now()

	log.Print("Kyma prerequisites uninstallation")

	cancelTimeout = calculateDuration(startTime, endTime, i.Cfg.CancelTimeoutSeconds)
	quitTimeout = calculateDuration(startTime, endTime, i.Cfg.QuitTimeoutSeconds)

	prerequisites, err := prerequisitesProvider.GetComponents()
	if err != nil {
		return err
	}

	err = i.uninstallPrerequisites(cancelCtx, cancel, prerequisites, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}

	// TODO: Delete namespace deletion once xip-patch is gone.
	coreClient, err := corev1.NewForConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("Unable to create K8S Client. Error: %v", err)
	}

	err = coreClient.Namespaces().Delete(context.Background(), "kyma-installer", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Unable to delete kyma-installer namespace. Error: %v", err)
	}

	return nil
}

func (i *Installation) logStatuses(statusMap map[string]string) {
	i.Cfg.Log("Components processed so far:")
	for k, v := range statusMap {
		i.Cfg.Log("Component: %s, Status: %s", k, v)
	}
}

func (i *Installation) installPrerequisites(ctx context.Context, cancelFunc context.CancelFunc, p []components.Component, cancelTimeout time.Duration, quitTimeout time.Duration) error {

	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	timeoutOccurred := false

	prereqStatusChan := prereq.InstallPrerequisites(ctx, p)

Prerequisites:
	for {
		select {
		case prerequisiteErr, ok := <-prereqStatusChan:
			if ok {
				if prerequisiteErr != nil {
					return fmt.Errorf("Kyma installation failed due to an error. Look at the preceeding logs to find out more")
				}
			} else {
				if timeoutOccurred {
					return fmt.Errorf("Kyma installation failed due to the timeout")
				}
				break Prerequisites
			}
		case <-cancelTimeoutChan:
			timeoutOccurred = true
			i.Cfg.Log("Timeout reached. Cancelling installation")
			cancelFunc()
		case <-quitTimeoutChan:
			i.Cfg.Log("Installation doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma installation failed due to the timeout")
		}
	}
	return nil
}

func (i *Installation) uninstallPrerequisites(ctx context.Context, cancelFunc context.CancelFunc, p []components.Component, cancelTimeout time.Duration, quitTimeout time.Duration) error {

	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	timeoutOccurred := false

	prereqStatusChan := prereq.UninstallPrerequisites(ctx, p)

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
					return fmt.Errorf("Kyma installation failed due to the timeout")
				}
				break Prerequisites
			}
		case <-cancelTimeoutChan:
			timeoutOccurred = true
			i.Cfg.Log("Timeout reached. Cancelling installation")
			cancelFunc()
		case <-quitTimeoutChan:
			i.Cfg.Log("Installation doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma installation failed due to the timeout")
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
		case <-time.After(cancelTimeout):
			timeoutOccured = true
			i.Cfg.Log("Timeout occurred after %v minutes. Cancelling uninstallation", cancelTimeout.Minutes())
			cancelFunc()
		case <-time.After(quitTimeout):
			i.Cfg.Log("Uninstallation doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Kyma uninstallation failed due to the timeout")
		}
	}
	return nil
}

func calculateDuration(start time.Time, end time.Time, duration int) time.Duration {
	elapsedTime := end.Sub(start)
	return time.Duration(duration)*time.Second - elapsedTime
}
