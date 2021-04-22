//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//these components will be removed from component list if running on a local cluster
var incompatibleLocalComponents = []string{"apiserver-proxy", "iam-kubeconfig-service"}

type core struct {
	// Contains list of components to install (inclusive pre-requisites)
	cfg       *config.Config
	overrides *Overrides
	// Used to send progress events of a running install/uninstall process
	processUpdates func(ProcessUpdate)
	kubeClient     kubernetes.Interface
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
func newCore(cfg *config.Config, ob *OverridesBuilder, kubeClient kubernetes.Interface, processUpdates func(ProcessUpdate)) (*core, error) {
	if isK3dCluster(kubeClient) {
		cfg.Log.Infof("Running in K3d cluster: removing incompatible components '%s'", strings.Join(incompatibleLocalComponents, "', '"))
		removeFromComponentList(cfg.ComponentList, incompatibleLocalComponents)
	}

	registerOverridesInterceptors(kubeClient, ob, cfg.Log)

	overrides, err := ob.Build()
	if err != nil {
		return nil, err
	}

	return &core{
		cfg:            cfg,
		overrides:      &overrides,
		processUpdates: processUpdates,
		kubeClient:     kubeClient,
	}, nil
}

func (i *core) logStatuses(statusMap map[string]string) {
	i.cfg.Log.Infof("Components processed so far:")
	for k, v := range statusMap {
		i.cfg.Log.Infof("Component: %s, Status: %s", k, v)
	}
}

func (i *core) getConfig() (overrides.Provider, *engine.Engine, *engine.Engine, error) {
	overridesProvider, err := overrides.New(i.kubeClient, i.overrides.Map(), i.cfg.Log)

	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Failed to create overrides provider: exiting")
	}

	//create KymaComponentMetadataTemplate and set prerequisites flag
	kymaMetadataTpl := helm.NewKymaComponentMetadataTemplate(i.cfg.Version, i.cfg.Profile)
	prerequisitesProvider := components.NewComponentsProvider(overridesProvider, i.cfg, i.cfg.ComponentList.Prerequisites, kymaMetadataTpl.ForPrerequisites())
	componentsProvider := components.NewComponentsProvider(overridesProvider, i.cfg, i.cfg.ComponentList.Components, kymaMetadataTpl.ForComponents())

	prerequisitesEngineCfg := engine.Config{
		// prerequisite components need to be installed sequentially, so only 1 worker should be used
		WorkersCount: 1,
		Log:          i.cfg.Log,
	}
	componentsEngineCfg := engine.Config{
		WorkersCount: i.cfg.WorkersCount,
		Log:          i.cfg.Log,
	}

	prerequisitesEng := engine.NewEngine(overridesProvider, prerequisitesProvider, prerequisitesEngineCfg)
	componentsEng := engine.NewEngine(overridesProvider, componentsProvider, componentsEngineCfg)

	return overridesProvider, prerequisitesEng, componentsEng, nil
}

func calculateDuration(start time.Time, end time.Time, duration time.Duration) time.Duration {
	elapsedTime := end.Sub(start)
	return duration - elapsedTime
}

// Send process update event
func (i *core) processUpdate(phase InstallationPhase, event ProcessEvent, err error) {
	if i.processUpdates == nil {
		return
	}
	//fire callback
	i.processUpdates(ProcessUpdate{
		Event:     event,
		Phase:     phase,
		Component: components.KymaComponent{},
		Error:     err,
	})
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
	//// fire callback
	i.processUpdates(ProcessUpdate{
		Event:     event,
		Phase:     phase,
		Component: comp,
	})
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

func removeFromComponentList(cl *config.ComponentList, componentNames []string) {
	for _, compName := range componentNames {
		cl.Remove(compName)
	}
}

func registerOverridesInterceptors(kubeClient kubernetes.Interface, o *OverridesBuilder, log logger.Interface) {
	//hide certificate data
	o.AddInterceptor([]string{"global.domainName", "global.ingress.domainName"}, NewDomainNameOverrideInterceptor(kubeClient, log))
	o.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey"))
	// make sure we don't install legacy CRDs
	o.AddInterceptor([]string{"global.installCRDs"}, NewInstallLegacyCRDsInterceptor())
	// disable kcproxy
	o.AddInterceptor([]string{"kcproxy.enabled"}, NewIDisableKCProxyInterceptor())
}
