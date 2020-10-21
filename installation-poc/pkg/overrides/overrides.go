package overrides

import (
	"context"
	"log"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/strvals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var commonListOpts = metav1.ListOptions{LabelSelector: "installer=overrides, !component"}
var componentListOpts = metav1.ListOptions{LabelSelector: "installer=overrides, component"}

type Provider struct {
	overrides          map[string]interface{}
	componentOverrides map[string]map[string]interface{}
	kubeClient         kubernetes.Interface
	OverridesYaml      string
}

type OverridesProvider interface {
	OverridesFor(name string) map[string]interface{}
	ReadOverridesFromCluster() error
}

func New(client kubernetes.Interface, overridesYaml string) (OverridesProvider, error) {
	res := Provider{
		kubeClient:    client,
		OverridesYaml: overridesYaml,
	}

	if overridesYaml != "" {
		err := res.parseAdditionalOverrides()
		if err != nil {
			return nil, err
		}
	}

	return &res, nil
}

func (p *Provider) OverridesFor(name string) map[string]interface{} {
	if val, ok := p.componentOverrides[name]; ok {
		val = mergeMaps(val, p.overrides)
		log.Printf("Overrides for %s: %v", name, val)
		return val
	}
	log.Printf("Overrides for %s: %v", name, p.overrides)
	return p.overrides
}

func (p *Provider) ReadOverridesFromCluster() error {
	//Read global overrides
	globalOverrideCMs, err := p.kubeClient.CoreV1().ConfigMaps("kyma-installer").List(context.TODO(), commonListOpts)
	if err != nil {
		return err
	}

	var globalValues []string
	for _, cm := range globalOverrideCMs.Items {
		log.Printf("%s data %v", cm.Name, cm.Data)
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
			log.Printf("Error parsing global overrides: %v", err)
			return err
		}
	}

	p.overrides = mergeMaps(p.overrides, globalFromCluster)

	//Read component overrides
	if p.componentOverrides == nil {
		p.componentOverrides = make(map[string]map[string]interface{})
	}

	componentOverrideCMs, err := p.kubeClient.CoreV1().ConfigMaps("kyma-installer").List(context.TODO(), componentListOpts)

	for _, cm := range componentOverrideCMs.Items {
		log.Printf("%s data %v", cm.Name, cm.Data)
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
				log.Printf("Error parsing overrides for %s: %v", name, err)
				return err
			}
		}

		p.componentOverrides[name] = mergeMaps(p.componentOverrides[name], componentsFromCluster)
	}

	log.Println("Reading the overrides from the cluster completed successfully!")
	return nil
}

func (p *Provider) parseAdditionalOverrides() error {

	if p.componentOverrides == nil {
		p.componentOverrides = make(map[string]map[string]interface{})
	}

	var additionalOverrides map[string]interface{}
	err := yaml.Unmarshal([]byte(p.OverridesYaml), &additionalOverrides)
	if err != nil {
		return err
	}

	for k, v := range additionalOverrides {
		if k == "global" {
			globalOverrides := make(map[string]interface{})
			globalOverrides[k] = v
			p.overrides = mergeMaps(p.overrides, globalOverrides)
		} else {
			if p.componentOverrides[k] == nil {
				p.componentOverrides[k] = make(map[string]interface{})
			}

			p.componentOverrides[k] = mergeMaps(p.componentOverrides[k], v.(map[string]interface{}))
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
