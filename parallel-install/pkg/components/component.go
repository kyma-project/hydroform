package components

import (
	"log"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
)

const StatusError = "Error"
const StatusInstalled = "Installed"
const StatusUninstalled = "Uninstalled"

type Component struct {
	Name       string
	Namespace  string
	Status     string
	ChartDir   string
	Overrides  map[string]interface{}
	HelmClient helm.ClientInterface
}

func NewComponent(name, namespace, chartDir string, overrides map[string]interface{}, helmClient helm.ClientInterface) *Component {
	return &Component{
		Name:       name,
		Namespace:  namespace,
		ChartDir:   chartDir,
		Overrides:  overrides,
		HelmClient: helmClient,
		Status:     "NotStarted",
	}
}

type ComponentInstallation interface {
	InstallComponent() error
	UnInstallComponent() error
}

func (c *Component) InstallComponent() error {
	log.Printf("Installing %s in %s from %s", c.Name, c.Namespace, c.ChartDir)

	err := c.HelmClient.InstallRelease(c.ChartDir, c.Namespace, c.Name, c.Overrides)
	if err != nil {
		log.Printf("Error installing %s: %v", c.Name, err)
		return err
	}

	log.Printf("Installed %s in %s", c.Name, c.Namespace)

	return nil
}

func (c *Component) UninstallComponent() error {
	log.Printf("Uninstalling %s in %s from %s", c.Name, c.Namespace, c.ChartDir)

	err := c.HelmClient.UninstallRelease(c.Namespace, c.Name)
	if err != nil {
		log.Printf("Error uninstalling %s: %v", c.Name, err)
		return err
	}

	log.Printf("Uninstalled %s in %s", c.Name, c.Namespace)

	return nil
}
