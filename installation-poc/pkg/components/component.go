package components

import (
	"log"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/helm"
)

type Component struct {
	Name       string
	Namespace  string
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
	}
}

type ComponentInstallation interface {
	InstallComponent() error
}

func (c *Component) InstallComponent() error {
	log.Printf("MST Installing %s in %s from %s", c.Name, c.Namespace, c.ChartDir)

	err := c.HelmClient.InstallRelease(c.ChartDir, c.Namespace, c.Name, c.Overrides)
	if err != nil {
		log.Printf("MST Error installing %s: %v", c.Name, err)
		return err
	}

	log.Printf("MST Installed %s in %s", c.Name, c.Namespace)

	return nil
}
