package gcp

import (
	"errors"
	"strings"

	"github.com/kyma-incubator/hydroform/internal/operator"
	"github.com/kyma-incubator/hydroform/types"
)

var mandatoryConfigFields = []string{
	"credentials_file_path",
}

type GoogleProvider struct {
	provisionOperator operator.Cluster
}

func (g *GoogleProvider) validatePlatform(p *types.Platform) bool {
	for _, field := range mandatoryConfigFields {
		if _, ok := p.Configuration[field]; !ok {
			return false
		}
	}
	return true
}

func (g *GoogleProvider) Provision(cluster *types.Cluster, platform *types.Platform) error {
	if !g.validatePlatform(platform) {
		return errors.New("incomplete platform information")
	}
	config := map[string]interface{}{}
	config["cluster_name"] = cluster.Name
	config["node_count"] = platform.NodesCount
	config["machine_type"] = platform.MachineType
	config["kubernetes_version"] = cluster.KubernetesVersion
	config["location"] = platform.Location
	config["project"] = platform.ProjectName
	config["credentials_file_path"] = platform.Configuration["credentials_file_path"]

	return g.provisionOperator.Create("google", config)
}

func (g *GoogleProvider) Status(clusterName string, platform *types.Platform) (*types.ClusterInfo, error) {
	return nil, nil
}

func (g *GoogleProvider) Credentials(clusterName string, platform *types.Platform) ([]byte, error) {
	return nil, nil
}

func (g *GoogleProvider) Deprovision(clusterName string, platform *types.Platform) error {
	return nil
}

func New(provisionOperator string) *GoogleProvider {
	var op operator.Cluster
	switch strings.ToLower(provisionOperator) {
	case "terraform":
		op = &operator.Terraform{}
	default:
		op = &operator.Unknown{}
	}
	return &GoogleProvider{
		provisionOperator: op,
	}
}
