package components

import (
	"fmt"
	"path"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"gopkg.in/yaml.v3"
)

type Provider interface {
	GetComponents() ([]Component, error)
}

type ComponentsProvider struct {
	overridesProvider overrides.OverridesProvider
	path              string
	componentListYaml string
	helmConfig        helm.Config
	log               func(format string, v ...interface{})
}

func NewComponentsProvider(overridesProvider overrides.OverridesProvider, path string, componentListYaml string, cfg config.Config) *ComponentsProvider {

	helmCfg := helm.Config{
		HelmTimeoutSeconds:            cfg.HelmTimeoutSeconds,
		BackoffInitialIntervalSeconds: cfg.BackoffInitialIntervalSeconds,
		BackoffMaxElapsedTimeSeconds:  cfg.BackoffMaxElapsedTimeSeconds,
		Log:                           cfg.Log,
	}

	return &ComponentsProvider{
		overridesProvider: overridesProvider,
		path:              path,
		componentListYaml: componentListYaml,
		helmConfig:        helmCfg,
		log:               cfg.Log,
	}
}

func (p *ComponentsProvider) GetComponents() ([]Component, error) {
	helmClient := helm.NewClient(p.helmConfig)

	var installationCR v1alpha1.Installation
	err := yaml.Unmarshal([]byte(p.componentListYaml), &installationCR)
	if err != nil {
		return nil, err
	}

	if len(installationCR.Spec.Components) < 1 {
		return nil, fmt.Errorf("Could not find any components to install on Installation CR")
	}

	var components []Component
	for _, component := range installationCR.Spec.Components {
		component := Component{
			Name:            component.Name,
			Namespace:       component.Namespace,
			OverridesGetter: p.overridesProvider.OverridesGetterFunctionFor(component.Name),
			ChartDir:        path.Join(p.path, component.Name),
			HelmClient:      helmClient,
			Log:             p.log,
		}
		components = append(components, component)
	}

	return components, nil
}
