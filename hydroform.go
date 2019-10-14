package hydroform

import (
	"errors"
	"github.com/kyma-incubator/hydroform/internal/aks"

	"github.com/kyma-incubator/hydroform/internal/gardener"

	"github.com/kyma-incubator/hydroform/internal/gcp"
	"github.com/kyma-incubator/hydroform/internal/operator"
	"github.com/kyma-incubator/hydroform/types"
)

const provisionOperator = operator.TerraformOperator

type Provisioner interface {
	Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error)
	Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error)
	Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error)
	Deprovision(cluster *types.Cluster, provider *types.Provider) error
}

func Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error) {
	switch provider.Type {
	case types.GCP:
		return newGCPProvisioner(provisionOperator).Provision(cluster, provider)
	case types.Gardener:
		return newGardenerProvisioner(provisionOperator).Provision(cluster, provider)
	case types.AWS:
		return nil, errors.New("aws not supported yet")
	case types.Azure:
		return newAzureProvisioner(provisionOperator).Provision(cluster, provider)
	default:
		return nil, errors.New("unknown provider")
	}
}

func Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error) {
	switch provider.Type {
	case types.GCP:
		return newGCPProvisioner(provisionOperator).Status(cluster, provider)
	case types.Gardener:
		return newGardenerProvisioner(provisionOperator).Status(cluster, provider)
	case types.AWS:
		return nil, errors.New("aws not supported yet")
	case types.Azure:
		return newAzureProvisioner(provisionOperator).Status(cluster, provider)
	default:
		return nil, errors.New("unknown provider")
	}
}

func Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error) {
	switch provider.Type {
	case types.GCP:
		return newGCPProvisioner(provisionOperator).Credentials(cluster, provider)
	case types.Gardener:
		return newGardenerProvisioner(provisionOperator).Credentials(cluster, provider)
	case types.AWS:
		return nil, errors.New("aws not supported yet")
	case types.Azure:
		return newAzureProvisioner(provisionOperator).Credentials(cluster, provider)
	default:
		return nil, errors.New("unknown provider")
	}
}

func Deprovision(cluster *types.Cluster, provider *types.Provider) error {
	switch provider.Type {
	case types.GCP:
		return newGCPProvisioner(provisionOperator).Deprovision(cluster, provider)
	case types.Gardener:
		return newGardenerProvisioner(provisionOperator).Deprovision(cluster, provider)
	case types.AWS:
		return errors.New("aws not supported yet")
	case types.Azure:
		return newAzureProvisioner(provisionOperator).Deprovision(cluster, provider)
	default:
		return errors.New("unknown provider")
	}
}

func newGCPProvisioner(operatorType operator.Type) Provisioner {
	return gcp.New(operatorType)
}

func newGardenerProvisioner(operatorType operator.Type) Provisioner {
	return gardener.New(operatorType)
}

func newAWSProvisioner(operatorType operator.Type) Provisioner {
	return nil
}

func newAzureProvisioner(operatorType operator.Type) Provisioner {
	return aks.New(operatorType)
}
