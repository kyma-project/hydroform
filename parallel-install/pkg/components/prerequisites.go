package components

import (
	"path"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
)

//PrerequisitesProvider implements a contract for getting Components that are considered prerequisites for Kyma installation.
//It implements the Provider interface.
type PrerequisitesProvider struct {
	overridesProvider overrides.OverridesProvider
	resourcesPath     string     ////A root directory where subdirectories of components' charts are located.
	componentList     [][]string // TODO: replace with []struct{name, namespace string}
	helmConfig        helm.Config
	log               func(format string, v ...interface{})
}

//NewPrerequisitesProvider returns a new PrerequisitesProvider instance.
//
//resourcesPath is a directory where subdirectories of components' charts are located.
//
//componentList is a slice of pairs: [component-name, namespace]
func NewPrerequisitesProvider(overridesProvider overrides.OverridesProvider, resourcesPath string, componentList [][]string, cfg config.Config) *PrerequisitesProvider {
	helmCfg := helm.Config{
		HelmTimeoutSeconds:            cfg.HelmTimeoutSeconds,
		BackoffInitialIntervalSeconds: cfg.BackoffInitialIntervalSeconds,
		BackoffMaxElapsedTimeSeconds:  cfg.BackoffMaxElapsedTimeSeconds,
		Log:                           cfg.Log,
	}

	return &PrerequisitesProvider{
		overridesProvider: overridesProvider,
		resourcesPath:     resourcesPath,
		componentList:     componentList,
		helmConfig:        helmCfg,
		log:               cfg.Log,
	}
}

//Implements Provider.GetComponents.
func (p *PrerequisitesProvider) GetComponents() ([]Component, error) {
	helmClient := helm.NewClient(p.helmConfig)

	var components []Component
	for _, componentNamespacePair := range p.componentList {
		name := componentNamespacePair[0]
		namespace := componentNamespacePair[1]

		components = append(components, Component{
			Name:            name,
			Namespace:       namespace,
			ChartDir:        path.Join(p.resourcesPath, name),
			OverridesGetter: p.overridesProvider.OverridesGetterFunctionFor(name),
			HelmClient:      helmClient,
			Log:             p.log,
		})
	}

	return components, nil
}
