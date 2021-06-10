//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apixv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

//Deletion removes Kyma from a cluster
type Deletion struct {
	*core
	mp              *helm.KymaMetadataProvider
	apixClient      apixv1beta1client.ApiextensionsV1beta1Interface
	dClient         dynamic.Interface
	resourceManager ResourceManager
	retryOptions    []retry.Option
}

//NewDeletion creates a new Deployment instance for deleting Kyma on a cluster.
func NewDeletion(cfg *config.Config, ob *overrides.Builder, processUpdates func(ProcessUpdate), retryOptions []retry.Option) (*Deletion, error) {
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

	dClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	apixClient, err := apixv1beta1client.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	resourceManager, err := NewDefaultResourceManager(cfg.KubeconfigSource, cfg.Log, retryOptions)
	if err != nil {
		return nil, err
	}

	registerOverridesInterceptors(ob, kubeClient, cfg.Log)

	core := newCore(cfg, ob, kubeClient, processUpdates)

	mp, err := helm.NewKymaMetadataProvider(cfg.KubeconfigSource)
	if err != nil {
		return nil, err
	}

	return &Deletion{core, mp, apixClient, dClient, resourceManager, retryOptions}, nil
}

//StartKymaUninstallation removes Kyma from a cluster
func (i *Deletion) StartKymaUninstallation() error {
	_, prerequisitesEng, componentsEng, err := i.getConfig()
	if err != nil {
		return err
	}

	return i.startKymaUninstallation(prerequisitesEng, componentsEng)
}

func (i *Deletion) startKymaUninstallation(prerequisitesEng *engine.Engine, componentsEng *engine.Engine) error {
	i.cfg.Log.Info("Kyma uninstallation started")

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancelTimeout := i.cfg.CancelTimeout
	quitTimeout := i.cfg.QuitTimeout

	namespaces, err := i.mp.Namespaces()
	if err != nil {
		return err
	}
	//TODO: Delete this when kyma-installer is not used any more.
	namespaces = append(namespaces, "kyma-installer")

	startTime := time.Now()
	err = i.uninstallComponents(cancelCtx, cancel, UninstallComponents, componentsEng, cancelTimeout, quitTimeout)
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

	err = i.deleteKymaNamespaces(namespaces)
	if err != nil {
		return err
	}

	return i.deleteKymaCrds()
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

func (i *Deletion) deleteKymaCrds() error {
	i.cfg.Log.Infof("Uninstalling CRDs labeled with: %s=%s", config.LABEL_KEY_ORIGIN, config.LABEL_VALUE_KYMA)

	selector, err := i.prepareKymaCrdLabelSelector()
	if err != nil {
		return err
	}

	gvks := i.retrieveKymaCrdGvks()
	for _, gvk := range gvks {
		i.cfg.Log.Infof("Uninstalling CRDs that belong to apiVersion: %s/%s", gvk.Group, gvk.Version)
		err = i.resourceManager.DeleteCollectionOfResources(gvk, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			i.cfg.Log.Error(err)
		}
	}

	i.cfg.Log.Infof("Kyma CRDs successfully uninstalled")

	return nil
}

func (i *Deletion) prepareKymaCrdLabelSelector() (selector labels.Selector, err error) {
	kymaCrdReq, err := labels.NewRequirement(config.LABEL_KEY_ORIGIN, selection.Equals, []string{config.LABEL_VALUE_KYMA})
	if err != nil {
		return nil, errors.Wrap(err, "Error occurred when preparing Kyma CRD label selector")
	}

	selector = labels.NewSelector()
	selector = selector.Add(*kymaCrdReq)
	return selector, nil
}

func (i *Deletion) retrieveKymaCrdGvks() []schema.GroupVersionKind {
	crdGvkV1Beta1 := i.crdGvkWith("v1beta1")
	crdGvkV1 := i.crdGvkWith("v1")
	return []schema.GroupVersionKind{crdGvkV1Beta1, crdGvkV1}
}

func (i *Deletion) crdGvkWith(version string) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: version,
		Kind:    "customresourcedefinition",
	}
}

func (i *Deletion) deleteKymaNamespaces(namespaces []string) error {
	var wg sync.WaitGroup
	wg.Add(len(namespaces))

	finishedCh := make(chan bool)
	errorCh := make(chan error)

	// start deletion in goroutines
	for _, namespace := range namespaces {
		err := retry.Do(func() error {
			// Check if there are any running Pods left on the namespace
			pods, err := i.kubeClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				errorCh <- err
			}

			if len(pods.Items) > 0 {
				for _, pod := range pods.Items {
					if pod.Status.Phase == v1.PodRunning {
						return errors.New(fmt.Sprintf("Namespace %s could not be deleted because of the running Pod: %s. Trying again..", namespace, pod.Name))
					}
				}
			}
			return nil
		}, i.retryOptions...)

		if err != nil {
			i.cfg.Log.Infof("Namespace %s could not be deleted because of running Pod(s)", namespace)
			wg.Done()
			continue
		}

		go func(ns string) {
			defer wg.Done()
			// All the hacks below should be deleted after this issue is done: https://github.com/kyma-project/kyma/issues/11298
			//HACK: Delete finalizers of leftover Secret
			secrets, err := i.kubeClient.CoreV1().Secrets(ns).List(context.Background(), metav1.ListOptions{LabelSelector: "serverless.kyma-project.io/config=credentials"})
			if err != nil && !apierr.IsNotFound(err) {
				errorCh <- err
			}
			if secrets != nil {
				for _, secret := range secrets.Items {
					secret.SetFinalizers(nil)
					if _, err := i.kubeClient.CoreV1().Secrets(ns).Update(context.Background(), &secret, metav1.UpdateOptions{}); err != nil {
						errorCh <- err
					}
					i.cfg.Log.Infof("Deleted finalizer from Secret: %s", secret.Name)
				}
			}

			//HACK: Delete finalizers of leftover Custom Resources
			crds, err := i.apixClient.CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{})
			if err != nil && !apierr.IsNotFound(err) {
				errorCh <- err
			}

			if crds != nil {
				for _, crd := range crds.Items {
					customResource := schema.GroupVersionResource{
						Group:    crd.Spec.Group,
						Version:  crd.Spec.Version,
						Resource: crd.Spec.Names.Plural,
					}

					customResourceList, err := i.dClient.Resource(customResource).Namespace(ns).List(context.Background(), metav1.ListOptions{})
					if err != nil && !apierr.IsNotFound(err) {
						errorCh <- err
					}
					if customResourceList != nil {
						for _, cr := range customResourceList.Items {
							cr.SetFinalizers(nil)
							_, err := i.dClient.Resource(customResource).Namespace(ns).Update(context.Background(), &cr, metav1.UpdateOptions{})
							if err != nil {
								errorCh <- err
							}
							i.cfg.Log.Infof("Deleted finalizer from %s: %s", cr.GetKind(), cr.GetName())
						}
					}
				}
			}

			err = retry.Do(func() error {
				//remove namespace
				if err := i.kubeClient.CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{}); err != nil && !apierr.IsNotFound(err) {
					errorCh <- err
				}

				nsT, err := i.kubeClient.CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
				if err != nil && !apierr.IsNotFound(err) {
					errorCh <- err
				} else if apierr.IsNotFound(err) {
					return nil
				}
				i.cfg.Log.Infof("Namespace '%s' still exists with Phase: '%s'", nsT.Name, nsT.Status.Phase)
				return fmt.Errorf("Namespace '%s' still exists with Phase: '%s'", nsT.Name, nsT.Status.Phase)
			}, i.retryOptions...)
			if err != nil {
				errorCh <- err
			}
			i.cfg.Log.Infof("Namespace '%s' is removed", ns)
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
