package components

import (
	"fmt"
	"path"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/helm"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"gopkg.in/yaml.v3"
)

type Provider struct {
	overridesProvider overrides.OverridesProvider
	path              string
	componentListYaml string
}

func NewComponents(overridesProvider overrides.OverridesProvider, path string, componentListYaml string) *Provider {
	return &Provider{
		overridesProvider: overridesProvider,
		path:              path,
		componentListYaml: componentListYaml,
	}
}

type ComponentsProvider interface {
	GetComponents() ([]Component, error)
}

func (p *Provider) GetComponents() ([]Component, error) {
	helmClient := &helm.Client{}

	err := p.overridesProvider.ReadOverridesFromCluster()
	if err != nil {
		return nil, err
	}

	var installationCR v1alpha1.Installation
	err = yaml.Unmarshal([]byte(p.componentListYaml), &installationCR)
	if err != nil {
		return nil, err
	}

	if len(installationCR.Spec.Components) < 1 {
		return nil, fmt.Errorf("Could not find any components to install on Installation CR")
	}

	var components []Component
	for _, component := range installationCR.Spec.Components {
		component := Component{
			Name:       component.Name,
			Namespace:  component.Namespace,
			Overrides:  p.overridesProvider.OverridesFor(component.Name),
			ChartDir:   path.Join(p.path, component.Name),
			HelmClient: helmClient,
		}
		components = append(components, component)
	}

	return components, nil
}
