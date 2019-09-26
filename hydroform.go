package hydroform

import (
	"errors"

	"github.com/kyma-incubator/hydroform/action"

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
	var err error
	var cl *types.Cluster

	if err = action.Before(); err != nil {
		return cl, err
	}

	switch provider.Type {
	case types.GCP:
		cl, err = newGCPProvisioner(provisionOperator).Provision(cluster, provider)
	case types.Gardener:
		cl, err = newGardenerProvisioner(provisionOperator).Provision(cluster, provider)
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

func Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error) {
	var err error
	var cs *types.ClusterStatus

	if err = action.Before(); err != nil {
		return cs, err
	}

	switch provider.Type {
	case types.GCP:
		cs, err = newGCPProvisioner(provisionOperator).Status(cluster, provider)
	case types.Gardener:
		cs, err = newGardenerProvisioner(provisionOperator).Status(cluster, provider)
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

func Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error) {
	var err error
	var cr []byte

	if err = action.Before(); err != nil {
		return cr, err
	}
	switch provider.Type {
	case types.GCP:
		cr, err = newGCPProvisioner(provisionOperator).Credentials(cluster, provider)
	case types.Gardener:
		cr, err = newGardenerProvisioner(provisionOperator).Credentials(cluster, provider)
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

func Deprovision(cluster *types.Cluster, provider *types.Provider) error {
	var err error

	if err = action.Before(); err != nil {
		return err
	}
	switch provider.Type {
	case types.GCP:
		err = newGCPProvisioner(provisionOperator).Deprovision(cluster, provider)
	case types.Gardener:
		err = newGardenerProvisioner(provisionOperator).Deprovision(cluster, provider)
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
