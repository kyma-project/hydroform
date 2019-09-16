package gcp

import (
	"strings"

	"github.com/kyma-incubator/hydroform/internal/operator"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
)

var mandatoryConfigFields = []string{
	"credentials_file_path",
}

type googleProvider struct {
	provisionOperator operator.Operator
}

func (g *googleProvider) validatePlatform(provider *types.Provider) bool {
	for _, field := range mandatoryConfigFields {
		if _, ok := provider.CustomConfigurations[field]; !ok {
			return false
		}
	}
	return true
}

func (g *googleProvider) Provision(cluster *types.Cluster, provider *types.Provider) (*types.ClusterInfo, error) {

	if !g.validatePlatform(provider) {
		return nil, errors.New("incomplete platform information")
	}

	config := map[string]interface{}{}
	config["cluster_name"] = cluster.Name
	config["node_count"] = cluster.NodeCount
	config["machine_type"] = cluster.MachineType
	config["disk_size"] = cluster.DiskSizeGB
	config["kubernetes_version"] = cluster.KubernetesVersion
	config["location"] = cluster.Location
	config["project"] = provider.ProjectName
	for k, v := range provider.CustomConfigurations {
		config[k] = v
	}

	err := g.provisionOperator.Create(provider.Type, config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to provision gcp cluster")
	}

	return g.Status(cluster.Name, provider)
}

func (g *googleProvider) Status(clusterName string, platform *types.Provider) (*types.ClusterInfo, error) {
	//containerClient, err := container.NewClient(context.Background(),
	//	platform.ProjectName,
	//	option.WithCredentialsFile(platform.Configuration["credentials_file_path"].(string)))
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to create GCP client")
	//}
	//cl, err := containerClient.Cluster(context.Background(), platform.Location, clusterName)
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to get cluster info")
	//}
	//info := &types.ClusterInfo{
	//	Status: string(cl.Status),
	//	IP:     cl.Endpoint,
	//}

	return nil, nil
}

func (g *googleProvider) Credentials(clusterName string, platform *types.Provider) ([]byte, error) {
	return nil, nil
}

func (g *googleProvider) Deprovision(clusterName string, platform *types.Provider) error {
	return nil
}

func New(operatorName string) *googleProvider {
	var op operator.Operator
	switch strings.ToLower(operatorName) {
	case "terraform":
		op = &operator.Terraform{}
	default:
		op = &operator.Unknown{}
	}
	return &googleProvider{
		provisionOperator: op,
	}
}
