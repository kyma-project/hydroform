package overrides

import (
	"log"
	"context"
	"helm.sh/helm/v3/pkg/strvals"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var commonListOpts = metav1.ListOptions{LabelSelector: "installer=overrides, !component"}
var componentListOpts = metav1.ListOptions{LabelSelector: "installer=overrides, component"}

type Provider struct {
	overrides          map[string]interface{}
	componentOverrides map[string]map[string]interface{}
	kubeClient         kubernetes.Interface
}

type OverridesProvider interface {
	OverridesFor(name string) map[string]interface{}
	ReadOverridesFromCluster() error
}

func New(client kubernetes.Interface) OverridesProvider {
	res := Provider{
		kubeClient:client,
	}

	return &res
}

func (p *Provider) OverridesFor(name string) map[string]interface{} {
	if val, ok := p.componentOverrides[name]; ok {
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

	p.overrides = make(map[string]interface{})
	for _, value := range globalValues {
		if err := strvals.ParseInto(value, p.overrides); err != nil {
			log.Printf("Error parsing global overrides: %v", err)
			return err
		}
	}

	//Read component overrides
	p.componentOverrides = make(map[string]map[string]interface{})

	componentOverrideCMs, err := p.kubeClient.CoreV1().ConfigMaps("kyma-installer").List(context.TODO(), componentListOpts)

	for _, cm := range componentOverrideCMs.Items {
		log.Printf("%s data %v", cm.Name, cm.Data)
		var componentValues []string
		name := cm.Labels["component"]

		for k, v := range cm.Data {
			componentValues = append(componentValues, k+"="+v)
		}

		//Merge global overrides to component overrides for each component
		componentValues = append(globalValues, componentValues...)

		p.componentOverrides[name] = make(map[string]interface{})
		for _, value := range componentValues {
			if err := strvals.ParseInto(value, p.componentOverrides[name]); err != nil {
				log.Printf("Error parsing overrides for %s: %v", name, err)
				return err
			}
		}
	}

	log.Println("Reading the overrides from the cluster completed successfully!")
	return nil
}
