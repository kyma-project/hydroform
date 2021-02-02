//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/metadata"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/prerequisites"
)

//TODO: has to be taken from component list! See https://github.com/kyma-incubator/hydroform/issues/181
var kymaNamespaces = []string{"kyma-system", "kyma-integration", "istio-system", "knative-eventing", "natss"}

//Deletion removes Kyma from a cluster
type Deletion struct {
	*core
}

//NewDeletion creates a new Deployment instance for deleting Kyma on a cluster.
func NewDeletion(cfg config.Config, overrides Overrides, kubeClient kubernetes.Interface, processUpdates chan<- ProcessUpdate) (*Deletion, error) {
	if err := cfg.ValidateDeletion(); err != nil {
		return nil, err
	}

	core, err := newCore(cfg, overrides, kubeClient, processUpdates)
	if err != nil {
		return nil, err
	}

	return &Deletion{core}, nil
}

//StartKymaUninstallation removes Kyma from a cluster
func (i *Deletion) StartKymaUninstallation() error {
	_, prerequisitesProvider, engine, err := i.getConfig()
	if err != nil {
		return err
	}

	attr, err := metadata.NewAttributes(i.cfg.Version, i.cfg.Profile, i.cfg.ComponentsListFile)
	if err != nil {
		return err
	}

	err = i.metadataProvider.WriteKymaUninstallationInProgress(attr)
	if err != nil {
		return err
	}

	err = i.startKymaUninstallation(prerequisitesProvider, engine)
	if err != nil {
		metaDataErr := i.metadataProvider.WriteKymaUninstallationError(attr, err.Error())
		if metaDataErr != nil {
			return metaDataErr
		}
	}

	err = i.deleteKymaNamespaces(kymaNamespaces)
	if err != nil {
		return err
	}

	err = i.metadataProvider.DeleteKymaMetadata()
	if err != nil {
		return err
	}

	return nil
}

func (i *Deletion) startKymaUninstallation(prerequisitesProvider components.Provider, eng *engine.Engine) error {
	i.cfg.Log("Kyma uninstallation started")

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancelTimeout := i.cfg.CancelTimeout
	quitTimeout := i.cfg.QuitTimeout

	startTime := time.Now()
	err := i.uninstallComponents(cancelCtx, cancel, eng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}
	endTime := time.Now()

	i.cfg.Log("Kyma prerequisites uninstallation")

	cancelTimeout = calculateDuration(startTime, endTime, i.cfg.CancelTimeout)
	quitTimeout = calculateDuration(startTime, endTime, i.cfg.QuitTimeout)

	prerequisites, err := prerequisitesProvider.GetComponents()
	if err != nil {
		return err
	}

	err = i.uninstallPrerequisites(cancelCtx, cancel, prerequisites, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (i *Deletion) uninstallPrerequisites(ctx context.Context, cancelFunc context.CancelFunc, p []components.KymaComponent, cancelTimeout time.Duration, quitTimeout time.Duration) error {

	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	timeoutOccurred := false

	prereq := prerequisites.Prerequisites{
		Context:       ctx,
		KubeClient:    i.kubeClient,
		Prerequisites: p,
		Log:           i.cfg.Log,
	}
	prereqStatusChan := prereq.UninstallPrerequisites()

	i.processUpdate(UninstallPreRequisites, ProcessStart)

Prerequisites:
	for {
		select {
		case prerequisiteErr, ok := <-prereqStatusChan:
			if ok {
				if prerequisiteErr != nil {
					i.processUpdate(UninstallPreRequisites, ProcessExecutionFailure)
					i.cfg.Log("Failed to uninstall a prerequisite: %s", prerequisiteErr)
				}
			} else {
				if timeoutOccurred {
					i.processUpdate(UninstallPreRequisites, ProcessTimeoutFailure)
					return fmt.Errorf("Kyma prerequisites uninstallation failed due to the timeout")
				}
				break Prerequisites
			}
		case <-cancelTimeoutChan:
			timeoutOccurred = true
			i.cfg.Log("Timeout reached. Cancelling uninstallation")
			cancelFunc()
		case <-quitTimeoutChan:
			i.processUpdate(UninstallPreRequisites, ProcessForceQuitFailure)
			i.cfg.Log("Uninstallation doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma prerequisites uninstallation failed due to the timeout")
		}
	}
	i.processUpdate(UninstallPreRequisites, ProcessFinished)
	return nil
}

func (i *Deletion) uninstallComponents(ctx context.Context, cancelFunc context.CancelFunc, eng *engine.Engine, cancelTimeout time.Duration, quitTimeout time.Duration) error {
	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	var statusMap = map[string]string{}
	var errCount int = 0
	var timeoutOccured bool = false

	statusChan, err := eng.Uninstall(ctx)
	if err != nil {
		return err
	}

	i.processUpdate(UninstallComponents, ProcessStart)

	//Await completion
UninstallLoop:
	for {
		select {
		case cmp, ok := <-statusChan:
			if ok {
				i.processUpdateComponent(UninstallComponents, cmp)
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
					i.processUpdate(UninstallComponents, ProcessTimeoutFailure)
					i.logStatuses(statusMap)
					return fmt.Errorf("Kyma uninstallation failed due to the timeout")
				}
				break UninstallLoop
			}
		case <-cancelTimeoutChan:
			timeoutOccured = true
			i.cfg.Log("Timeout occurred after %v minutes. Cancelling uninstallation", cancelTimeout.Minutes())
			cancelFunc()
		case <-quitTimeoutChan:
			i.processUpdate(UninstallComponents, ProcessForceQuitFailure)
			i.cfg.Log("Uninstallation doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma uninstallation failed due to the timeout")
		}
	}
	i.processUpdate(UninstallComponents, ProcessFinished)
	return nil
}

func (i *Deletion) deleteKymaNamespaces(namespaces []string) error {
	var wg sync.WaitGroup
	wg.Add(len(namespaces))

	finishedCh := make(chan bool)
	errorCh := make(chan error)

	// start deletion in goroutines
	for _, namespace := range namespaces {
		go func(ns string) {
			defer wg.Done()
			//HACK: drop kyma-system finalizers -> TBD: remove this hack after issue is fixed (https://github.com/kyma-project/kyma/issues/10470)
			if ns == "kyma-system" {
				_, err := i.kubeClient.CoreV1().Namespaces().Finalize(context.Background(), &v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name:       ns,
						Finalizers: []string{},
					},
				}, metav1.UpdateOptions{})
				if err != nil {
					errorCh <- err
				}
			}
			//remove namespace
			if err := i.kubeClient.CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{}); err != nil {
				errorCh <- err
			}
		}(namespace)
	}

	// wait until parallel deletion is finished
	go func() {
		wg.Wait()
		close(errorCh)
		close(finishedCh)
	}()

	// process deletion results
	var errWrapped error
	for {
		select {
		case <-finishedCh:
			return errWrapped
		case err := <-errorCh:
			if err != nil {
				if errWrapped == nil {
					errWrapped = err
				} else {
					errWrapped = errors.Wrap(err, errWrapped.Error())
				}
			}
		}
	}
}
