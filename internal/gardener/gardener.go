package gardener

import (
	"fmt"
	"regexp"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyma-incubator/hydroform/internal/errs"
	"github.com/kyma-incubator/hydroform/internal/operator"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var mandatoryConfigFields = []string{
	"target_provider",
	"target_secret",
	"disk_type",
}

type gardenerProvisioner struct {
	operator operator.Operator
}

func New(operatorType operator.OperatorType) *gardenerProvisioner {
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
	return nil, errors.New("Gardener cluster status not implemented")
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
	if _, ok := provider.CustomConfigurations["disk_type"]; !ok {
		errMessage += fmt.Sprintf(errs.CannotBeEmpty, "Provider.CustomConfigurations['disk_type']")
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
