//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
)

//these components will be removed from component list if running on a local cluster
var incompatibleLocalComponents = []string{"apiserver-proxy", "iam-kubeconfig-service"}

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

	if isK3dCluster(kubeClient) {
		cfg.Log.Infof("Running in K3d cluster: removing incompatible components '%s'", strings.Join(incompatibleLocalComponents, "', '"))
		removeFromComponentList(clList, incompatibleLocalComponents)
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
	i.cfg.Log.Infof("Components processed so far:")
	for k, v := range statusMap {
		i.cfg.Log.Infof("Component: %s, Status: %s", k, v)
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
		Log:          i.cfg.Log,
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

func isK3dCluster(kubeClient kubernetes.Interface) bool {
	nodeList, err := kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return false
	}
	for _, node := range nodeList.Items {
		if strings.HasPrefix(node.GetName(), "k3d-") {
			return true
		}
	}
	return false
}

func removeFromComponentList(cl *components.ComponentList, componentNames []string) {
	for _, compName := range componentNames {
		cl.Remove(compName)
	}
}
