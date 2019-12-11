package hydroform

import (
	"errors"

	"github.com/kyma-incubator/hydroform/provision/action"

	"github.com/kyma-incubator/hydroform/provision/internal/gardener"

	"github.com/kyma-incubator/hydroform/provision/internal/gcp"
	"github.com/kyma-incubator/hydroform/provision/internal/operator"
	"github.com/kyma-incubator/hydroform/provision/types"
)

const provisioningOperator = operator.TerraformOperator

// Provisioner is the Hydroform interface that groups Provision, Status, Credentials, and Deprovision functions used to create and manage a cluster.
type Provisioner interface {
	Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error)
	Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error)
	Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error)
	Deprovision(cluster *types.Cluster, provider *types.Provider) error
}

// Provision creates a new cluster for a given provider based on specific cluster and provider parameters. It returns a cluster object enriched with information from the provider, such as the IP address or the connection endpoint. This object is necessary for the other operations, such as retrieving the cluster status or deprovisioning the cluster. If the cluster cannot be created, the function returns an error.
func Provision(cluster *types.Cluster, provider *types.Provider, ops ...types.Option) (*types.Cluster, error) {
	var err error
	var cl *types.Cluster

	if err = action.Before(); err != nil {
		return cl, err
	}

	switch provider.Type {
	case types.GCP:
		cl, err = newGCPProvisioner(provisioningOperator, ops...).Provision(cluster, provider)
	case types.Gardener:
		cl, err = newGardenerProvisioner(provisioningOperator, ops...).Provision(cluster, provider)
	case types.AWS:
		err = errors.New("aws not supported yet")
	case types.Azure:
		err = errors.New("azure not supported yet")
	default:
		err = errors.New("unknown provider")
	}

	if err != nil {
		return cl, err
	}
	return cl, action.After()
}

// Status returns the cluster status for a given provider, or an error if providing the status is not possible. The possible status values are defined in the ClusterStatus type.
func Status(cluster *types.Cluster, provider *types.Provider, ops ...types.Option) (*types.ClusterStatus, error) {
	var err error
	var cs *types.ClusterStatus

	if err = action.Before(); err != nil {
		return cs, err
	}

	switch provider.Type {
	case types.GCP:
		cs, err = newGCPProvisioner(provisioningOperator, ops...).Status(cluster, provider)
	case types.Gardener:
		cs, err = newGardenerProvisioner(provisioningOperator, ops...).Status(cluster, provider)
	case types.AWS:
		err = errors.New("aws not supported yet")
	case types.Azure:
		err = errors.New("azure not supported yet")
	default:
		err = errors.New("unknown provider")
	}

	if err != nil {
		return cs, err
	}
	return cs, action.After()
}

// Credentials returns the kubeconfig for a specific cluster as a byte array.
func Credentials(cluster *types.Cluster, provider *types.Provider, ops ...types.Option) ([]byte, error) {
	var err error
	var cr []byte

	if err = action.Before(); err != nil {
		return cr, err
	}
	switch provider.Type {
	case types.GCP:
		cr, err = newGCPProvisioner(provisioningOperator, ops...).Credentials(cluster, provider)
	case types.Gardener:
		cr, err = newGardenerProvisioner(provisioningOperator, ops...).Credentials(cluster, provider)
	case types.AWS:
		err = errors.New("aws not supported yet")
	case types.Azure:
		err = errors.New("azure not supported yet")
	default:
		err = errors.New("unknown provider")
	}

	if err != nil {
		return cr, err
	}
	return cr, action.After()
}

// Deprovision removes an existing cluster along or returns an error if removing the cluster is not possible.
func Deprovision(cluster *types.Cluster, provider *types.Provider, ops ...types.Option) error {
	var err error

	if err = action.Before(); err != nil {
		return err
	}
	switch provider.Type {
	case types.GCP:
		err = newGCPProvisioner(provisioningOperator, ops...).Deprovision(cluster, provider)
	case types.Gardener:
		err = newGardenerProvisioner(provisioningOperator, ops...).Deprovision(cluster, provider)
	case types.AWS:
		err = errors.New("aws not supported yet")
	case types.Azure:
		err = errors.New("azure not supported yet")
	default:
		err = errors.New("unknown provider")
	}
	if err != nil {
		return err
	}
	return action.After()
}

func newGCPProvisioner(operatorType operator.Type, ops ...types.Option) Provisioner {
	return gcp.New(operatorType, ops...)
}

func newGardenerProvisioner(operatorType operator.Type, ops ...types.Option) Provisioner {
	return gardener.New(operatorType, ops...)
}

func newAWSProvisioner(operatorType operator.Type, ops ...types.Option) Provisioner {
	return nil
}

func newAzureProvisioner(operatorType operator.Type, ops ...types.Option) Provisioner {
	return nil
}
