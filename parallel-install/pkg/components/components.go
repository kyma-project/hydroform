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

//Provider is an entity that produces a list of components for installation or uninstallation of Kyma.
type Provider interface {
	GetComponents() ([]Component, error)
}

//ComponentsProvider implements Provider interface
type ComponentsProvider struct {
	overridesProvider overrides.OverridesProvider
	rootDir           string //a root directory where components' charts subdirectories are located.
	componentListYaml string
	helmConfig        helm.Config
	log               func(format string, v ...interface{})
}

//NewComponentsProvider returns a ComponentsProvider instance.
//
//rootDir is a directory where components' charts subdirectories are located.
//
//componentListYaml is a string containing YAML with an Installation CR
func NewComponentsProvider(overridesProvider overrides.OverridesProvider, rootDir string, componentListYaml string, cfg config.Config) *ComponentsProvider {

	helmCfg := helm.Config{
		HelmTimeoutSeconds:            cfg.HelmTimeoutSeconds,
		BackoffInitialIntervalSeconds: cfg.BackoffInitialIntervalSeconds,
		BackoffMaxElapsedTimeSeconds:  cfg.BackoffMaxElapsedTimeSeconds,
		Log:                           cfg.Log,
	}

	return &ComponentsProvider{
		overridesProvider: overridesProvider,
		rootDir:           rootDir,
		componentListYaml: componentListYaml,
		helmConfig:        helmCfg,
		log:               cfg.Log,
	}
}

//Implements Provider.GetComponents.
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
			ChartDir:        path.Join(p.rootDir, component.Name),
			HelmClient:      helmClient,
			Log:             p.log,
		}
		components = append(components, component)
	}

	return components, nil
}
