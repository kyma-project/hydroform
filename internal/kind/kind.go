package kind

import (
	"github.com/kyma-incubator/hydroform/internal/operator"
	"github.com/kyma-incubator/hydroform/types"
)

// kindProvisioner implements Provisioner
type kindProvisioner struct {
	provisionOperator operator.Operator
}

// Provision requests provisioning of a new Kubernetes cluster on GCP with the given configurations.
func (g *kindProvisioner) Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error) {
	return cluster, nil
}

// Status returns the ClusterStatus for the requested cluster.
func (g *kindProvisioner) Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error) {
	return cluster.ClusterInfo.Status, nil
}

// Credentials returns the Kubeconfig file as a byte array for the requested cluster.
func (g *kindProvisioner) Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error) {

	return nil, nil
}

// Deprovision requests deprovisioning of an existing cluster on GCP with the given configurations.
func (g *kindProvisioner) Deprovision(cluster *types.Cluster, provider *types.Provider) error {

	return nil
}

// New creates a new instance of gcpProvisioner.
func New(operatorType operator.Type) *kindProvisioner {
	var op operator.Operator

	switch operatorType {
	case operator.TerraformOperator:
		op = &operator.Terraform{}
	default:
		op = &operator.Unknown{}
	}

	return &kindProvisioner{
		provisionOperator: op,
	}
}

func (g *kindProvisioner) validateInputs(cluster *types.Cluster, provider *types.Provider) error {

	return nil
}

func (g *kindProvisioner) loadConfigurations(cluster *types.Cluster, provider *types.Provider) map[string]interface{} {
	config := map[string]interface{}{}
	config["cluster_name"] = cluster.Name
	config["kubernetes_version"] = cluster.KubernetesVersion
	for k, v := range provider.CustomConfigurations {
		config[k] = v
	}
	return config
}
