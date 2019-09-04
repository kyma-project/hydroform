package gcp

import (
	tcli "github.com/kyma-incubator/hydroform/api/terraform"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/terraform-providers/terraform-provider-google/google"
)

const clusterTemplate string = `

`

type GoogleProvider struct {
}

func (g *GoogleProvider) Provision(cluster *types.Cluster, platform *types.Platform) error {
	pltfrm := tcli.NewPlatform(clusterTemplate)
	pltfrm.AddProvider("google", google.Provider)
	return nil
}

func (g *GoogleProvider) Status(clusterName string, platform *types.Platform) (*types.ClusterInfo, error) {
	return nil, nil
}

func (g *GoogleProvider) Credentials(clusterName string, platform *types.Platform) ([]byte, error) {
	return nil, nil
}

func (g *GoogleProvider) Deprovision(clusterName string, platform *types.Platform) error {
	return nil
}

//Instantiate GCP provider
func New() *GoogleProvider {
	return &GoogleProvider{}
}
