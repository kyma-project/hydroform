package kind

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/hashicorp/terraform/states/statefile"
	"github.com/kyma-incubator/hydroform/provision/internal/errs"
	"github.com/kyma-incubator/hydroform/provision/internal/operator"
	terraform_operator "github.com/kyma-incubator/hydroform/provision/internal/operator/terraform"
	"github.com/kyma-incubator/hydroform/provision/types"

	"github.com/pkg/errors"
)

// kindProvisioner implements Provisioner
type kindProvisioner struct {
	provisionOperator operator.Operator
}

// Provision requests provisioning of a new Kubernetes cluster on Kind with the given configurations.
func (k *kindProvisioner) Provision(cluster *types.Cluster, p *types.Provider) (*types.Cluster, error) {
	if err := k.validateInputs(cluster, p); err != nil {
		return nil, err
	}

	config := k.loadConfigurations(cluster, p)

	clusterInfo, err := k.provisionOperator.Create(p.Type, config)
	if err != nil {
		return cluster, errors.Wrap(err, "unable to provision kind cluster")
	}

	cluster.ClusterInfo = clusterInfo
	return cluster, nil
}

// Status returns the ClusterStatus for the requested cluster.
func (k *kindProvisioner) Status(cluster *types.Cluster, p *types.Provider) (*types.ClusterStatus, error) {
	var state *statefile.File
	if cluster.ClusterInfo != nil && cluster.ClusterInfo.InternalState != nil {
		state = cluster.ClusterInfo.InternalState.TerraformState
	}

	if err := k.validateInputs(cluster, p); err != nil {
		return nil, err
	}

	cfg := k.loadConfigurations(cluster, p)

	return k.provisionOperator.Status(state, p.Type, cfg)
}

// Credentials returns the Kubeconfig file as a byte array for the requested cluster.
func (k *kindProvisioner) Credentials(cluster *types.Cluster, p *types.Provider) ([]byte, error) {
	if err := k.validateInputs(cluster, p); err != nil {
		return nil, err
	}
	if cluster.ClusterInfo == nil || cluster.ClusterInfo.InternalState == nil || cluster.ClusterInfo.InternalState.TerraformState == nil {
		// TODO add a way to get the kubeconfig from the state file if possible
		return nil, errors.New(errs.EmptyClusterInfo)
	}

	// TODO find a better way to do this. Maybe this will be solved with the module way of doing things
	attrs := cluster.ClusterInfo.InternalState.TerraformState.State.Modules[""].Resources["kind.kind-cluster"].Instances[nil].Current.AttrsJSON

	var obj interface{}

	err := json.Unmarshal(attrs, &obj)
	if err != nil {
		return nil, err
	}

	kubeconfigPath := obj.(map[string]interface{})["kubeconfig_path"]
	kubeconfigBytes, err := ioutil.ReadFile(kubeconfigPath.(string))
	if err != nil {
		return nil, err
	}
	return kubeconfigBytes, nil
}

// Deprovision requests deprovisioning of an existing cluster on Kind with the given configurations.
func (k *kindProvisioner) Deprovision(cluster *types.Cluster, p *types.Provider) error {
	if err := k.validateInputs(cluster, p); err != nil {
		return err
	}

	config := k.loadConfigurations(cluster, p)

	var state *statefile.File
	if cluster.ClusterInfo != nil && cluster.ClusterInfo.InternalState != nil {
		state = cluster.ClusterInfo.InternalState.TerraformState
	}

	err := k.provisionOperator.Delete(state, p.Type, config)
	if err != nil {
		return errors.Wrap(err, "unable to deprovision kind cluster")
	}

	return nil
}

// New creates a new instance of gcpProvisioner.
func New(operatorType operator.Type, ops ...types.Option) *kindProvisioner {
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

	return &kindProvisioner{
		provisionOperator: op,
	}
}

func (k *kindProvisioner) validateInputs(cluster *types.Cluster, provider *types.Provider) error {

	var errMessage string
	// Matches the regex for a GCP cluster name.
	if match, _ := regexp.MatchString(`^(?:[a-z](?:[-a-z0-9]{0,37}[a-z0-9])?)$`, cluster.Name); !match {
		errMessage += fmt.Sprintf(errs.Custom, "Cluster.Name must start with a lowercase letter followed by up to 39 lowercase letters, "+
			"numbers, or hyphens, and cannot end with a hyphen")
	}
	if provider.ProjectName == "" {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.ProjectName")
	}

	if provider.CustomConfigurations != nil {
		if _, ok := provider.CustomConfigurations["node_image"]; !ok {
			errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfiguration.node_image")
		}
	}

	if errMessage != "" {
		return errors.New("input validation failed with the following information: " + errMessage)
	}

	return nil
}

func (k *kindProvisioner) loadConfigurations(cluster *types.Cluster, p *types.Provider) map[string]interface{} {
	config := map[string]interface{}{}
	config["cluster_name"] = cluster.Name
	config["project"] = p.ProjectName
	for k, v := range p.CustomConfigurations {
		config[k] = v
	}
	return config
}
