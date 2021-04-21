//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/namespace"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//Deployment deploys Kyma on a cluster
type Deployment struct {
	*core
}

//NewDeployment creates a new Deployment instance for deploying Kyma on a cluster.
func NewDeployment(cfg *config.Config, ob *OverridesBuilder, processUpdates func(ProcessUpdate)) (*Deployment, error) {
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

	overrides, err := registerOverridesInterceptors(ob, kubeClient, cfg.Log)
	if err != nil {
		return nil, err
	}

	core, err := newCore(cfg, overrides, kubeClient, processUpdates)
	if err != nil {
		return nil, err
	}

	return &Deployment{core}, nil
}

//StartKymaDeployment deploys Kyma to a cluster
func (i *Deployment) StartKymaDeployment() error {
	err := i.deployKymaNamespaces(kymaNamespaces)
	if err != nil {
		return err
	}

	overridesProvider, prerequisitesEng, componentsEng, err := i.getConfig()
	if err != nil {
		return err
	}

	return i.startKymaDeployment(overridesProvider, prerequisitesEng, componentsEng)
}

func (i *Deployment) startKymaDeployment(overridesProvider overrides.Provider, prerequisitesEng *engine.Engine, componentsEng *engine.Engine) error {
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	i.cfg.Log.Info("Kyma prerequisites deployment")

	err := overridesProvider.ReadOverridesFromCluster()
	if err != nil {
		return fmt.Errorf("error while reading overrides: %v", err)
	}

	cancelTimeout := i.cfg.CancelTimeout
	quitTimeout := i.cfg.QuitTimeout

	startTime := time.Now()
	ns := namespace.Namespace{
		KubeClient: i.kubeClient,
		Log:        i.cfg.Log,
	}
	err = ns.DeployInstallerNamespace()
	if err != nil {
		return err
	}
	err = i.deployComponents(cancelCtx, cancel, InstallPreRequisites, prerequisitesEng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}
	endTime := time.Now()

	i.cfg.Log.Info("Kyma deployment")

	cancelTimeout = calculateDuration(startTime, endTime, i.cfg.CancelTimeout)
	quitTimeout = calculateDuration(startTime, endTime, i.cfg.QuitTimeout)

	return i.deployComponents(cancelCtx, cancel, InstallComponents, componentsEng, cancelTimeout, quitTimeout)
}

func (i *Deployment) deployComponents(ctx context.Context, cancelFunc context.CancelFunc, phase InstallationPhase, eng *engine.Engine, cancelTimeout time.Duration, quitTimeout time.Duration) error {
	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	timeoutOccurred := false
	statusMap := map[string]string{}
	errCount := 0

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
	i.processUpdate(phase, ProcessFinished, nil)
	return nil
}

func (i *Deployment) deployKymaNamespaces(namespaces []string) error {
	for _, namespace := range namespaces {
		_, err := i.kubeClient.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})

		if err != nil {
			if apierrors.IsNotFound(err) {
				nsErr := i.createKymaNamespace(namespace)
				if nsErr != nil {
					return nsErr
				}
			} else {
				return err
			}
		} else {
			nsErr := i.updateKymaNamespace(namespace)
			if nsErr != nil {
				return nsErr
			}
		}
	}
	return nil
}

func (i *Deployment) createKymaNamespace(namespace string) error {
	_, err := i.kubeClient.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})

	return err
}

func (i *Deployment) updateKymaNamespace(namespace string) error {
	_, err := i.kubeClient.CoreV1().Namespaces().Update(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.UpdateOptions{})

	return err
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
