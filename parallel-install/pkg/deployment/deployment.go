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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/prerequisites"
)

var kymaNamespaces = []string{"kyma-system", "kyma-integration", "istio-system", "knative-eventing", "natss"}

type Deployment struct {
	// Contains list of components to install (inclusive pre-requisites)
	componentList *components.ComponentList
	cfg           config.Config
	overrides     Overrides
	// Used to send progress events of a running install/uninstall process
	processUpdates   chan<- ProcessUpdate
	kubeClient       kubernetes.Interface
	metadataProvider metadata.MetadataProvider
}

//NewDeployment should be used to create Deployment instances.
//
//cfg includes configuration parameters for the installer lib
//
//overrides bundles all overrides which have to be considered by Helm
//
//kubeClient is the kubernetes client
//
//processUpdates can be an optional feedback channel provided by the caller
func NewDeployment(cfg config.Config, overrides Overrides, kubeClient kubernetes.Interface, processUpdates chan<- ProcessUpdate) (*Deployment, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	clList, err := components.NewComponentList(cfg.ComponentsListFile)
	if err != nil {
		return nil, err
	}

	metadataProvider := metadata.New(kubeClient)

	return &Deployment{
		componentList:    clList,
		cfg:              cfg,
		overrides:        overrides,
		processUpdates:   processUpdates,
		kubeClient:       kubeClient,
		metadataProvider: metadataProvider,
	}, nil
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
	}

	err = i.metadataProvider.WriteKymaDeployed(attr)
	if err != nil {
		return err
	}

	return nil
}

//StartKymaUninstallation removes Kyma from a cluster
func (i *Deployment) StartKymaUninstallation() error {
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

	if err := i.metadataProvider.DeleteKymaMetadata(); err != nil {
		return err
	}

	return nil
}

//ReadKymaMetadata returns Kyma metadata
func (i *Deployment) ReadKymaMetadata() (*metadata.KymaMetadata, error) {
	return i.metadataProvider.ReadKymaMetadata()
}

func (i *Deployment) startKymaDeployment(prerequisitesProvider components.Provider, overridesProvider overrides.OverridesProvider, eng *engine.Engine) error {
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	i.cfg.Log("Kyma prerequisites deployment")

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

	i.cfg.Log("Kyma deployment")

	cancelTimeout = calculateDuration(startTime, endTime, i.cfg.CancelTimeout)
	quitTimeout = calculateDuration(startTime, endTime, i.cfg.QuitTimeout)

	err = i.deployComponents(cancelCtx, cancel, eng, cancelTimeout, quitTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (i *Deployment) startKymaUninstallation(prerequisitesProvider components.Provider, eng *engine.Engine) error {
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

func (i *Deployment) logStatuses(statusMap map[string]string) {
	i.cfg.Log("Components processed so far:")
	for k, v := range statusMap {
		i.cfg.Log("Component: %s, Status: %s", k, v)
	}
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
			i.cfg.Log("Timeout reached. Cancelling deployment")
			cancelFunc()
		case <-quitTimeoutChan:
			i.processUpdate(InstallPreRequisites, ProcessForceQuitFailure)
			i.cfg.Log("Deployment doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma prerequisites deployment failed due to the timeout")
		}
	}
	i.processUpdate(InstallPreRequisites, ProcessFinished)
	return nil
}

func (i *Deployment) uninstallPrerequisites(ctx context.Context, cancelFunc context.CancelFunc, p []components.KymaComponent, cancelTimeout time.Duration, quitTimeout time.Duration) error {

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
			i.cfg.Log("Timeout occurred after %v minutes. Cancelling deployment", cancelTimeout.Minutes())
			cancelFunc()
		case <-quitTimeoutChan:
			i.processUpdate(InstallComponents, ProcessForceQuitFailure)
			i.cfg.Log("Deployment doesn't stop after it's canceled. Enforcing quit")
			return fmt.Errorf("Force quit: Kyma deployment failed due to the timeout")
		}
	}
	i.processUpdate(InstallComponents, ProcessFinished)
	return nil
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

func (i *Deployment) getConfig() (overrides.OverridesProvider, components.Provider, *engine.Engine, error) {
	overridesMerged, err := i.overrides.Merge()
	if err != nil {
		return nil, nil, nil, err
	}
	overridesProvider, err := overrides.New(i.kubeClient, overridesMerged, i.cfg.Log)

	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to create overrides provider. Exiting...")
	}

	prerequisitesProvider := components.NewPrerequisitesProvider(overridesProvider, i.cfg.ResourcePath, i.componentList.Prerequisites, i.cfg)
	componentsProvider := components.NewComponentsProvider(overridesProvider, i.cfg.ResourcePath, i.componentList.Components, i.cfg)

	engineCfg := engine.Config{
		WorkersCount: i.cfg.WorkersCount,
		Verbose:      false,
	}
	eng := engine.NewEngine(overridesProvider, componentsProvider, engineCfg)

	return overridesProvider, prerequisitesProvider, eng, nil
}

func calculateDuration(start time.Time, end time.Time, duration time.Duration) time.Duration {
	elapsedTime := end.Sub(start)
	return duration - elapsedTime
}

// Send process update event
func (i *Deployment) processUpdate(phase InstallationPhase, event ProcessEvent) {
	if i.processUpdates == nil {
		return
	}
	// fire event
	i.processUpdates <- ProcessUpdate{
		Event:     event,
		Phase:     phase,
		Component: components.KymaComponent{},
	}
}

// Send process update event related to a component
func (i *Deployment) processUpdateComponent(phase InstallationPhase, comp components.KymaComponent) {
	if i.processUpdates == nil {
		return
	}
	// define event type
	event := ProcessRunning
	if comp.Status == components.StatusError {
		event = ProcessExecutionFailure
	}
	// fire event
	i.processUpdates <- ProcessUpdate{
		Event:     event,
		Phase:     phase,
		Component: comp,
	}
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

func (i *Deployment) deleteKymaNamespaces(namespaces []string) error {
	var wg sync.WaitGroup
	wg.Add(len(namespaces))

	finishedCh := make(chan bool)
	errorCh := make(chan error)

	// start deletion in goroutines
	for _, namespace := range namespaces {
		go func(ns string) {
			defer wg.Done()
			errorCh <- i.kubeClient.CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{})
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
	select {
	case <-finishedCh:
	case err := <-errorCh:
		if err != nil {
			errWrapped = errors.Wrap(err, errWrapped.Error())
		}
	}

	return errWrapped
}
