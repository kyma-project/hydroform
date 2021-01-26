package components

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"go.uber.org/zap"
	"path"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
)

//Provider is an entity that produces a list of components for Kyma installation or uninstallation.
type Provider interface {
	GetComponents() ([]KymaComponent, error)
}

//ComponentsProvider implements the Provider interface.
type ComponentsProvider struct {
	overridesProvider overrides.OverridesProvider
	resourcesPath     string //A root directory where subdirectories of components' charts are located.
	components        []ComponentDefinition
	helmConfig        helm.Config
	log               *zap.SugaredLogger
	profile           string
}

//NewComponentsProvider returns a ComponentsProvider instance.
//
//resourcesPath is a directory where subdirectories of components' charts are located.
//
//componentListYaml is a string containing YAML with an Installation CR.
func NewComponentsProvider(overridesProvider overrides.OverridesProvider, resourcesPath string, components []ComponentDefinition, cfg config.Config) *ComponentsProvider {

	helmCfg := helm.Config{
		HelmTimeoutSeconds:            cfg.HelmTimeoutSeconds,
		BackoffInitialIntervalSeconds: cfg.BackoffInitialIntervalSeconds,
		BackoffMaxElapsedTimeSeconds:  cfg.BackoffMaxElapsedTimeSeconds,
		Log:                           logger.NewLogger(cfg.Verbose),
		MaxHistory:                    cfg.HelmMaxRevisionHistory,
	}

	return &ComponentsProvider{
		overridesProvider: overridesProvider,
		resourcesPath:     resourcesPath,
		components:        components,
		helmConfig:        helmCfg,
		log:               logger.NewLogger(cfg.Verbose),
		profile:           cfg.Profile,
	}
}

//Implements Provider.GetComponents.
func (p *ComponentsProvider) GetComponents() ([]KymaComponent, error) {
	helmClient := helm.NewClient(p.helmConfig)

	if len(p.components) < 1 {
		return nil, fmt.Errorf("Could not find any components to install on Installation CR")
	}

	var components []KymaComponent
	for _, component := range p.components {
		cmp := KymaComponent{
			Name:            component.Name,
			Namespace:       component.Namespace,
			Profile:         p.profile,
			OverridesGetter: p.overridesProvider.OverridesGetterFunctionFor(component.Name),
			ChartDir:        path.Join(p.resourcesPath, component.Name),
			HelmClient:      helmClient,
			Log:             p.log,
		}
		components = append(components, cmp)
	}

	return components, nil
}
