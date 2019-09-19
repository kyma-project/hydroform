package hydroform

import (
	"errors"

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
		return nil, errors.New("azure not supported yet")
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
		return nil, errors.New("azure not supported yet")
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
		return nil, errors.New("azure not supported yet")
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
		return errors.New("azure not supported yet")
	default:
		return errors.New("unknown provider")
	}
}

func newGCPProvisioner(operatorType operator.OperatorType) Provisioner {
	return gcp.New(operatorType)
}

func newGardenerProvisioner(operatorType operator.OperatorType) Provisioner {
	return gardener.New(operatorType)
}

func newAWSProvisioner(operatorType operator.OperatorType) Provisioner {
	return nil
}

func newAzureProvisioner(operatorType operator.OperatorType) Provisioner {
	return nil
}
