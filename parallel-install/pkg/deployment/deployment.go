//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/metadata"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/prerequisites"
)

//Deployment deploys Kyma on a cluster
type Deployment struct {
	*core
}

//NewDeployment creates a new Deployment instance for deploying Kyma on a cluster.
func NewDeployment(cfg *config.Config, ob *OverridesBuilder, kubeClient kubernetes.Interface, processUpdates chan<- ProcessUpdate) (*Deployment, error) {
	if err := cfg.ValidateDeployment(); err != nil {
		return nil, err
	}

	core, err := newCore(cfg, ob, kubeClient, processUpdates)
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

	attr, err := metadata.NewAttributes(i.cfg.Version, i.cfg.Profile, i.cfg.ComponentsListFile)
	if err != nil {
		return err
	}

	err = i.metadataProvider.WriteKymaDeploymentInProgress(attr)
	if err != nil {
		return err
	}

	overridesProvider, prerequisitesProvider, engine, err := i.getConfig()
	if err != nil {
		return err
	}

	err = i.startKymaDeployment(prerequisitesProvider, overridesProvider, engine)
	if err != nil {
		metaDataErr := i.metadataProvider.WriteKymaDeploymentError(attr, err.Error())
		if metaDataErr != nil {
			return metaDataErr
		}
		return err
	}

	err = i.metadataProvider.WriteKymaDeployed(attr)
	return err
}

func (i *Deployment) startKymaDeployment(prerequisitesProvider components.Provider, overridesProvider overrides.OverridesProvider, eng *engine.Engine) error {
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	i.cfg.Log.Info("Kyma prerequisites deployment")

	prerequisites, err := prerequisitesProvider.GetComponents()
	if err != nil {
		return err
	}
	err = overridesProvider.ReadOverridesFromCluster()
	if err != nil {
		return fmt.Errorf("error while reading overrides: %v", err)
	}

	cancelTimeout := i.cfg.CancelTimeout
	quitTimeout := i.cfg.QuitTimeout

	startTime := time.Now()
	err = i.deployPrerequisites(cancelCtx, cancel, prerequisites, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}
	endTime := time.Now()

	i.cfg.Log.Info("Kyma deployment")

	cancelTimeout = calculateDuration(startTime, endTime, i.cfg.CancelTimeout)
	quitTimeout = calculateDuration(startTime, endTime, i.cfg.QuitTimeout)

	err = i.deployComponents(cancelCtx, cancel, eng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (i *Deployment) deployPrerequisites(ctx context.Context, cancelFunc context.CancelFunc, p []components.KymaComponent, cancelTimeout time.Duration, quitTimeout time.Duration) error {

	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	timeoutOccurred := false

	prereq := prerequisites.Prerequisites{
		Context:       ctx,
		KubeClient:    i.kubeClient,
		Prerequisites: p,
		Log:           i.cfg.Log,
	}
	prereqStatusChan := prereq.InstallPrerequisites()

	i.processUpdate(InstallPreRequisites, ProcessStart)

Prerequisites:
	for {
		select {
		case prerequisiteErr, ok := <-prereqStatusChan:
			if ok {
				if prerequisiteErr != nil {
					i.processUpdate(InstallPreRequisites, ProcessExecutionFailure)
					return fmt.Errorf("Kyma deployment failed due to an error: %s", prerequisiteErr)
				}
			} else {
				if timeoutOccurred {
					i.processUpdate(InstallPreRequisites, ProcessTimeoutFailure)
					return fmt.Errorf("Kyma prerequisites deployment failed due to the timeout")
				}
				break Prerequisites
			}
		case <-cancelTimeoutChan:
			timeoutOccurred = true
			i.cfg.Log.Error("Timeout reached. Cancelling deployment")
			cancelFunc()
		case <-quitTimeoutChan:
			i.processUpdate(InstallPreRequisites, ProcessForceQuitFailure)
			i.cfg.Log.Error("Deployment doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma prerequisites deployment failed due to the timeout")
		}
	}
	i.processUpdate(InstallPreRequisites, ProcessFinished)
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

	i.processUpdate(InstallComponents, ProcessStart)

	//Await completion
InstallLoop:
	for {
		select {
		case cmp, ok := <-statusChan:
			if ok {
				i.processUpdateComponent(InstallComponents, cmp)
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
					i.processUpdate(InstallComponents, ProcessTimeoutFailure)
					i.logStatuses(statusMap)
					return fmt.Errorf("Kyma deployment failed due to the timeout")
				}
				break InstallLoop
			}
		case <-cancelTimeoutChan:
			timeoutOccurred = true
			i.cfg.Log.Errorf("Timeout occurred after %v minutes. Cancelling deployment", cancelTimeout.Minutes())
			cancelFunc()
		case <-quitTimeoutChan:
			i.processUpdate(InstallComponents, ProcessForceQuitFailure)
			i.cfg.Log.Errorf("Deployment doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma deployment failed due to the timeout")
		}
	}
	i.processUpdate(InstallComponents, ProcessFinished)
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

	if err != nil {
		return err
	}

	return nil
}

func (i *Deployment) updateKymaNamespace(namespace string) error {
	_, err := i.kubeClient.CoreV1().Namespaces().Update(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.UpdateOptions{})

	if err != nil {
		return err
	}

	return nil
}
