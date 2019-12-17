package azure

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/kyma-incubator/hydroform/internal/errs"
	terraform_operator "github.com/kyma-incubator/hydroform/internal/operator/terraform"

	"github.com/kyma-incubator/hydroform/internal/operator"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
)

// azureProvisioner implements Provisioner
type azureProvisioner struct {
	provisionOperator operator.Operator
}

// Provision requests provisioning of a new Kubernetes cluster on Azure with the given configurations.
func (a *azureProvisioner) Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error) {
	if err := a.validateInputs(cluster, provider); err != nil {
		return nil, err
	}

	config := a.loadConfigurations(cluster, provider)

	clusterInfo, err := a.provisionOperator.Create(provider.Type, config)
	if err != nil {
		return cluster, errors.Wrap(err, "unable to provision azure cluster")
	}

	cluster.ClusterInfo = clusterInfo
	return cluster, nil
}

// Status returns the ClusterStatus for the requested cluster.
func (a *azureProvisioner) Status(cluster *types.Cluster, p *types.Provider) (*types.ClusterStatus, error) {
	var state *statefile.File
	if cluster.ClusterInfo != nil && cluster.ClusterInfo.InternalState != nil {
		state = cluster.ClusterInfo.InternalState.TerraformState
	}

	if err := a.validateInputs(cluster, p); err != nil {
		return nil, err
	}

	cfg := a.loadConfigurations(cluster, p)

	return a.provisionOperator.Status(state, p.Type, cfg)
}

// Credentials returns the Kubeconfig file as a byte array for the requested cluster.
func (a *azureProvisioner) Credentials(cluster *types.Cluster, p *types.Provider) ([]byte, error) {
	if err := a.validateInputs(cluster, p); err != nil {
		return nil, err
	}
	if cluster.ClusterInfo == nil || cluster.ClusterInfo.InternalState == nil || cluster.ClusterInfo.InternalState.TerraformState == nil {
		// TODO add a way to get the kubeconfig from the state file if possible
		return nil, errors.New(errs.EmptyClusterInfo)
	}

	kubeconfig := cluster.ClusterInfo.InternalState.TerraformState.State.Modules[""].OutputValues["kube_config"].Value.AsString()

	return []byte(kubeconfig), nil
}

// Deprovision requests deprovisioning of an existing cluster on Azure with the given configurations.
func (a *azureProvisioner) Deprovision(cluster *types.Cluster, p *types.Provider) error {
	if err := a.validateInputs(cluster, p); err != nil {
		return err
	}

	config := a.loadConfigurations(cluster, p)

	var state *statefile.File
	if cluster.ClusterInfo != nil && cluster.ClusterInfo.InternalState != nil {
		state = cluster.ClusterInfo.InternalState.TerraformState
	}

	err := a.provisionOperator.Delete(state, p.Type, config)
	if err != nil {
		return errors.Wrap(err, "unable to deprovision azure cluster")
	}

	return nil
}

// New creates a new instance of azureProvisioner.
func New(operatorType operator.Type, ops ...types.Option) *azureProvisioner {
	// parse config
	os := &types.Options{}
	for _, o := range ops {
		o(os)
	}

	var op operator.Operator
	switch operatorType {
	case operator.TerraformOperator:
		tfOps := terraform_operator.ToTerraformOptions(os)
		op = terraform_operator.New(tfOps...)
	default:
		op = &operator.Unknown{}
	}

	return &azureProvisioner{
		provisionOperator: op,
	}
}

func (a *azureProvisioner) validateInputs(cluster *types.Cluster, provider *types.Provider) error {
	var errMessage string
	if cluster.NodeCount < 1 {
		errMessage += fmt.Sprintf(errs.CannotBeLess, "Cluster.NodeCount", 1)
	}
	// Matches the regex for a Azure cluster name.
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

	if errMessage != "" {
		return errors.New("input validation failed with the following information: " + errMessage)
	}

	return nil
}

func (a *azureProvisioner) loadConfigurations(cluster *types.Cluster, provider *types.Provider) map[string]interface{} {
	config := map[string]interface{}{}
	config["cluster_name"] = cluster.Name
	config["node_count"] = cluster.NodeCount
	config["machine_type"] = cluster.MachineType
	config["disk_size"] = cluster.DiskSizeGB
	config["kubernetes_version"] = cluster.KubernetesVersion
	config["location"] = cluster.Location
	config["project"] = provider.ProjectName
	config["client_id"], config["client_secret"] = azureCredentials(provider.CredentialsFilePath)
	for k, v := range provider.CustomConfigurations {
		config[k] = v
	}
	return config
}

// azureCredentials extracts the values of a credentials file to authenticate on azure.
// It expects a file following the TOML format (https://github.com/toml-lang/toml#user-content-spec), containing at least the CLIENT_ID and CLIENT_SECRET.
func azureCredentials(path string) (clientID, clientSecret string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	c := struct {
		ID     string `toml:"CLIENT_ID"`
		Secret string `toml:"CLIENT_SECRET"`
	}{}

	if _, err = toml.Decode(string(data), &c); err != nil {
		return
	}

	clientID = c.ID
	clientSecret = c.Secret

	return
}
