package components

import (
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/helm"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
	"path"
)

type PrerequisitesProvider struct {
	overridesProvider overrides.OverridesProvider
	path              string
	componentListYaml string
}

func NewPrerequisitesProvider(overridesProvider overrides.OverridesProvider, path string, componentListYaml string) *PrerequisitesProvider {
	return &PrerequisitesProvider{
		overridesProvider: overridesProvider,
		path:              path,
		componentListYaml: componentListYaml,
	}
}


func (p *PrerequisitesProvider) GetComponents() ([]Component, error) {
	helmClient := &helm.Client{}

	err := p.overridesProvider.ReadOverridesFromCluster()
	if err != nil {
		return nil, err
	}

	//TODO: fetch prerequisites data.
	//var installationCR v1alpha1.Installation
	//err = yaml.Unmarshal([]byte(p.componentListYaml), &installationCR)
	//if err != nil {
	//	return nil, err
	//}
	//
	//if len(installationCR.Spec.Components) < 1 {
	//	return nil, fmt.Errorf("Could not find any components to install on Installation CR")
	//}
	//
	//var components []Component
	//for _, component := range installationCR.Spec.Components {
	//	component := Component{
	//		Name:       component.Name,
	//		Namespace:  component.Namespace,
	//		Overrides:  p.overridesProvider.OverridesFor(component.Name),
	//		ChartDir:   path.Join(p.path, component.Name),
	//		HelmClient: helmClient,
	//	}
	//	components = append(components, component)
	//}
	components := []Component{
		{
			Name:       "cluster-essentials",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("cluster-essentials"),
			ChartDir:   path.Join(p.path, "cluster-essentials"),
			HelmClient: helmClient,
		},
		{
			Name:       "istio",
			Namespace:  "istio-system",
			Overrides:  p.overridesProvider.OverridesFor("istio"),
			ChartDir:   path.Join(p.path, "istio"),
			HelmClient: helmClient,
		},
		{
			Name:       "xip-patch",
			Namespace:  "kyma-installer",
			Overrides:  p.overridesProvider.OverridesFor("xip-patch"),
			ChartDir:   path.Join(p.path, "xip-patch"),
			HelmClient: helmClient,
		},
	}

	return components, nil
}

