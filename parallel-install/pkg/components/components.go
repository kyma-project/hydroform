package components

import (
	"path"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
)

//Provider is an entity that produces a list of components for Kyma installation or uninstallation.
type Provider interface {
	GetComponents() []KymaComponent
}

//ComponentsProvider implements the Provider interface.
type ComponentsProvider struct {
	overridesProvider overrides.OverridesProvider
	resourcesPath     string //A root directory where subdirectories of components' charts are located.
	components        []config.ComponentDefinition
	helmConfig        helm.Config
	log               logger.Interface
	profile           string
}

//NewComponentsProvider returns a ComponentsProvider instance.
func NewComponentsProvider(overridesProvider overrides.OverridesProvider, cfg *config.Config, components []config.ComponentDefinition, kymaMetadata *helm.KymaMetadata) *ComponentsProvider {

	helmCfg := helm.Config{
		HelmTimeoutSeconds:            cfg.HelmTimeoutSeconds,
		BackoffInitialIntervalSeconds: cfg.BackoffInitialIntervalSeconds,
		BackoffMaxElapsedTimeSeconds:  cfg.BackoffMaxElapsedTimeSeconds,
		Log:                           cfg.Log,
		MaxHistory:                    cfg.HelmMaxRevisionHistory,
		Atomic:                        cfg.Atomic,
		KymaMetadata:                  kymaMetadata,
	}

	return &ComponentsProvider{
		overridesProvider: overridesProvider,
		resourcesPath:     cfg.ResourcePath,
		components:        components,
		helmConfig:        helmCfg,
		log:               cfg.Log,
		profile:           cfg.Profile,
	}
}

//Implements Provider.GetComponents.
func (p *ComponentsProvider) GetComponents() []KymaComponent {
	helmClient := helm.NewClient(p.helmConfig)

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

	return components
}
