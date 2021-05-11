//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"context"
	"strings"
	"time"
	"unicode"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/avast/retry-go"
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

type core struct {
	// Contains list of components to install (inclusive pre-requisites)
	cfg       *config.Config
	overrides *OverridesBuilder
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
func newCore(cfg *config.Config, overrides *OverridesBuilder, kubeClient kubernetes.Interface, processUpdates func(ProcessUpdate)) *core {
	return &core{
		cfg:            cfg,
		overrides:      overrides,
		processUpdates: processUpdates,
		kubeClient:     kubeClient,
	}
}

func (i *core) logStatuses(statusMap map[string]string) {
	i.cfg.Log.Infof("Components processed so far:")
	for k, v := range statusMap {
		i.cfg.Log.Infof("Component: %s, Status: %s", k, v)
	}
}

func (i *core) getConfig() (overrides.Provider, *engine.Engine, *engine.Engine, error) {
	o, err := i.overrides.Build()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Failed to create overrides provider: exiting")
	}

	overridesProvider, err := overrides.New(i.kubeClient, o.Map(), i.cfg.Log)

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

func isK3dCluster(kubeClient kubernetes.Interface) (isK3d bool, err error) {

	retryOptions := []retry.Option{
		retry.Delay(2 * time.Second),
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
	}

	err = retry.Do(func() error {
		nodeList, err := kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, node := range nodeList.Items {
			if strings.HasPrefix(node.GetName(), "k3d-") {
				isK3d = true
				return nil
			}
		}

		return nil
	}, retryOptions...)
	if err != nil {
		return isK3d, err
	}

	return isK3d, nil
}

func getK3dClusterName(kubeClient kubernetes.Interface) (k3dName string, err error) {
	retryOptions := []retry.Option{
		retry.Delay(2 * time.Second),
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
	}

	err = retry.Do(func() error {
		labelSelector := metav1.LabelSelector{
			MatchLabels: map[string]string{"node-role.kubernetes.io/master": "true"},
		}
		listOptions := metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}
		nodeList, err := kubeClient.CoreV1().Nodes().List(context.Background(), listOptions)
		if err != nil {
			return err
		}

		for _, node := range nodeList.Items {
			nodeName := node.GetName()
			if !strings.HasPrefix(nodeName, "k3d-") {
				k3dName = ""
				return errors.New("Cluster is not a k3d cluster")
			}
			// K3d cluster name can be derived from master node names, which has the form k3d-<cluster-name>-server-<id>.
			// E.g., with the Kyma CLI default flags k3d-kyma-server-0
			k3dName = strings.TrimSuffix(strings.TrimRightFunc(strings.TrimPrefix(nodeName, "k3d-"), func(r rune) bool {
				return unicode.IsNumber(r) || r == '-'
			}), "-server")
		}

		return nil
	}, retryOptions...)
	if err != nil {
		return k3dName, err
	}

	return k3dName, nil
}

func registerOverridesInterceptors(ob *OverridesBuilder, kubeClient kubernetes.Interface, log logger.Interface) {
	//hide certificate data
	ob.AddInterceptor([]string{"global.domainName", "global.ingress.domainName"}, NewDomainNameOverrideInterceptor(kubeClient, log))
	ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient))
	// make sure we don't install legacy CRDs
	ob.AddInterceptor([]string{"global.installCRDs"}, NewInstallLegacyCRDsInterceptor())

	// make sure we don't install kcproxy for kiali and tracing
	ob.AddInterceptor([]string{"tracing.kcproxy.enabled", "kiali.kcproxy.enabled"}, NewDisableKCProxyInterceptor())

	// make sure k3d clusters use k3d container registry
	ob.AddInterceptor([]string{"serverless.dockerRegistry.internalServerAddress", "serverless.dockerRegistry.serverAddress", "serverless.dockerRegistry.registryAddress"}, NewRegistryInterceptor(kubeClient))

	// make sure k3d clusters disable internal container registry
	ob.AddInterceptor([]string{"serverless.dockerRegistry.enableInternal"}, NewRegistryDisableInterceptor(kubeClient))
}
