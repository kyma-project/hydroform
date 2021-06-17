//Package deployment provides a top-level API to control Kyma deployment and uninstallation.
package deployment

import (
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/engine"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	"k8s.io/client-go/kubernetes"
)

//Templating renders Kyma charts
type Templating struct {
	*core
}

//NewTemplating creates a new Templating instance for rendering Kyma charts.
func NewTemplating(cfg *config.Config, ob *overrides.Builder) (*Templating, error) {
	if err := cfg.ValidateDeployment(); err != nil { //Templating requires same configuration as Deployment
		return nil, err
	}

	restConfig, err := config.RestConfig(cfg.KubeconfigSource)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	registerOverridesInterceptors(ob, kubeClient, cfg.Log)

	core := newCore(cfg, ob, kubeClient, nil)

	return &Templating{core}, nil
}

//Render renders the Kyma component templates
func (d *Templating) Render() ([]*components.Manifest, error) {
	//Prepare cluster before Kyma installation
	preInstallerCfg := inputConfig{
		InstallationResourcePath: d.cfg.InstallationResourcePath,
		Log:                      d.cfg.Log,
		KubeconfigSource:         d.cfg.KubeconfigSource,
	}

	preInstaller, err := newPreInstaller(preInstallerCfg)
	if err != nil {
		d.cfg.Log.Fatalf("Failed to create Kyma pre-installer: %v", err)
	}

	result, err := preInstaller.Manifests()
	if err != nil {
		return nil, err
	}

	_, prerequisitesEng, componentsEng, err := d.getConfig()
	if err != nil {
		return result, err
	}

	for _, eng := range []*engine.Engine{prerequisitesEng, componentsEng} {
		manifests, err := eng.Manifests()
		if err != nil {
			return nil, err
		}
		result = append(result, manifests...)
	}

	return result, nil
}
