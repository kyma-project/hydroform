//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/metadata"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
)

type core struct {
	// Contains list of components to install (inclusive pre-requisites)
	componentList *components.ComponentList
	cfg           config.Config
	overrides     Overrides
	// Used to send progress events of a running install/uninstall process
	processUpdates   chan<- ProcessUpdate
	kubeClient       kubernetes.Interface
	metadataProvider metadata.MetadataProvider
}

//new creates a new core instance
//
//cfg includes configuration parameters for the installer lib
//
//overrides bundles all overrides which have to be considered by Helm
//
//kubeClient is the kubernetes client
//
//processUpdates can be an optional feedback channel provided by the caller
func newCore(cfg config.Config, overrides Overrides, kubeClient kubernetes.Interface, processUpdates chan<- ProcessUpdate) (*core, error) {
	clList, err := components.NewComponentList(cfg.ComponentsListFile)
	if err != nil {
		return nil, err
	}

	metadataProvider := metadata.New(kubeClient)

	return &core{
		componentList:    clList,
		cfg:              cfg,
		overrides:        overrides,
		processUpdates:   processUpdates,
		kubeClient:       kubeClient,
		metadataProvider: metadataProvider,
	}, nil
}

//ReadKymaMetadata returns Kyma metadata
func (i *core) ReadKymaMetadata() (*metadata.KymaMetadata, error) {
	return i.metadataProvider.ReadKymaMetadata()
}

func (i *core) logStatuses(statusMap map[string]string) {
	i.cfg.Log("Components processed so far:")
	for k, v := range statusMap {
		i.cfg.Log("Component: %s, Status: %s", k, v)
	}
}

func (i *core) getConfig() (overrides.OverridesProvider, components.Provider, *engine.Engine, error) {
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
func (i *core) processUpdate(phase InstallationPhase, event ProcessEvent) {
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
func (i *core) processUpdateComponent(phase InstallationPhase, comp components.KymaComponent) {
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
