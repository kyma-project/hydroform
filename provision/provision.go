package provision

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kyma-project/hydroform/provision/action"

	"github.com/kyma-project/hydroform/provision/internal/azure"
	"github.com/kyma-project/hydroform/provision/internal/gardener"
	"github.com/kyma-project/hydroform/provision/internal/kind"

	"github.com/kyma-project/hydroform/provision/internal/gcp"
	"github.com/kyma-project/hydroform/provision/internal/operator"
	"github.com/kyma-project/hydroform/provision/types"
)

// Currently the operator can not be changed at runtime, but the lib is designed so that it might be changed in the future
const provisioningOperator = operator.NativeOperator

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

	if runtime.GOOS == "windows" {
		provider.CredentialsFilePath = updateWindowsPath(provider.CredentialsFilePath)
	}

	switch provider.Type {
	case types.GCP:
		cl, err = gcp.New(provisioningOperator, ops...).Provision(cluster, provider)
	case types.Gardener:
		cl, err = gcp.New(provisioningOperator, ops...).Provision(cluster, provider)
	case types.AWS:
		err = errors.New("aws not supported yet")
	case types.Azure:
		cl, err = azure.New(provisioningOperator, ops...).Provision(cluster, provider)
	case types.Kind:
		cl, err = kind.New(provisioningOperator, ops...).Provision(cluster, provider)
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

	if runtime.GOOS == "windows" {
		provider.CredentialsFilePath = updateWindowsPath(provider.CredentialsFilePath)
	}

	switch provider.Type {
	case types.GCP:
		cs, err = gcp.New(provisioningOperator, ops...).Status(cluster, provider)
	case types.Gardener:
		cs, err = gardener.New(provisioningOperator, ops...).Status(cluster, provider)
	case types.AWS:
		err = errors.New("aws not supported yet")
	case types.Azure:
		cs, err = azure.New(provisioningOperator, ops...).Status(cluster, provider)
	case types.Kind:
		cs, err = kind.New(provisioningOperator, ops...).Status(cluster, provider)
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

	if runtime.GOOS == "windows" {
		provider.CredentialsFilePath = updateWindowsPath(provider.CredentialsFilePath)
	}

	switch provider.Type {
	case types.GCP:
		cr, err = gcp.New(provisioningOperator, ops...).Credentials(cluster, provider)
	case types.Gardener:
		cr, err = gardener.New(provisioningOperator, ops...).Credentials(cluster, provider)
	case types.AWS:
		err = errors.New("aws not supported yet")
	case types.Azure:
		cr, err = azure.New(provisioningOperator, ops...).Credentials(cluster, provider)
	case types.Kind:
		cr, err = kind.New(provisioningOperator, ops...).Credentials(cluster, provider)
	default:
		err = errors.New("unknown provider")
	}

	if err != nil {
		return cr, err
	}
	return cr, action.After()
}

func updateWindowsPath(windowsPath string) string {
	cleanWindowsPath := filepath.Clean(windowsPath)
	return strings.Replace(cleanWindowsPath, `\`, `\\`, -1)
}
