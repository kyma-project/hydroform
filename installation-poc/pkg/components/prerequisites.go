package components

import (
	"path"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/config"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/helm"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
)

type PrerequisitesProvider struct {
	overridesProvider overrides.OverridesProvider
	path              string
	componentList     [][]string
	helmConfig        helm.Config
}

func NewPrerequisitesProvider(overridesProvider overrides.OverridesProvider, path string, componentList [][]string, cfg config.Config) *PrerequisitesProvider {
	helmCfg := helm.Config{
		HelmTimeoutSeconds:            cfg.HelmTimeoutSeconds,
		BackoffInitialIntervalSeconds: cfg.BackoffInitialIntervalSeconds,
		BackoffMaxElapsedTimeSeconds:  cfg.BackoffMaxElapsedTimeSeconds,
	}

	return &PrerequisitesProvider{
		overridesProvider: overridesProvider,
		path:              path,
		componentList:     componentList,
		helmConfig:        helmCfg,
	}
}

func (p *PrerequisitesProvider) GetComponents() ([]Component, error) {
	helmClient := helm.NewClient(p.helmConfig)

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
