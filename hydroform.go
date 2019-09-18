package hydroform

import (
	"errors"

	"github.com/kyma-incubator/hydroform/internal/gcp"
	"github.com/kyma-incubator/hydroform/types"
)

const operator = types.Terraform

type Provisioner interface {
	Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error)
	Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error)
	Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error)
	Deprovision(cluster *types.Cluster, provider *types.Provider) error
}

func Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error) {
	switch provider.Type {
	case types.GCP:
		return newGoogleProvisioner(operator).Provision(cluster, provider)
	case types.AWS:
		return nil, errors.New("aws not supported yet")
	case types.Azure:
		return nil, errors.New("azure not supported yet")
	default:
		return nil, errors.New("unknown provider")
	}
}

func Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error) {
	switch provider.Type {
	case types.GCP:
		return newGoogleProvisioner(operator).Status(cluster, provider)
	case types.AWS:
		return nil, errors.New("aws not supported yet")
	case types.Azure:
		return nil, errors.New("azure not supported yet")
	default:
		return nil, errors.New("unknown provider")
	}
}

func Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error) {
	switch provider.Type {
	case types.GCP:
		return newGoogleProvisioner(operator).Credentials(cluster, provider)
	case types.AWS:
		return nil, errors.New("aws not supported yet")
	case types.Azure:
		return nil, errors.New("azure not supported yet")
	default:
		return nil, errors.New("unknown provider")
	}
}

func Deprovision(cluster *types.Cluster, provider *types.Provider) error {
	switch provider.Type {
	case types.GCP:
		return newGoogleProvisioner(operator).Deprovision(cluster, provider)
	case types.AWS:
		return errors.New("aws not supported yet")
	case types.Azure:
		return errors.New("azure not supported yet")
	default:
		return errors.New("unknown provider")
	}
}

func newGoogleProvisioner(operatorType types.OperatorType) Provisioner {
	return gcp.New(operatorType)
}

func newAWSProvisioner(operatorType types.OperatorType) Provisioner {
	return nil
}

func newAzureProvisioner(operatorType types.OperatorType) Provisioner {
	return nil
}
