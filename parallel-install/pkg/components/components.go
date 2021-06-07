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
	// GetComponents returns the component list in the provider either in natural or reversed order
	GetComponents(reversed bool) []KymaComponent
}

//ComponentsProvider implements the Provider interface.
type ComponentsProvider struct {
	overridesProvider overrides.Provider
	resourcesPath     string //A root directory where subdirectories of components' charts are located.
	components        []config.ComponentDefinition
	helmConfig        helm.Config
	log               logger.Interface
	profile           string
}

//NewComponentsProvider returns a ComponentsProvider instance.
func NewComponentsProvider(overridesProvider overrides.Provider, cfg *config.Config, components []config.ComponentDefinition, tpl *helm.KymaComponentMetadataTemplate) *ComponentsProvider {
	helmCfg := helm.Config{
		HelmTimeoutSeconds:            cfg.HelmTimeoutSeconds,
		BackoffInitialIntervalSeconds: cfg.BackoffInitialIntervalSeconds,
		BackoffMaxElapsedTimeSeconds:  cfg.BackoffMaxElapsedTimeSeconds,
		Log:                           cfg.Log,
		MaxHistory:                    cfg.HelmMaxRevisionHistory,
		Atomic:                        cfg.Atomic,
		KymaComponentMetadataTemplate: tpl,
		KubeconfigSource:              cfg.KubeconfigSource,
		ReuseValues:                   cfg.ReuseHelmValues,
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
func (p *ComponentsProvider) GetComponents(reversed bool) []KymaComponent {
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
		if reversed {
			// prepend component to reverse the list
			components = append([]KymaComponent{cmp}, components...)
		} else {
			components = append(components, cmp)
		}
	}

	return components
}
