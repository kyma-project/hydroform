package components

import (
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/helm"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
	"path"
)

type PrerequisitesProvider struct {
	overridesProvider overrides.OverridesProvider
	path              string
	componentList 	map[string]string
}

func NewPrerequisitesProvider(overridesProvider overrides.OverridesProvider, path string, componentList map[string]string) *PrerequisitesProvider {
	return &PrerequisitesProvider{
		overridesProvider: overridesProvider,
		path:              path,
		componentList: 	   componentList,
	}
}


func (p *PrerequisitesProvider) GetComponents() ([]Component, error) {
	helmClient := &helm.Client{}

	err := p.overridesProvider.ReadOverridesFromCluster()
	if err != nil {
		return nil, err
	}

	var components []Component
	for component, namespace := range p.componentList {
		components = append(components, Component{
			Name:       component,
			Namespace:  namespace,
			ChartDir:   path.Join(p.path, component),
			Overrides:  p.overridesProvider.OverridesFor(component),
			HelmClient: helmClient,
		})
	}

	return components, nil
}

