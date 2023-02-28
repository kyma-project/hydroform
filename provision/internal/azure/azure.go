package azure

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/kyma-project/hydroform/provision/internal/errs"
	"github.com/kyma-project/hydroform/provision/internal/operator"
	"github.com/kyma-project/hydroform/provision/internal/operator/native"
	"github.com/kyma-project/hydroform/provision/types"
	"github.com/pkg/errors"
)

// AzureProvisioner implements Provisioner
// nolint:revive
type AzureProvisioner struct {
	provisionOperator operator.Operator
}

// New creates a new instance of AzureProvisioner.
func New(operatorType operator.Type, ops ...types.Option) *AzureProvisioner {
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

	return &AzureProvisioner{
		provisionOperator: op,
	}
}

// Provision requests provisioning of a new Kubernetes cluster on Azure with the given configurations.
func (a *AzureProvisioner) Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error) {
	if err := a.validateInputs(cluster, provider); err != nil {
		return cluster, err
	}

	config, err := a.loadConfigurations(cluster, provider)
	if err != nil {
		return cluster, err
	}

	clusterInfo, err := a.provisionOperator.Create(provider.Type, config)
	if err != nil {
		return cluster, errors.Wrap(err, "unable to provision azure cluster")
	}

	cluster.ClusterInfo = clusterInfo
	return cluster, nil
}

// Status returns the ClusterStatus for the requested cluster.
func (a *AzureProvisioner) Status(cluster *types.Cluster, p *types.Provider) (*types.ClusterStatus, error) {
	if err := a.validateInputs(cluster, p); err != nil {
		return nil, err
	}

	cfg, err := a.loadConfigurations(cluster, p)
	if err != nil {
		return nil, err
	}

	return a.provisionOperator.Status(cluster.ClusterInfo, p.Type, cfg)
}

// Credentials returns the Kubeconfig file as a byte array for the requested cluster.
func (a *AzureProvisioner) Credentials(cluster *types.Cluster, p *types.Provider) ([]byte, error) {
	return nil, errors.New("Not supported")
}

// Deprovision requests deprovisioning of an existing cluster on Azure with the given configurations.
func (a *AzureProvisioner) Deprovision(cluster *types.Cluster, p *types.Provider) error {
	if err := a.validateInputs(cluster, p); err != nil {
		return err
	}

	config, err := a.loadConfigurations(cluster, p)
	if err != nil {
		return err
	}

	if err = a.provisionOperator.Delete(cluster.ClusterInfo, p.Type, config); err != nil {
		return errors.Wrap(err, "unable to deprovision azure cluster")
	}

	return nil
}

func (a *AzureProvisioner) validateInputs(cluster *types.Cluster, provider *types.Provider) error {
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

func (a *AzureProvisioner) loadConfigurations(cluster *types.Cluster, provider *types.Provider) (map[string]interface{}, error) {
	config := map[string]interface{}{}
	config["cluster_name"] = cluster.Name
	config["agent_count"] = cluster.NodeCount
	config["agent_vm_size"] = cluster.MachineType
	config["agent_disk_size"] = cluster.DiskSizeGB
	config["kubernetes_version"] = cluster.KubernetesVersion
	config["location"] = cluster.Location
	config["project"] = provider.ProjectName
	config["resource_group"] = provider.ProjectName

	var err error
	config["subscription_id"], config["tenant_id"], config["client_id"], config["client_secret"], err = azureCredentials(provider.CredentialsFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading credentials")
	}

	for k, v := range provider.CustomConfigurations {
		config[k] = v
	}

	return config, nil
}

// azureCredentials extracts the values of a credentials file to authenticate on azure.
// It expects a JSON file containing the subscription ID, tenant ID, client ID and client secret.
func azureCredentials(path string) (subscriptionID, tenantID, clientID, clientSecret string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	c := struct {
		SubscriptionID string `json:"subscription_id"`
		TenantID       string `json:"tenant_id"`
		ClientID       string `json:"client_id"`
		Secret         string `json:"client_secret"`
	}{}

	if err = json.Unmarshal(data, &c); err != nil {
		return
	}

	subscriptionID = c.SubscriptionID
	tenantID = c.TenantID
	clientID = c.ClientID
	clientSecret = c.Secret

	return
}
