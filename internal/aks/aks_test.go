package aks


import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/hydroform/internal/terraform"

	"github.com/kyma-incubator/hydroform/internal/operator/mocks"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/require"
)

const convertError = "Status [%s] should be converted to [%s]"

func TestConvertaksState(t *testing.T) {

	require.Equal(t, types.Unknown, convertAKSStatus(""), fmt.Sprintf(convertError, "\"\"", types.Unknown))
	require.Equal(t, types.Provisioning, convertAKSStatus("PROVISIONING"), fmt.Sprintf(convertError, "PROVISIONING", types.Provisioning))
	require.Equal(t, types.Pending, convertAKSStatus("RECONCILING"), fmt.Sprintf(convertError, "RECONCILING", types.Pending))
	require.Equal(t, types.Stopping, convertAKSStatus("STOPPING"), fmt.Sprintf(convertError, "STOPPING", types.Stopping))
	require.Equal(t, types.Errored, convertAKSStatus("ERROR"), fmt.Sprintf(convertError, "ERROR", types.Errored))
	require.Equal(t, types.Errored, convertAKSStatus("DEGRADED"), fmt.Sprintf(convertError, "DEGRADED", types.Errored))
	require.Equal(t, types.Provisioned, convertAKSStatus("RUNNING"), fmt.Sprintf(convertError, "RUNNING", types.Provisioned))
}

func TestValidateInputs(t *testing.T) {

	a := &aksProvisioner{}

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
		ProjectName:         "my-project",
		CredentialsFilePath: "/path/to/credentials",
		CustomConfigurations: map[string]interface{}{
			"target_provider": "azure",
			"target_secret":   "secret-name",
			"disk_type":       "pd-standard",
			"region":            "europe-west3-b",
		},
	}

	require.NoError(t, a.validateInputs(cluster, provider), "Validation should pass")

	cluster.NodeCount = -5
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when number of nodes is < 1")
	cluster.NodeCount = 2

	cluster.Name = ""
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when cluster name is empty")
	cluster.Name = "This_name_is_for_sure_way_too_long_for_the_cluster"
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when cluster name is too long")
	cluster.Name = "-invalid-start"
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when cluster name starts with '-'")
	cluster.Name = "invalid-end-"
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when cluster name ends with '-'")
	cluster.Name = "hydro-cluster"

	cluster.Location = ""
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when cluster location is empty")
	cluster.Location = "europe-west3"

	cluster.MachineType = ""
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when cluster machine type is empty")
	cluster.Location = "type1"

	cluster.KubernetesVersion = ""
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when Kubernetes version is empty")
	cluster.KubernetesVersion = "1.12"

	cluster.DiskSizeGB = 0
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when disk size is 0 or less")
	cluster.DiskSizeGB = 30

	provider.CredentialsFilePath = ""
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when credentials file path is empty")
	provider.CredentialsFilePath = "/path/to/credentials"

	provider.ProjectName = ""
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when project name is empty")
	provider.CredentialsFilePath = "/my-project"

	delete(provider.CustomConfigurations, "target_provider")
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when target provider is empty")
	provider.CustomConfigurations["target_provider"] = "nimbus"
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when target provider is not supported")
	provider.CustomConfigurations["target_provider"] = "aks"

	delete(provider.CustomConfigurations, "target_secret")
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when target secret is empty")
	provider.CustomConfigurations["target_secret"] = "secret_name"

	delete(provider.CustomConfigurations, "disk_type")
	require.Error(t, a.validateInputs(cluster, provider), "Validation should fail when disk type is empty")
}

func TestLoadConfigurations(t *testing.T) {
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
		ProjectName:         "my-project",
		CredentialsFilePath: "/path/to/credentials",
		CustomConfigurations: map[string]interface{}{
			"target_provider": "azure",
			"target_secret":   "secret-name",
			"disk_type":       "pd-standard",
			"region":            "europe-west3-b",
		},
	}

	config := loadConfigurations(cluster, provider)

	require.Equal(t, cluster.Name, config["cluster_name"])
	require.Equal(t, provider.CredentialsFilePath, config["credentials_file_path"])
	require.Equal(t, cluster.NodeCount, config["node_count"])
	require.Equal(t, cluster.MachineType, config["machine_type"])
	require.Equal(t, cluster.DiskSizeGB, config["disk_size"])
	require.Equal(t, cluster.KubernetesVersion, config["kubernetes_version"])
	require.Equal(t, cluster.Location, config["location"])
	require.Equal(t, provider.ProjectName, config["project"])

	for k, v := range provider.CustomConfigurations {
		require.Equal(t, v, config[k], fmt.Sprintf("Custom config %s is incorrect", k))
	}
}

func TestProvision(t *testing.T) {
	mockOp := &mocks.Operator{}
	a := aksProvisioner{
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
		ProjectName:         "my-project",
		CredentialsFilePath: "/path/to/credentials",
		CustomConfigurations: map[string]interface{}{
			"target_provider": "azure",
			"target_secret":   "secret-name",
			"disk_type":       "pd-standard",
			"region":            "europe-west3-b",
		},
	}

	result := &types.ClusterInfo{
		CertificateAuthorityData: []byte("My cert"),
		Endpoint:                 "https://cluster-url.fake",
		Status: &types.ClusterStatus{
			Phase: types.Provisioned,
		},
		InternalState: &types.InternalState{
			TerraformState: nil,
		},
	}
	mockOp.On("Create", types.Azure, loadConfigurations(cluster, provider)).Return(result, nil)

	cluster, err := a.Provision(cluster, provider)
	require.NoError(t, err, "Provision should succeed")
	require.Equal(t, result, cluster.ClusterInfo, "The cluster info returned from the operator should be in the cluster returned by Provision")

	badCluster := &types.Cluster{
		CPU: 1,
	}
	mockOp.On("Create", types.Azure, loadConfigurations(badCluster, provider)).Return(badCluster, errors.New("Unable to provision cluster"))

	_, err = a.Provision(badCluster, provider)
	require.Error(t, err, "Provision should fail")
}

func TestDeprovision(t *testing.T) {
	mockOp := &mocks.Operator{}
	a := aksProvisioner{
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
		ProjectName:         "my-project",
		CredentialsFilePath: "/path/to/credentials",
		CustomConfigurations: map[string]interface{}{
			"target_provider": "azure",
			"target_secret":   "secret-name",
			"disk_type":       "pd-standard",
			"zone":            "europe-west3-b",
		},
	}
	goodState := &types.InternalState{
		TerraformState: terraform.NewState(),
	}
	cluster.ClusterInfo.InternalState = goodState
	mockOp.On("Delete", goodState, types.Azure, loadConfigurations(cluster, provider)).Return(nil)

	err := a.Deprovision(cluster, provider)
	require.NoError(t, err, "Deprovision should succeed")

	badState := &types.InternalState{
		TerraformState: nil,
	}
	cluster.ClusterInfo.InternalState = badState
	mockOp.On("Delete", badState, types.Azure, loadConfigurations(cluster, provider)).Return(errors.New("Unable to deprovision cluster"))

	err = a.Deprovision(cluster, provider)
	require.Error(t, err, "Deprovision should fail")
}

