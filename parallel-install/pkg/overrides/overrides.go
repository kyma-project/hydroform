//Package overrides implements the logic related to handling overrides.
//The manually-provided overrides have precedence over standard Kyma overrides defined in the cluster.
//
//The code in the package uses the user-provided function for logging.
package overrides

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"helm.sh/helm/v3/pkg/strvals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const logPrefix = "[overrides/overrides.go]"

var commonListOpts = metav1.ListOptions{LabelSelector: "installer=overrides, !component"}
var componentListOpts = metav1.ListOptions{LabelSelector: "installer=overrides, component"}

//Provider type caches overrides for further use.
//It contains overrides from the cluster and the manually-provided ones.
type Provider struct {
	overrides                    map[string]interface{}
	additionalOverrides          map[string]interface{}
	componentOverrides           map[string]map[string]interface{}
	additionalComponentOverrides map[string]map[string]interface{}
	kubeClient                   kubernetes.Interface
	log                          logger.Interface
}

//OverridesProvider defines the contract for reading overrides for a given Helm release.
type OverridesProvider interface {
	//OverridesGetterFunctionFor returns a function returning overrides for a Helm release with the provided name.
	//Before using this function, ensure that the overrides cache is populated by calling the ReadOverridesFromCluster function.
	OverridesGetterFunctionFor(name string) func() map[string]interface{}

	//Populates overrides cache by reading data from the cluster. You have to call this function before using OverridesGetterFunctionFor.
	ReadOverridesFromCluster() error
}

//New returns a new Provider.
//
//overridesYaml contains a list of manually-provided overrides.
//Every value in the list contains data in the YAML format.
//The structure of the file should follow the Helm's values.yaml convention.
//There is one difference from the plain Helm's values.yaml file: These are not values for a single release but for the entire Kyma installation.
//Because of that, you have to put values for a specific Component (e.g: Component name is "foo") under a key equal to the component's name (i.e: "foo").
//You can also put overrides under a "global" key. These will merge with the top-level "global" Helm key for every Helm chart.
func New(client kubernetes.Interface, overrides map[string]interface{}, log logger.Interface) (OverridesProvider, error) {
	provider := Provider{
		kubeClient: client,
		log:        log,
	}

	err := provider.parseAdditionalOverrides(overrides)
	if err != nil {
		return nil, err
	}

	return &provider, nil
}

func (p *Provider) OverridesGetterFunctionFor(name string) func() map[string]interface{} {
	return func() map[string]interface{} {
		if val, ok := p.componentOverrides[name]; ok {
			val = MergeMaps(val, p.overrides)
			return val
		}
		return p.overrides
	}
}

func (p *Provider) ReadOverridesFromCluster() error {

	// TODO: add retries
	//Read global overrides
	globalOverrideCMs, err := p.kubeClient.CoreV1().ConfigMaps("kyma-installer").List(context.TODO(), commonListOpts)
	if err != nil {
		return err
	}

	var globalValues []string
	for _, cm := range globalOverrideCMs.Items {
		for k, v := range cm.Data {
			globalValues = append(globalValues, k+"="+v)
		}
	}

	if p.overrides == nil {
		p.overrides = make(map[string]interface{})
	}

	globalFromCluster := make(map[string]interface{})

	for _, value := range globalValues {
		if err := strvals.ParseInto(value, globalFromCluster); err != nil {
			p.log.Errorf("%s Error parsing global overrides: %v", logPrefix, err)
			return err
		}
	}

	p.overrides = MergeMaps(p.overrides, globalFromCluster)
	p.overrides = MergeMaps(p.overrides, p.additionalOverrides) // always keep additionalOverrides on top

	//Read component overrides
	if p.componentOverrides == nil {
		p.componentOverrides = make(map[string]map[string]interface{})
	}

	componentOverrideCMs, err := p.kubeClient.CoreV1().ConfigMaps("kyma-installer").List(context.TODO(), componentListOpts)

	for _, cm := range componentOverrideCMs.Items {
		var componentValues []string
		name := cm.Labels["component"]

		for k, v := range cm.Data {
			componentValues = append(componentValues, k+"="+v)
		}

		if p.componentOverrides[name] == nil {
			p.componentOverrides[name] = make(map[string]interface{})
		}

		componentsFromCluster := make(map[string]interface{})

		for _, value := range componentValues {
			if err := strvals.ParseInto(value, componentsFromCluster); err != nil {
				p.log.Infof("%s Error parsing overrides for %s: %v", logPrefix, name, err)
				return err
			}
		}

		p.componentOverrides[name] = MergeMaps(p.componentOverrides[name], componentsFromCluster)
		p.componentOverrides[name] = MergeMaps(p.componentOverrides[name], p.additionalComponentOverrides[name]) // always keep additionalOverrides on top
	}

	p.log.Infof("%s Reading the overrides from the cluster completed successfully!", logPrefix)
	return nil
}

func (p *Provider) parseAdditionalOverrides(additionalOverrides map[string]interface{}) error {

	if p.additionalComponentOverrides == nil {
		p.additionalComponentOverrides = make(map[string]map[string]interface{})
	}

	if p.componentOverrides == nil {
		p.componentOverrides = make(map[string]map[string]interface{})
	}

	for k, v := range additionalOverrides {
		if k == "global" {
			globalOverrides := make(map[string]interface{})
			globalOverrides[k] = v
			p.overrides = MergeMaps(p.overrides, globalOverrides)
			p.additionalOverrides = MergeMaps(p.additionalOverrides, globalOverrides)
		} else {
			if p.additionalComponentOverrides[k] == nil {
				p.additionalComponentOverrides[k] = make(map[string]interface{})
			}
			if p.componentOverrides[k] == nil {
				p.componentOverrides[k] = make(map[string]interface{})
			}

			if vTypeSafe, ok := v.(map[string]interface{}); ok {
				p.componentOverrides[k] = MergeMaps(p.componentOverrides[k], vTypeSafe)
				p.additionalComponentOverrides[k] = MergeMaps(p.additionalComponentOverrides[k], vTypeSafe)
			} else {
				return fmt.Errorf("Cannot add override '%s=%s' as the value has to be a map", k, v)
			}
		}
	}

	return nil
}

func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
