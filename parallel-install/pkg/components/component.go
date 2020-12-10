package components

import (
	"context"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
)

const StatusError = "Error"
const StatusInstalled = "Installed"
const StatusUninstalled = "Uninstalled"

const logPrefix = "[components/component.go]"

//ComponentDeployment interface defines a contract for Component installation and uninstallation.
type ComponentDeploy interface {
	//DeployComponent installs a component.
	//The function is blocking until the component is installed or an error (including Helm timeout) occurs.
	//See the helm.HelmClient.DeployRelease documentation for how context.Context is used for cancellation.
	DeployComponent(context.Context) error

	//UninstallComponent uninstalls a component.
	//The function is blocking until the component is uninstalled or an error (including Helm timeout) occurs.
	//See the helm.HelmClient.UninstallRelease documentation for how context.Context is used for cancellation.
	UninstallComponent(context.Context) error
}

//Component implements the ComponentDeployment interface.
type Component struct {
	Name            string
	Namespace       string
	Status          string
	ChartDir        string
	OverridesGetter func() map[string]interface{}
	HelmClient      helm.ClientInterface
	Log             func(format string, v ...interface{})
}

//NewComponent instantiates a new Component.
//"name" and "namespace" parameters define the Helm release name and namespace.
//
//"chartDir" is a local filesystem directory with the component's chart.
//
//"overrides" is a function that returns overrides for the release.
func NewComponent(name, namespace, chartDir string, overrides func() map[string]interface{}, helmClient helm.ClientInterface, log func(string, ...interface{})) *Component {
	return &Component{
		Name:            name,
		Namespace:       namespace,
		ChartDir:        chartDir,
		OverridesGetter: overrides,
		HelmClient:      helmClient,
		Status:          "NotStarted",
		Log:             log,
	}
}

//DeployComponent implements ComponentDeployment.DeployComponent.
func (c *Component) DeployComponent(ctx context.Context) error {
	c.Log("%s Installing %s in %s from %s", logPrefix, c.Name, c.Namespace, c.ChartDir)

	overrides := c.OverridesGetter()

	err := c.HelmClient.DeployRelease(ctx, c.ChartDir, c.Namespace, c.Name, overrides)
	if err != nil {
		c.Log("%s Error installing %s: %v", logPrefix, c.Name, err)
		return err
	}

	c.Log("%s Installed %s in %s", logPrefix, c.Name, c.Namespace)

	return nil
}

//UninstallComponent implements ComponentInstallation.UninstallComponent.
func (c *Component) UninstallComponent(ctx context.Context) error {
	c.Log("%s Uninstalling %s in %s from %s", logPrefix, c.Name, c.Namespace, c.ChartDir)

	err := c.HelmClient.UninstallRelease(ctx, c.Namespace, c.Name)
	if err != nil {
		c.Log("%s Error uninstalling %s: %v", logPrefix, c.Name, err)
		return err
	}

	c.Log("%s Uninstalled %s in %s", logPrefix, c.Name, c.Namespace)

	return nil
}
