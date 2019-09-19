package hydroform

import (
	"errors"

	"github.com/kyma-incubator/hydroform/internal/gardener"

	"github.com/kyma-incubator/hydroform/internal/gcp"
	"github.com/kyma-incubator/hydroform/internal/operator"
	"github.com/kyma-incubator/hydroform/types"
)

const provisioningOperator = operator.TerraformOperator

// Provisioner is the generic hydroform interface for the provisioners.
type Provisioner interface {
	Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error)
	Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error)
	Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error)
	Deprovision(cluster *types.Cluster, provider *types.Provider) error
}

// Provision requests provisioning of a new cluster with the given configuration on the given provider.
func Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error) {
	switch provider.Type {
	case types.GCP:
		return newGCPProvisioner(provisioningOperator).Provision(cluster, provider)
	case types.Gardener:
		return newGardenerProvisioner(provisioningOperator).Provision(cluster, provider)
	case types.AWS:
		return nil, errors.New("aws not supported yet")
	case types.Azure:
		return nil, errors.New("azure not supported yet")
	default:
		return nil, errors.New("unknown provider")
	}
}

// Status returns the ClusterStatus for the requested cluster.
func Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error) {
	switch provider.Type {
	case types.GCP:
		return newGCPProvisioner(provisioningOperator).Status(cluster, provider)
	case types.Gardener:
		return newGardenerProvisioner(provisioningOperator).Status(cluster, provider)
	case types.AWS:
		return nil, errors.New("aws not supported yet")
	case types.Azure:
		return nil, errors.New("azure not supported yet")
	default:
		return nil, errors.New("unknown provider")
	}
}

// Credentials returns the Kubeconfig file as a byte array for the requested cluster.
func Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error) {
	switch provider.Type {
	case types.GCP:
		return newGCPProvisioner(provisioningOperator).Credentials(cluster, provider)
	case types.Gardener:
		return newGardenerProvisioner(provisioningOperator).Credentials(cluster, provider)
	case types.AWS:
		return nil, errors.New("aws not supported yet")
	case types.Azure:
		return nil, errors.New("azure not supported yet")
	default:
		return nil, errors.New("unknown provider")
	}
}

// Deprovision requests deprovisioning of the given cluster on the given provider.
func Deprovision(cluster *types.Cluster, provider *types.Provider) error {
	switch provider.Type {
	case types.GCP:
		return newGCPProvisioner(provisioningOperator).Deprovision(cluster, provider)
	case types.Gardener:
		return newGardenerProvisioner(provisioningOperator).Deprovision(cluster, provider)
	case types.AWS:
		return errors.New("aws not supported yet")
	case types.Azure:
		return errors.New("azure not supported yet")
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
	return nil
}
