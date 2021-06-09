//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment/k3d"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/jobmanager"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/namespace"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	"k8s.io/client-go/kubernetes"
)

//Deployment deploys Kyma on a cluster
type Deployment struct {
	*core
}

//NewDeployment creates a new Deployment instance for deploying Kyma on a cluster.
func NewDeployment(cfg *config.Config, ob *overrides.Builder, processUpdates func(ProcessUpdate)) (*Deployment, error) {
	if err := cfg.ValidateDeployment(); err != nil {
		return nil, err
	}

	restConfig, err := config.RestConfig(cfg.KubeconfigSource)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	registerOverridesInterceptors(ob, kubeClient, cfg.Log)

	core := newCore(cfg, ob, kubeClient, processUpdates)

	jobmanager.SetConfig(cfg)
	jobmanager.SetKubeClient(kubeClient)

	return &Deployment{core}, nil
}

//StartKymaDeployment deploys Kyma to a cluster
func (d *Deployment) StartKymaDeployment() error {
	//Prepare cluster before Kyma installation
	retryOpts := []retry.Option{
		retry.Delay(time.Duration(d.cfg.BackoffInitialIntervalSeconds) * time.Second),
		retry.Attempts(uint(d.cfg.BackoffMaxElapsedTimeSeconds / d.cfg.BackoffInitialIntervalSeconds)),
		retry.DelayType(retry.FixedDelay),
	}

	preInstallerCfg := inputConfig{
		InstallationResourcePath: d.cfg.InstallationResourcePath,
		Log:                      d.cfg.Log,
		KubeconfigSource:         d.cfg.KubeconfigSource,
		RetryOptions:             retryOpts,
	}

	preInstaller, err := newPreInstaller(preInstallerCfg)
	if err != nil {
		d.cfg.Log.Fatalf("Failed to create Kyma pre-installer: %v", err)
	}

	err = preInstaller.InstallCRDs()
	if err != nil {
		return err
	}

	err = preInstaller.CreateNamespaces()
	if err != nil {
		return err
	}

	overridesProvider, prerequisitesEng, componentsEng, err := d.getConfig()
	if err != nil {
		return err
	}

	return d.startKymaDeployment(overridesProvider, prerequisitesEng, componentsEng)
}

func (d *Deployment) startKymaDeployment(overridesProvider overrides.Provider, prerequisitesEng *engine.Engine, componentsEng *engine.Engine) error {
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d.cfg.Log.Info("Kyma prerequisites deployment")

	err := overridesProvider.ReadOverridesFromCluster()
	if err != nil {
		return fmt.Errorf("error while reading overrides: %v", err)
	}

	isK3s, err := k3d.IsK3dCluster(d.kubeClient)
	if err != nil {
		return err
	}
	if _, err := patchCoreDNS(d.kubeClient, d.overrides, isK3s, d.cfg.Log); err != nil {
		return err
	}

	cancelTimeout := d.cfg.CancelTimeout
	quitTimeout := d.cfg.QuitTimeout

	startTime := time.Now()
	ns := namespace.Namespace{
		KubeClient: d.kubeClient,
		Log:        d.cfg.Log,
	}
	err = ns.DeployInstallerNamespace()
	if err != nil {
		return err
	}
	err = d.deployComponents(cancelCtx, cancel, InstallPreRequisites, prerequisitesEng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}
	endTime := time.Now()

	d.cfg.Log.Info("Kyma deployment")

	cancelTimeout = calculateDuration(startTime, endTime, d.cfg.CancelTimeout)
	quitTimeout = calculateDuration(startTime, endTime, d.cfg.QuitTimeout)

	return d.deployComponents(cancelCtx, cancel, InstallComponents, componentsEng, cancelTimeout, quitTimeout)
}

func (i *Deployment) deployComponents(ctx context.Context, cancelFunc context.CancelFunc, phase InstallationPhase, eng *engine.Engine, cancelTimeout time.Duration, quitTimeout time.Duration) error {
	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	timeoutOccurred := false
	statusMap := map[string]string{}
	errCount := 0

	if phase == InstallPreRequisites {
		jobmanager.ExecutePre(ctx, "global")
	}
	statusChan, err := eng.Deploy(ctx)
	if err != nil {
		return fmt.Errorf("Kyma deployment failed. Error: %v", err)
	}
	i.processUpdate(phase, ProcessStart, nil)

	//Await completion
InstallLoop:
	for {
		select {
		case cmp, ok := <-statusChan:
			if ok {
				i.processUpdateComponent(phase, cmp)
				//Received a status update
				if cmp.Status == components.StatusError {
					// prerequisites fail fast
					if phase == InstallPreRequisites {
						err := errors.Wrapf(cmp.Error, "Error deploying prerequisite: %s", cmp.Name)
						i.processUpdate(phase, ProcessExecutionFailure, err)
						i.logStatuses(statusMap)
						return err
					}
					errCount++
				}
				statusMap[cmp.Name] = cmp.Status
			} else {
				//statusChan is closed
				if errCount > 0 {
					err := fmt.Errorf("Kyma deployment failed due to errors in %d component(s)", errCount)
					i.processUpdate(phase, ProcessExecutionFailure, err)
					i.logStatuses(statusMap)
					return err
				}
				if timeoutOccurred {
					err := fmt.Errorf("Kyma deployment failed due to the timeout")
					i.processUpdate(phase, ProcessTimeoutFailure, err)
					i.logStatuses(statusMap)
					return err
				}
				break InstallLoop
			}
		case <-cancelTimeoutChan:
			timeoutOccurred = true
			i.cfg.Log.Errorf("Timeout occurred after %v minutes. Cancelling deployment", cancelTimeout.Minutes())
			cancelFunc()
		case <-quitTimeoutChan:
			err := fmt.Errorf("Force quit: Kyma deployment failed due to the timeout")
			i.processUpdate(phase, ProcessForceQuitFailure, err)
			i.cfg.Log.Errorf("Deployment doesn't stop after it's canceled. Enforcing quit")
			return err
		}
	}
	// Only will be executed if Kyma deploy was successfull
	//if phase == InstallComponents {
	//	jobmanager.ExecutePost("global")
	//}

	i.processUpdate(phase, ProcessFinished, nil)
	return nil
}

func (i *Deployment) DefaultUpdater() func(update ProcessUpdate) {

	return func(update ProcessUpdate) {

		showCompStatus := func(comp components.KymaComponent) {
			if comp.Name != "" {
				i.cfg.Log.Infof("Status of component '%s': %s", comp.Name, comp.Status)
			}
		}

		switch update.Event {
		case ProcessStart:
			i.cfg.Log.Infof("Starting installation phase '%s'", update.Phase)
		case ProcessRunning:
			showCompStatus(update.Component)
		case ProcessFinished:
			i.cfg.Log.Infof("Finished installation phase '%s' successfully", update.Phase)
		default:
			//any failure case
			i.cfg.Log.Infof("Process failed in phase '%s' with error state '%s':", update.Phase, update.Event)
			showCompStatus(update.Component)
		}
	}
}
