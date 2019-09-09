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

func NewGoogleProvider(provisionOperator string) Provider {
	return gcp.New(provisionOperator)
}

func NewAWSProvider(provisionOperator string) Provider {
	return nil
}

func NewAzureProvider(provisionOperator string) Provider {
	return nil
}
