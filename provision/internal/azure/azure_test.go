package azure

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/kyma-project/hydroform/provision/internal/operator/mocks"
	"github.com/pkg/errors"

	"github.com/kyma-project/hydroform/provision/types"
	"github.com/stretchr/testify/require"
)

func TestValidateInputs(t *testing.T) {
	g := &azureProvisioner{}

	cluster := &types.Cluster{
		CPU:               1,
		KubernetesVersion: "1.12",
		Name:              "hydro-cluster",
		DiskSizeGB:        30,
		NodeCount:         2,
		Location:          "europe-west3",
		MachineType:       "type1",
	}
	provider := &types.Provider{
		Type:                types.Azure,
		ProjectName:         "my-resource-group",
		CredentialsFilePath: "/path/to/credentials",
		CustomConfigurations: map[string]interface{}{
			"target_provider": "azure",
			"target_secret":   "secret-name",
			"disk_type":       "pd-standard",
			"zones":           "europe-west3-b",
		},
	}

	require.NoError(t, g.validateInputs(cluster, provider), "Validation should pass")

	cluster.NodeCount = -5
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when number of nodes is < 1")
	cluster.NodeCount = 2

	cluster.Name = ""
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when cluster name is empty")
	cluster.Name = "This_name_is_for_sure_way_too_long_for_the_cluster"
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when cluster name is too long")
	cluster.Name = "-invalid-start"
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when cluster name starts with '-'")
	cluster.Name = "invalid-end-"
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when cluster name ends with '-'")
	cluster.Name = "hydro-cluster"

	cluster.Location = ""
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when cluster location is empty")
	cluster.Location = "europe-west3"

	cluster.MachineType = ""
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when cluster machine type is empty")
	cluster.Location = "type1"

	cluster.KubernetesVersion = ""
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when Kubernetes version is empty")
	cluster.KubernetesVersion = "1.12"

	cluster.DiskSizeGB = 0
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when disk size is 0 or less")
	cluster.DiskSizeGB = 30

	provider.CredentialsFilePath = ""
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when credentials file path is empty")
	provider.CredentialsFilePath = "/path/to/credentials"

	provider.ProjectName = ""
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when project name is empty")
	provider.ProjectName = "/my-resource-group"

	delete(provider.CustomConfigurations, "target_provider")
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when target provider is empty")
	provider.CustomConfigurations["target_provider"] = "nimbus"
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when target provider is not supported")
	provider.CustomConfigurations["target_provider"] = "azure"

	delete(provider.CustomConfigurations, "target_secret")
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when target secret is empty")
	provider.CustomConfigurations["target_secret"] = "secret_name"

	delete(provider.CustomConfigurations, "disk_type")
	require.Error(t, g.validateInputs(cluster, provider), "Validation should fail when disk type is empty")
}

func TestLoadConfigurations(t *testing.T) {
	g := &azureProvisioner{}

	cluster := &types.Cluster{
		CPU:               1,
		KubernetesVersion: "1.12",
		Name:              "hydro-cluster",
		DiskSizeGB:        30,
		NodeCount:         2,
		Location:          "europe-west3",
		MachineType:       "type1",
	}
	provider := &types.Provider{
		Type:                types.Azure,
		ProjectName:         "my-resource-group",
		CredentialsFilePath: "./credentials-load-config.json",
		CustomConfigurations: map[string]interface{}{
			"target_provider": "azure",
			"target_secret":   "secret-name",
			"disk_type":       "pd-standard",
			"zones":           "europe-west3-b",
		},
	}

	err := fakeCredentials(provider.CredentialsFilePath)
	require.NoError(t, err, "Creating a fake credentials file should not have an error")
	defer os.Remove(provider.CredentialsFilePath)

	// happy path
	config, err := g.loadConfigurations(cluster, provider)
	require.NoError(t, err)

	require.Equal(t, cluster.Name, config["cluster_name"])
	require.Equal(t, "fake-subscription-id", config["subscription_id"])
	require.Equal(t, "fake-tenant-id", config["tenant_id"])
	require.Equal(t, "fake-client-id", config["client_id"])
	require.Equal(t, "fake-client-secret", config["client_secret"])
	require.Equal(t, cluster.NodeCount, config["agent_count"])
	require.Equal(t, cluster.MachineType, config["agent_vm_size"])
	require.Equal(t, cluster.DiskSizeGB, config["agent_disk_size"])
	require.Equal(t, cluster.KubernetesVersion, config["kubernetes_version"])
	require.Equal(t, cluster.Location, config["location"])
	require.Equal(t, provider.ProjectName, config["project"])

	for k, v := range provider.CustomConfigurations {
		require.Equal(t, v, config[k], fmt.Sprintf("Custom config %s is incorrect", k))
	}

	// credentials file not found
	provider.CredentialsFilePath = "/wrong/credentials/path"
	_, err = g.loadConfigurations(cluster, provider)
	require.Error(t, err)
}

func fakeCredentials(file string) error {
	fake := `{
  "subscription_id": "fake-subscription-id",
  "tenant_id": "fake-tenant-id",
  "client_id": "fake-client-id",
  "client_secret": "fake-client-secret"
}`

	return ioutil.WriteFile(file, []byte(fake), 0700)
}

func TestProvision(t *testing.T) {
	t.Parallel()
	mockOp := &mocks.Operator{}
	g := azureProvisioner{
		provisionOperator: mockOp,
	}

	cluster := &types.Cluster{
		CPU:               1,
		KubernetesVersion: "1.12",
		Name:              "hydro-cluster",
		DiskSizeGB:        30,
		NodeCount:         2,
		Location:          "europe-west3",
		MachineType:       "type1",
	}
	provider := &types.Provider{
		Type:                types.Azure,
		ProjectName:         "my-resource-group",
		CredentialsFilePath: "./credentials-provision.json",
		CustomConfigurations: map[string]interface{}{
			"target_provider": "azure",
			"target_secret":   "secret-name",
			"disk_type":       "pd-standard",
			"zones":           "europe-west3-b",
		},
	}
	err := fakeCredentials(provider.CredentialsFilePath)
	require.NoError(t, err, "Creating a fake credentials file should not have an error")
	defer os.Remove(provider.CredentialsFilePath)

	result := &types.ClusterInfo{
		CertificateAuthorityData: []byte("My cert"),
		Endpoint:                 "https://cluster-url.fake",
		Status: &types.ClusterStatus{
			Phase: types.Provisioned,
		},
	}

	cfg, err := g.loadConfigurations(cluster, provider)
	require.NoError(t, err)

	mockOp.On("Create", types.Azure, cfg).Return(result, nil)

	cluster, err = g.Provision(cluster, provider)
	require.NoError(t, err, "Provision should succeed")
	require.Equal(t, result, cluster.ClusterInfo, "The cluster info returned from the operator should be in the cluster returned by Provision")

	badCluster := &types.Cluster{
		CPU: 1,
	}

	cfg, err = g.loadConfigurations(badCluster, provider)
	require.NoError(t, err)
	mockOp.On("Create", types.Azure, cfg).Return(badCluster, errors.New("Unable to provision cluster"))

	_, err = g.Provision(badCluster, provider)
	require.Error(t, err, "Provision should fail")
}

func TestDeprovision(t *testing.T) {
	t.Parallel()
	mockOp := &mocks.Operator{}
	g := azureProvisioner{
		provisionOperator: mockOp,
	}

	cluster := &types.Cluster{
		CPU:               1,
		KubernetesVersion: "1.12",
		Name:              "hydro-cluster",
		DiskSizeGB:        30,
		NodeCount:         2,
		Location:          "europe-west3",
		MachineType:       "type1",
		ClusterInfo:       &types.ClusterInfo{},
	}
	provider := &types.Provider{
		Type:                types.Azure,
		ProjectName:         "my-resource-group",
		CredentialsFilePath: "./credentials-deprovision.json",
		CustomConfigurations: map[string]interface{}{
			"target_provider": "azure",
			"target_secret":   "secret-name",
			"disk_type":       "pd-standard",
			"zones":           "europe-west3-b",
		},
	}

	err := fakeCredentials(provider.CredentialsFilePath)
	require.NoError(t, err, "Creating a fake credentials file should not have an error")
	defer os.Remove(provider.CredentialsFilePath)

	cfg, err := g.loadConfigurations(cluster, provider)
	require.NoError(t, err)

	mockOp.On("Delete", cluster.ClusterInfo, types.Azure, cfg).Return(nil)

	err = g.Deprovision(cluster, provider)
	require.NoError(t, err, "Deprovision should succeed")

	provider.ProjectName = "invalid-resource-group"
	cfg, err = g.loadConfigurations(cluster, provider)
	require.NoError(t, err)

	mockOp.On("Delete", cluster.ClusterInfo, types.Azure, cfg).Return(errors.New("Unable to deprovision cluster"))

	err = g.Deprovision(cluster, provider)
	require.Error(t, err, "Deprovision should fail")
}
