package gcp

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/option"

	"cloud.google.com/go/container"
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

func (g *GoogleProvider) Provision(cluster *types.Cluster,
	platform *types.Platform) (*types.ClusterInfo, error) {
	if !g.validatePlatform(platform) {
		return nil, errors.New("incomplete platform information")
	}
	config := map[string]interface{}{}
	config["cluster_name"] = cluster.Name
	config["node_count"] = platform.NodesCount
	config["machine_type"] = platform.MachineType
	config["disk_size"] = cluster.DiskSizeGB
	config["kubernetes_version"] = cluster.KubernetesVersion
	config["location"] = platform.Location
	config["project"] = platform.ProjectName
	config["credentials_file_path"] = platform.Configuration["credentials_file_path"]

	err := g.provisionOperator.Create("google", config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to provision gcp cluster")
	}

	return g.Status(cluster.Name, platform)
}

func (g *GoogleProvider) Status(clusterName string, platform *types.Platform) (*types.ClusterInfo, error) {
	containerClient, err := container.NewClient(context.Background(),
		platform.ProjectName,
		option.WithCredentialsFile(platform.Configuration["credentials_file_path"].(string)))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create GCP client")
	}
	cl, err := containerClient.Cluster(context.Background(), platform.Location, clusterName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster info")
	}
	info := &types.ClusterInfo{
		Status: string(cl.Status),
		IP:     cl.Endpoint,
	}

	return info, nil
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
