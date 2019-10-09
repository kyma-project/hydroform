package gardener

import (
	"fmt"
	"regexp"

	gardener_core "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	gardener_types "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	gardener_api "github.com/gardener/gardener/pkg/client/garden/clientset/versioned/typed/garden/v1beta1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyma-incubator/hydroform/internal/errs"
	"github.com/kyma-incubator/hydroform/internal/operator"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type gardenerProvisioner struct {
	operator operator.Operator
}

func New(operatorType operator.Type) *gardenerProvisioner {
	var op operator.Operator
	switch operatorType {
	case operator.TerraformOperator:
		op = &operator.Terraform{}
	default:
		op = &operator.Unknown{}
	}
	return &gardenerProvisioner{
		operator: op,
	}
}

func (g *gardenerProvisioner) Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error) {
	if err := g.validate(cluster, provider); err != nil {
		return nil, err
	}

	config := g.loadConfigurations(cluster, provider)

	clusterInfo, err := g.operator.Create(provider.Type, config)
	if err != nil {
		return cluster, errors.Wrap(err, "unable to provision gardener cluster")
	}
	cluster.ClusterInfo = clusterInfo
	return cluster, nil
}

func (g *gardenerProvisioner) Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error) {
	if err := g.validate(cluster, provider); err != nil {
		return nil, err
	}

	c, err := clientcmd.BuildConfigFromFlags("", provider.CredentialsFilePath)
	if err != nil {
		return nil, err
	}

	gardenerClient, err := gardener_api.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	shoot, err := gardenerClient.Shoots(fmt.Sprintf("garden-%s", provider.ProjectName)).Get(cluster.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &types.ClusterStatus{
		Phase: convertGardenertatus(shoot.Status),
	}, nil
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

func (g *gardenerProvisioner) Deprovision(cluster *types.Cluster, provider *types.Provider) error {
	if err := g.validate(cluster, provider); err != nil {
		return err
	}

	config := g.loadConfigurations(cluster, provider)

	err := g.operator.Delete(cluster.ClusterInfo.InternalState, provider.Type, config)
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
	if v, ok := provider.CustomConfigurations["target_provider"]; ok {
		if v != string(types.GCP) && v != string(types.AWS) && v != string(types.Azure) {
			errMessage += fmt.Sprintf(errs.Custom, "Provider.CustomConfigurations['target_provider'] has to be one of: gcp, azure, aws")
		}
	} else {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['target_provider']")
	}
	if _, ok := provider.CustomConfigurations["target_secret"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['target_secret']")
	}
	if _, ok := provider.CustomConfigurations["zone"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['zone']")
	}
	if _, ok := provider.CustomConfigurations["disk_type"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['disk_type']")
	}
	if _, ok := provider.CustomConfigurations["autoscaler_min"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['autoscaler_min']")
	}
	if _, ok := provider.CustomConfigurations["autoscaler_max"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['autoscaler_max']")
	}
	if _, ok := provider.CustomConfigurations["max_surge"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['max_surge']")
	}
	if _, ok := provider.CustomConfigurations["max_unavailable"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['max_unavailable']")
	}
	if _, ok := provider.CustomConfigurations["cidr"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['cidr']")
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

	for k, v := range provider.CustomConfigurations {
		config[k] = v
	}
	return config
}

// Possible values for the Gardener Cluster Status:
// Processing - indicates the cluster is being created.
// Succeeded - indicates the cluster has been created and is fully usable.
// Error  - indicates the cluster may be unusable.
// Failed - indicates that the creation operation failed.
// Pending - indicates that the creation has not started yet.
// Aborted - indicates that an external agent aborted the operation.
func convertGardenertatus(status gardener_types.ShootStatus) types.Phase {
	switch status.LastOperation.State {
	case gardener_core.LastOperationStateProcessing:
		return types.Provisioning
	case gardener_core.LastOperationStatePending:
		return types.Pending
	case gardener_core.LastOperationStateSucceeded:
		return types.Provisioned
	case gardener_core.LastOperationStateError:
		return types.Errored
	case gardener_core.LastOperationStateFailed:
		return types.Errored
	case gardener_core.LastOperationStateAborted:
		return types.Errored
	default:
		return types.Unknown
	}
}
