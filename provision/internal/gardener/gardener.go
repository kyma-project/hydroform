package gardener

import (
	"fmt"
	"regexp"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/hashicorp/terraform/states/statefile"
	"github.com/kyma-incubator/hydroform/provision/internal/errs"
	"github.com/kyma-incubator/hydroform/provision/internal/operator"
	terraform_operator "github.com/kyma-incubator/hydroform/provision/internal/operator/terraform"
	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	gcpProfile   string = "gcp"
	awsProfile   string = "aws"
	azureProfile string = "az"
)

type gardenerProvisioner struct {
	operator operator.Operator
}

func New(operatorType operator.Type, ops ...types.Option) *gardenerProvisioner {
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
	return &gardenerProvisioner{
		operator: op,
	}
}

func (g *gardenerProvisioner) Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error) {
	if err := g.validate(cluster, provider); err != nil {
		return cluster, err
	}

	config := g.loadConfigurations(cluster, provider)

	clusterInfo, err := g.operator.Create(provider.Type, config)
	if err != nil {
		return cluster, errors.Wrap(err, "unable to provision gardener cluster")
	}
	cluster.ClusterInfo = clusterInfo
	return cluster, nil
}

// Status returns the ClusterStatus for the requested cluster.
func (g *gardenerProvisioner) Status(cluster *types.Cluster, p *types.Provider) (*types.ClusterStatus, error) {
	var state *statefile.File
	if cluster.ClusterInfo != nil && cluster.ClusterInfo.InternalState != nil {
		state = cluster.ClusterInfo.InternalState.TerraformState
	}

	if err := g.validate(cluster, p); err != nil {
		return nil, err
	}

	cfg := g.loadConfigurations(cluster, p)

	return g.operator.Status(state, p.Type, cfg)
}

func (g *gardenerProvisioner) Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error) {
	if err := g.validate(cluster, provider); err != nil {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags("", provider.CredentialsFilePath)
	if err != nil {
		return nil, err
	}

	k8s, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	s, err := k8s.CoreV1().Secrets(fmt.Sprintf("garden-%s", provider.ProjectName)).Get(fmt.Sprintf("%s.kubeconfig", cluster.Name), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return s.Data["kubeconfig"], nil
}

func (g *gardenerProvisioner) Deprovision(cluster *types.Cluster, p *types.Provider) error {
	if err := g.validate(cluster, p); err != nil {
		return err
	}

	config := g.loadConfigurations(cluster, p)

	var state *statefile.File
	if cluster.ClusterInfo != nil && cluster.ClusterInfo.InternalState != nil {
		state = cluster.ClusterInfo.InternalState.TerraformState
	}

	err := g.operator.Delete(state, p.Type, config)
	if err != nil {
		return errors.Wrap(err, "unable to deprovision gardener cluster")
	}

	return nil
}

func (g *gardenerProvisioner) validate(cluster *types.Cluster, provider *types.Provider) error {
	var errMessage string

	// Cluster
	if cluster.NodeCount < 1 {
		errMessage += fmt.Sprintf(errs.CannotBeLess, "Cluster.NodeCount", 1)
	}
	// Matches the regex for a Gardener cluster name.
	if match, _ := regexp.MatchString(`^(?:[a-z](?:[-a-z0-9]{0,19}[a-z0-9])?)$`, cluster.Name); !match {
		errMessage += fmt.Sprintf(errs.Custom, "Cluster.Name must start with a lowercase letter followed by up to 19 lowercase letters, "+
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
	if cluster.DiskSizeGB <= 0 {
		errMessage += fmt.Sprintf(errs.CannotBeLess, "Cluster.DiskSizeGB", 0)
	}

	// Provider
	if provider.CredentialsFilePath == "" {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CredentialsFilePath")
	}
	if provider.ProjectName == "" {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.ProjectName")
	}

	// Custom gardener configuration
	targetProvider, ok := provider.CustomConfigurations["target_provider"]
	if ok {
		if targetProvider != string(types.GCP) && targetProvider != string(types.AWS) && targetProvider != string(types.Azure) {
			errMessage += fmt.Sprintf(errs.Custom, "Provider.CustomConfigurations['target_provider'] has to be one of: gcp, azure, aws")
		}
	} else {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['target_provider']")
	}
	if _, ok := provider.CustomConfigurations["target_secret"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['target_secret']")
	}

	if _, ok := provider.CustomConfigurations["disk_type"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['disk_type']")
	}
	if _, ok := provider.CustomConfigurations["worker_minimum"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['worker_minimum']")
	}
	if _, ok := provider.CustomConfigurations["worker_maximum"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['worker_maximum']")
	}
	if _, ok := provider.CustomConfigurations["worker_max_surge"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['worker_max_surge']")
	}
	if _, ok := provider.CustomConfigurations["worker_max_unavailable"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['worker_max_unavailable']")
	}
	if _, ok := provider.CustomConfigurations["workercidr"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['workercidr']")
	}
	if _, ok := provider.CustomConfigurations["zone"]; !ok && (targetProvider == string(types.GCP) || targetProvider == string(types.AWS)) {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['zone']")
	}
	if _, ok := provider.CustomConfigurations["publicscidr"]; !ok && targetProvider == string(types.AWS) {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['publicscidr']")
	}
	if _, ok := provider.CustomConfigurations["vpccidr"]; !ok && targetProvider == string(types.AWS) {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['vpccidr']")
	}
	if _, ok := provider.CustomConfigurations["internalscidr"]; !ok && targetProvider == string(types.AWS) {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['internalscidr']")
	}
	if _, ok := provider.CustomConfigurations["vnetcidr"]; !ok && targetProvider == string(types.Azure) {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['vnetcidr']")
	}

	if _, ok := provider.CustomConfigurations["machine_image_name"]; !ok && targetProvider == string(types.Azure) {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['machine_image_name']")
	}
	if _, ok := provider.CustomConfigurations["machine_image_version"]; !ok && targetProvider == string(types.Azure) {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['machine_image_version']")
	}

	if _, ok := provider.CustomConfigurations["networking_nodes"]; !ok && targetProvider == string(types.Azure) {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['networking_nodes']")
	}
	if _, ok := provider.CustomConfigurations["networking_pods"]; !ok && targetProvider == string(types.Azure) {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['networking_pods']")
	}
	if _, ok := provider.CustomConfigurations["networking_services"]; !ok && targetProvider == string(types.Azure) {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['networking_services']")
	}
	if _, ok := provider.CustomConfigurations["networking_type"]; !ok && targetProvider == string(types.Azure) {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['networking_type']")
	}

	if errMessage != "" {
		return errors.New("input validation failed with the following information: " + errMessage)
	}
	return nil
}

func (*gardenerProvisioner) loadConfigurations(cluster *types.Cluster, provider *types.Provider) map[string]interface{} {
	config := map[string]interface{}{}
	config["cluster_name"] = cluster.Name
	config["credentials_file_path"] = provider.CredentialsFilePath
	config["node_count"] = cluster.NodeCount
	config["machine_type"] = cluster.MachineType
	config["disk_size"] = cluster.DiskSizeGB
	config["kubernetes_version"] = cluster.KubernetesVersion
	config["location"] = cluster.Location
	config["project"] = provider.ProjectName
	config["namespace"] = fmt.Sprintf("garden-%s", provider.ProjectName)

	for k, v := range provider.CustomConfigurations {
		config[k] = v
	}
	switch config["target_provider"] {
	case string(types.GCP):
		config["target_profile"] = gcpProfile
	case string(types.AWS):
		config["target_profile"] = awsProfile
	case string(types.Azure):
		config["target_profile"] = azureProfile
	}
	return config
}
