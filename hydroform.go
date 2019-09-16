package hydroform

import (
	"errors"

	"github.com/kyma-incubator/hydroform/internal/gcp"
	"github.com/kyma-incubator/hydroform/types"
)

const operator = "Terraform"

type Provider interface {
	Provision(cluster *types.Cluster, platform *types.Provider) (*types.ClusterInfo, error)
	Status(clusterName string, platform *types.Provider) (*types.ClusterInfo, error)
	Credentials(clusterName string, platform *types.Provider) ([]byte, error)
	Deprovision(clusterName string, platform *types.Provider) error
}

func Provision(cluster *types.Cluster, provider *types.Provider) (*types.ClusterInfo, error) {
	switch provider.Type {
	case types.GCP:
		return newGoogleProvider(operator).Provision(cluster, provider)
	case types.AWS:
		return nil, errors.New("aws not supported yet")
	case types.Azure:
		return nil, errors.New("azure not supported yet")
	default:
		return nil, errors.New("unknown provider")
	}
}

func Status(clusterName string, provider *types.Provider) (*types.ClusterStatus, error) {
	return nil, nil
}

func Credentials(clusterName string, provider *types.Provider) ([]byte, error) {
	return nil, nil
}

func Deprovision(clusterName string, provider *types.Provider) error {
	return nil
}

func newGoogleProvider(operatorName string) Provider {
	return gcp.New(operatorName)
}

func newAWSProvider(operatorName string) Provider {
	return nil
}

func newAzureProvider(operatorName string) Provider {
	return nil
}
