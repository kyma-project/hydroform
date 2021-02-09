package components

import (
	"context"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
)

const StatusError = "Error"
const StatusInstalled = "Installed"
const StatusUninstalled = "Uninstalled"

const logPrefix = "[components/component.go]"

//Component interface defines a contract for Component deployment and uninstallation.
type Component interface {
	//Deploy installs a component.
	//The function is blocking until the component is installed or an error (including Helm timeout) occurs.
	//See the helm.HelmClient.DeployRelease documentation for how context.Context is used for cancellation.
	Deploy(context.Context) error

	//Uninstall uninstalls a component.
	//The function is blocking until the component is uninstalled or an error (including Helm timeout) occurs.
	//See the helm.HelmClient.UninstallRelease documentation for how context.Context is used for cancellation.
	Uninstall(context.Context) error
}

//KymaComponent implements the Component interface.
type KymaComponent struct {
	//Name defines the Helm release name
	Name string
	//Namespace defines the Helm release namespace
	Namespace string
	//Profile defines the Kyma release namespace
	Profile string
	Status  string
	//ChartDir is a local filesystem directory with the component's chart.
	ChartDir string
	//OverridesGetter is a function that returns overrides for the release.
	OverridesGetter func() map[string]interface{}
	HelmClient      helm.ClientInterface
	Log             logger.Interface
}

//Deploy implements Component.Deploy
func (c *KymaComponent) Deploy(ctx context.Context) error {
	c.Log.Infof("%s Deploying %s in %s from %s", logPrefix, c.Name, c.Namespace, c.ChartDir)

	overrides := c.OverridesGetter()

	err := c.HelmClient.DeployRelease(ctx, c.ChartDir, c.Namespace, c.Name, overrides, c.Profile)
	if err != nil {
		c.Log.Errorf("%s Error deploying %s: %v", logPrefix, c.Name, err)
		return err
	}

	c.Log.Infof("%s Deployed %s in %s", logPrefix, c.Name, c.Namespace)

	return nil
}

//Uninstall implements Component.Uninstall.
func (c *KymaComponent) Uninstall(ctx context.Context) error {
	c.Log.Infof("%s Uninstalling %s in %s from %s", logPrefix, c.Name, c.Namespace, c.ChartDir)

	err := c.HelmClient.UninstallRelease(ctx, c.Namespace, c.Name)
	if err != nil {
		c.Log.Infof("%s Error uninstalling %s: %v", logPrefix, c.Name, err)
		return err
	}

	c.Log.Infof("%s Uninstalled %s in %s", logPrefix, c.Name, c.Namespace)

	return nil
}
