//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/namespace"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//TODO: has to be taken from component list! See https://github.com/kyma-incubator/hydroform/issues/181
var kymaNamespaces = []string{"kyma-system", "kyma-integration", "istio-system", "knative-eventing", "natss"}

//Deletion removes Kyma from a cluster
type Deletion struct {
	*core
}

//NewDeletion creates a new Deployment instance for deleting Kyma on a cluster.
func NewDeletion(cfg *config.Config, ob *OverridesBuilder, processUpdates func(ProcessUpdate)) (*Deletion, error) {
	if err := cfg.ValidateDeletion(); err != nil {
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

	core := newCore(cfg, overrides, kubeClient, processUpdates)

	return &Deletion{core}, nil
}

//StartKymaUninstallation removes Kyma from a cluster
func (i *Deletion) StartKymaUninstallation() error {
	_, prerequisitesEng, componentsEng, err := i.getConfig()
	if err != nil {
		return err
	}

	err = i.startKymaUninstallation(prerequisitesEng, componentsEng)
	if err != nil {
		return err
	}

	err = i.deleteKymaNamespaces(kymaNamespaces)
	return err
}

func (i *Deletion) startKymaUninstallation(prerequisitesEng *engine.Engine, componentsEng *engine.Engine) error {
	i.cfg.Log.Info("Kyma uninstallation started")

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancelTimeout := i.cfg.CancelTimeout
	quitTimeout := i.cfg.QuitTimeout

	startTime := time.Now()
	err := i.uninstallComponents(cancelCtx, cancel, UninstallComponents, componentsEng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}
	endTime := time.Now()

	i.cfg.Log.Info("Kyma prerequisites uninstallation")

	cancelTimeout = calculateDuration(startTime, endTime, i.cfg.CancelTimeout)
	quitTimeout = calculateDuration(startTime, endTime, i.cfg.QuitTimeout)

	err = i.uninstallComponents(cancelCtx, cancel, UninstallPreRequisites, prerequisitesEng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}
	ns := namespace.Namespace{
		KubeClient: i.kubeClient,
		Log:        i.cfg.Log,
	}
	err = ns.DeleteInstallerNamespace()
	if err != nil {
		return err
	}

	return nil
}

func (i *Deletion) uninstallComponents(ctx context.Context, cancelFunc context.CancelFunc, phase InstallationPhase, eng *engine.Engine, cancelTimeout time.Duration, quitTimeout time.Duration) error {
	cancelTimeoutChan := time.After(cancelTimeout)
	quitTimeoutChan := time.After(quitTimeout)
	var statusMap = map[string]string{}
	var errCount int = 0
	var timeoutOccured bool = false

	statusChan, err := eng.Uninstall(ctx)
	if err != nil {
		return err
	}

	i.processUpdate(phase, ProcessStart, nil)

	//Await completion
UninstallLoop:
	for {
		select {
		case cmp, ok := <-statusChan:
			if ok {
				i.processUpdateComponent(phase, cmp)
				if cmp.Status == components.StatusError {
					errCount++
				}
				statusMap[cmp.Name] = cmp.Status
			} else {
				if errCount > 0 {
					err := fmt.Errorf("Kyma uninstallation failed due to errors in %d component(s)", errCount)
					i.processUpdate(phase, ProcessExecutionFailure, err)
					i.logStatuses(statusMap)
					return err
				}
				if timeoutOccured {
					err := fmt.Errorf("Kyma uninstallation failed due to the timeout")
					i.processUpdate(phase, ProcessTimeoutFailure, err)
					i.logStatuses(statusMap)
					return err
				}
				break UninstallLoop
			}
		case <-cancelTimeoutChan:
			timeoutOccured = true
			i.cfg.Log.Errorf("Timeout occurred after %v minutes. Cancelling uninstallation", cancelTimeout.Minutes())
			cancelFunc()
		case <-quitTimeoutChan:
			err := fmt.Errorf("Force quit: Kyma uninstallation failed due to the timeout")
			i.processUpdate(phase, ProcessForceQuitFailure, err)
			i.cfg.Log.Error("Uninstallation doesn't stop after it's canceled. Enforcing quit")
			return err
		}
	}
	i.processUpdate(phase, ProcessFinished, nil)
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
