package components

import (
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/helm"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
	"path"
)

type PrerequisitesProvider struct {
	overridesProvider overrides.OverridesProvider
	path              string
	componentList     [][]string
}

func NewPrerequisitesProvider(overridesProvider overrides.OverridesProvider, path string, componentList [][]string) *PrerequisitesProvider {
	return &PrerequisitesProvider{
		overridesProvider: overridesProvider,
		path:              path,
		componentList:     componentList,
	}
}

func (p *PrerequisitesProvider) GetComponents() ([]Component, error) {
	helmClient := &helm.Client{}

	err := p.overridesProvider.ReadOverridesFromCluster()
	if err != nil {
		return nil, err
	}

	var components []Component
	for _, componentNamespacePair := range p.componentList {
		name := componentNamespacePair[0]
		namespace := componentNamespacePair[1]

		components = append(components, Component{
			Name:       name,
			Namespace:  namespace,
			ChartDir:   path.Join(p.path, name),
			Overrides:  p.overridesProvider.OverridesFor(name),
			HelmClient: helmClient,
		})
	}

	return components, nil
}
