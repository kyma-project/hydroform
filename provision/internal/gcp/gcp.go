package gcp

import (
	"fmt"
	"regexp"

	"github.com/kyma-project/hydroform/provision/internal/errs"
	"github.com/kyma-project/hydroform/provision/internal/operator"
	"github.com/kyma-project/hydroform/provision/internal/operator/native"
	"github.com/kyma-project/hydroform/provision/types"
	"github.com/pkg/errors"
)

// gcpProvisioner implements Provisioner
type gcpProvisioner struct {
	provisionOperator operator.Operator
}

// New creates a new instance of gcpProvisioner.
func New(operatorType operator.Type, ops ...types.Option) *gcpProvisioner {
	// parse config
	os := &types.Options{}
	for _, o := range ops {
		o(os)
	}

	var op operator.Operator
	switch operatorType {
	case operator.NativeOperator:
		op = native.New(os)
	default:
		op = &operator.Unknown{}
	}

	return &gcpProvisioner{
		provisionOperator: op,
	}
}

// Provision requests provisioning of a new Kubernetes cluster on GCP with the given configurations.
func (g *gcpProvisioner) Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error) {
	if err := g.validateInputs(cluster, provider); err != nil {
		return cluster, err
	}

	config := g.loadConfigurations(cluster, provider)

	clusterInfo, err := g.provisionOperator.Create(provider.Type, config)
	if err != nil {
		return cluster, errors.Wrap(err, "unable to provision gcp cluster")
	}

	cluster.ClusterInfo = clusterInfo
	return cluster, nil
}

// Status returns the ClusterStatus for the requested cluster.
func (g *gcpProvisioner) Status(cluster *types.Cluster, p *types.Provider) (*types.ClusterStatus, error) {
	if err := g.validateInputs(cluster, p); err != nil {
		return nil, err
	}

	cfg := g.loadConfigurations(cluster, p)

	return g.provisionOperator.Status(cluster.ClusterInfo, p.Type, cfg)
}

// Credentials returns the Kubeconfig file as a byte array for the requested cluster.
func (g *gcpProvisioner) Credentials(cluster *types.Cluster, p *types.Provider) ([]byte, error) {
	return nil, errors.New("Not supported")
}

// Deprovision requests deprovisioning of an existing cluster on GCP with the given configurations.
func (g *gcpProvisioner) Deprovision(cluster *types.Cluster, p *types.Provider) error {
	if err := g.validateInputs(cluster, p); err != nil {
		return err
	}

	config := g.loadConfigurations(cluster, p)

	err := g.provisionOperator.Delete(cluster.ClusterInfo, p.Type, config)
	if err != nil {
		return errors.Wrap(err, "unable to deprovision gcp cluster")
	}

	return nil
}

func (g *gcpProvisioner) validateInputs(cluster *types.Cluster, provider *types.Provider) error {
	var errMessage string
	if cluster.NodeCount < 1 {
		errMessage += fmt.Sprintf(errs.CannotBeLess, "Cluster.NodeCount", 1)
	}
	// Matches the regex for a GCP cluster name.
	if match, _ := regexp.MatchString(`^(?:[a-z](?:[-a-z0-9]{0,37}[a-z0-9])?)$`, cluster.Name); !match {
		errMessage += fmt.Sprintf(errs.Custom, "Cluster.Name must start with a lowercase letter followed by up to 39 lowercase letters, "+
			"numbers, or hyphens, and cannot end with a hyphen")
	}
	if cluster.Location == "" {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Cluster.Location")
	}
	if cluster.MachineType == "" {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Cluster.MachineType")
	}
	if cluster.KubernetesVersion == "" {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Cluster.KubernetesVersion")
	}
	if cluster.DiskSizeGB < 0 {
		errMessage += fmt.Sprintf(errs.CannotBeLess, "Cluster.DiskSizeGB", 0)
	}

	if provider.CredentialsFilePath == "" {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CredentialsFilePath")
	}
	if provider.ProjectName == "" {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.ProjectName")
	}

	if errMessage != "" {
		return errors.New("input validation failed with the following information: " + errMessage)
	}

	return nil
}

func (g *gcpProvisioner) loadConfigurations(cluster *types.Cluster, provider *types.Provider) map[string]interface{} {
	config := map[string]interface{}{}
	config["cluster_name"] = cluster.Name
	config["node_count"] = cluster.NodeCount
	config["machine_type"] = cluster.MachineType
	config["disk_size"] = cluster.DiskSizeGB
	config["kubernetes_version"] = cluster.KubernetesVersion
	config["location"] = cluster.Location
	config["project"] = provider.ProjectName
	config["credentials_file_path"] = provider.CredentialsFilePath
	for k, v := range provider.CustomConfigurations {
		config[k] = v
	}
	return config
}
