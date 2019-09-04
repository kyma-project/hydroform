package hydroform

import (
	gcp "github.com/kyma-incubator/hydroform/internal/gcp"
	"github.com/kyma-incubator/hydroform/types"
)

type Provider interface {
	Provision(cluster *types.Cluster, platform *types.Platform) error
	Status(clusterName string, platform *types.Platform) (*types.ClusterInfo, error)
	Credentials(clusterName string, platform *types.Platform) ([]byte, error)
	Deprovision(clusterName string, platform *types.Platform) error
}

func NewGoogleProvider() Provider {
	return gcp.New()
}

func NewAWSProvider() Provider {
	return nil
}

func NewAzureProvider() Provider {
	return nil
}
