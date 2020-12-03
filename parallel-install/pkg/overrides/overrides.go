//Package overrides implements logic related to handling overrides.
//The manually-provided overrides have precedence over standard Kyma overrides defined in the cluster.
package overrides

import (
	"context"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/strvals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const logPrefix = "[overrides/overrides.go]"

var commonListOpts = metav1.ListOptions{LabelSelector: "installer=overrides, !component"}
var componentListOpts = metav1.ListOptions{LabelSelector: "installer=overrides, component"}

//Provider type caches overrides from the cluster and manually-provided ones.
type Provider struct {
	overrides                    map[string]interface{}
	additionalOverrides          map[string]interface{}
	componentOverrides           map[string]map[string]interface{}
	additionalComponentOverrides map[string]map[string]interface{}
	kubeClient                   kubernetes.Interface
	log                          func(format string, v ...interface{})
}

//OverridesProvider defines contract for reading overrides for given Helm release.
type OverridesProvider interface {
	//OverridesGetterFunctionFor returns overrides for Helm release with provided name.
	//Before using this function, ensure that local overrides cache is populated by calling ReadOverridesFromCluster.
	OverridesGetterFunctionFor(name string) func() map[string]interface{}
	//Reads overrides from the cluster. You have to call this function before using OverridesGetterFunctionFor().
	ReadOverridesFromCluster() error
}

func New(client kubernetes.Interface, overridesYamls []string, log func(string, ...interface{})) (OverridesProvider, error) {
	provider := Provider{
		kubeClient: client,
		log:        log,
	}

	for _, overridesYaml := range overridesYamls {
		if overridesYaml != "" {
			err := provider.parseAdditionalOverrides(overridesYaml)
			if err != nil {
				return nil, err
			}
		}
	}

	return &provider, nil
}

func (p *Provider) OverridesGetterFunctionFor(name string) func() map[string]interface{} {
	return func() map[string]interface{} {
		if val, ok := p.componentOverrides[name]; ok {
			val = mergeMaps(val, p.overrides)
			p.log("%s Overrides for %s: %v", logPrefix, name, val)
			return val
		}
		p.log("%s Overrides for %s: %v", logPrefix, name, p.overrides)
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
		p.log("%s %s data %v", logPrefix, cm.Name, cm.Data)
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
			p.log("%s Error parsing global overrides: %v", logPrefix, err)
			return err
		}
	}

	p.overrides = mergeMaps(p.overrides, globalFromCluster)
	p.overrides = mergeMaps(p.overrides, p.additionalOverrides) // always keep additionalOverrides on top

	//Read component overrides
	if p.componentOverrides == nil {
		p.componentOverrides = make(map[string]map[string]interface{})
	}

	componentOverrideCMs, err := p.kubeClient.CoreV1().ConfigMaps("kyma-installer").List(context.TODO(), componentListOpts)

	for _, cm := range componentOverrideCMs.Items {
		p.log("%s %s data %v", logPrefix, cm.Name, cm.Data)
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
				p.log("%s Error parsing overrides for %s: %v", logPrefix, name, err)
				return err
			}
		}

		p.componentOverrides[name] = mergeMaps(p.componentOverrides[name], componentsFromCluster)
		p.componentOverrides[name] = mergeMaps(p.componentOverrides[name], p.additionalComponentOverrides[name]) // always keep additionalOverrides on top
	}

	p.log("%s Reading the overrides from the cluster completed successfully!", logPrefix)
	return nil
}

func (p *Provider) parseAdditionalOverrides(overridesYaml string) error {

	if p.additionalComponentOverrides == nil {
		p.additionalComponentOverrides = make(map[string]map[string]interface{})
	}

	if p.componentOverrides == nil {
		p.componentOverrides = make(map[string]map[string]interface{})
	}

	var additionalOverrides map[string]interface{}
	err := yaml.Unmarshal([]byte(overridesYaml), &additionalOverrides)
	if err != nil {
		return err
	}

	for k, v := range additionalOverrides {
		if k == "global" {
			globalOverrides := make(map[string]interface{})
			globalOverrides[k] = v
			p.overrides = mergeMaps(p.overrides, globalOverrides)
			p.additionalOverrides = mergeMaps(p.additionalOverrides, globalOverrides)
		} else {
			if p.additionalComponentOverrides[k] == nil {
				p.additionalComponentOverrides[k] = make(map[string]interface{})
			}
			if p.componentOverrides[k] == nil {
				p.componentOverrides[k] = make(map[string]interface{})
			}
			p.componentOverrides[k] = mergeMaps(p.componentOverrides[k], v.(map[string]interface{}))
			p.additionalComponentOverrides[k] = mergeMaps(p.additionalComponentOverrides[k], v.(map[string]interface{}))
		}
	}

	return nil
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
