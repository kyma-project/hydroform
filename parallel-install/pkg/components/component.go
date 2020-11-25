package components

import (
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
)

const StatusError = "Error"
const StatusInstalled = "Installed"
const StatusUninstalled = "Uninstalled"

type Component struct {
	Name            string
	Namespace       string
	Status          string
	ChartDir        string
	OverridesGetter func() map[string]interface{}
	HelmClient      helm.ClientInterface
	Log             func(format string, v ...interface{})
}

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

type ComponentInstallation interface {
	InstallComponent() error
	UnInstallComponent() error
}

func (c *Component) InstallComponent() error {
	c.Log("Installing %s in %s from %s", c.Name, c.Namespace, c.ChartDir)

	overrides := c.OverridesGetter()

	err := c.HelmClient.InstallRelease(c.ChartDir, c.Namespace, c.Name, overrides)
	if err != nil {
		c.Log("Error installing %s: %v", c.Name, err)
		return err
	}

	c.Log("Installed %s in %s", c.Name, c.Namespace)

	return nil
}

func (c *Component) UninstallComponent() error {
	c.Log("Uninstalling %s in %s from %s", c.Name, c.Namespace, c.ChartDir)

	err := c.HelmClient.UninstallRelease(c.Namespace, c.Name)
	if err != nil {
		c.Log("Error uninstalling %s: %v", c.Name, err)
		return err
	}

	c.Log("Uninstalled %s in %s", c.Name, c.Namespace)

	return nil
}
