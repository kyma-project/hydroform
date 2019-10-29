package gardener

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/hydroform/internal/terraform"

	"github.com/kyma-incubator/hydroform/internal/operator/mocks"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/require"

	gardener_core "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	gardener_types "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
)

const convertError = "Status [%s] should be converted to [%s]"

func TestConvertGardenerState(t *testing.T) {
	status := gardener_types.ShootStatus{
		LastOperation: &gardener_core.LastOperation{
			State: "",
		},
	}

	require.Equal(t, types.Unknown, convertGardenertatus(status), fmt.Sprintf(convertError, "\"\"", types.Unknown))
	status.LastOperation.State = gardener_core.LastOperationStatePending
	require.Equal(t, types.Pending, convertGardenertatus(status), fmt.Sprintf(convertError, gardener_core.LastOperationStatePending, types.Pending))
	status.LastOperation.State = gardener_core.LastOperationStateProcessing
	require.Equal(t, types.Provisioning, convertGardenertatus(status), fmt.Sprintf(convertError, gardener_core.LastOperationStateProcessing, types.Provisioning))
	status.LastOperation.State = gardener_core.LastOperationStateError
	require.Equal(t, types.Errored, convertGardenertatus(status), fmt.Sprintf(convertError, gardener_core.LastOperationStateError, types.Errored))
	status.LastOperation.State = gardener_core.LastOperationStateFailed
	require.Equal(t, types.Errored, convertGardenertatus(status), fmt.Sprintf(convertError, gardener_core.LastOperationStateFailed, types.Errored))
	status.LastOperation.State = gardener_core.LastOperationStateAborted
	require.Equal(t, types.Errored, convertGardenertatus(status), fmt.Sprintf(convertError, gardener_core.LastOperationStateAborted, types.Errored))
	status.LastOperation.State = gardener_core.LastOperationStateSucceeded
	require.Equal(t, types.Provisioned, convertGardenertatus(status), fmt.Sprintf(convertError, gardener_core.LastOperationStateSucceeded, types.Provisioned))
}

func TestValidate(t *testing.T) {
	t.Run("Validate GCP config", func(t *testing.T) {
		g := gardenerProvisioner{}

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
			Type:                types.Gardener,
			ProjectName:         "my-project",
			CredentialsFilePath: "/path/to/credentials",
			CustomConfigurations: map[string]interface{}{
				"target_provider": "gcp",
				"target_secret":   "secret-name",
				"disk_type":       "pd-standard",
				"zone":            "europe-west3-b",
				"workercidr":      "10.250.0.0/19",
				"autoscaler_min":  2,
				"autoscaler_max":  4,
				"max_surge":       4,
				"max_unavailable": 1,
			},
		}

		performBasicValidation(t, g, cluster, provider)

		//gcp specific validation
		delete(provider.CustomConfigurations, "zone")
		require.Error(t, g.validate(cluster, provider), "Validation should fail when zone is empty")
		provider.CustomConfigurations["zone"] = "europe-west3-b"
	})

	t.Run("Validate Azure config", func(t *testing.T) {
		g := gardenerProvisioner{}

		cluster := &types.Cluster{
			CPU:               1,
			KubernetesVersion: "1.12",
			Name:              "hydro-cluster",
			DiskSizeGB:        35,
			NodeCount:         2,
			Location:          "eastus",
			MachineType:       "Standard_D2_v3",
		}
		provider := &types.Provider{
			Type:                types.Gardener,
			ProjectName:         "my-project",
			CredentialsFilePath: "/path/to/credentials",
			CustomConfigurations: map[string]interface{}{
				"target_provider": "azure",
				"target_secret":   "secret-name",
				"disk_type":       "standard",
				"workercidr":      "10.250.0.0/19",
				"vnetcidr":        "10.250.0.0/19",
				"autoscaler_min":  2,
				"autoscaler_max":  4,
				"max_surge":       4,
				"max_unavailable": 1,
			},
		}

		//azure specific validation
		performBasicValidation(t, g, cluster, provider)

		delete(provider.CustomConfigurations, "vnetcidr")
		require.Error(t, g.validate(cluster, provider), "Validation should fail when vnetcidr is empty")
		provider.CustomConfigurations["vnetcidr"] = "10.250.0.0/19"
	})
}

func performBasicValidation(t *testing.T, g gardenerProvisioner, cluster *types.Cluster, provider *types.Provider) {
	require.NoError(t, g.validate(cluster, provider), "Validation should pass")
	cluster.NodeCount = -5
	require.Error(t, g.validate(cluster, provider), "Validation should fail when number of nodes is < 1")
	cluster.NodeCount = 2

	cluster.Name = ""
	require.Error(t, g.validate(cluster, provider), "Validation should fail when cluster name is empty")
	cluster.Name = "This_name_is_for_sure_way_too_long_for_the_cluster"
	require.Error(t, g.validate(cluster, provider), "Validation should fail when cluster name is too long")
	cluster.Name = "-invalid-start"
	require.Error(t, g.validate(cluster, provider), "Validation should fail when cluster name starts with '-'")
	cluster.Name = "invalid-end-"
	require.Error(t, g.validate(cluster, provider), "Validation should fail when cluster name ends with '-'")
	cluster.Name = "hydro-cluster"

	cluster.Location = ""
	require.Error(t, g.validate(cluster, provider), "Validation should fail when cluster location is empty")
	cluster.Location = "europe-west3"

	cluster.MachineType = ""
	require.Error(t, g.validate(cluster, provider), "Validation should fail when cluster machine type is empty")
	cluster.MachineType = "type1"

	cluster.KubernetesVersion = ""
	require.Error(t, g.validate(cluster, provider), "Validation should fail when Kubernetes version is empty")
	cluster.KubernetesVersion = "1.12"

	cluster.DiskSizeGB = 0
	require.Error(t, g.validate(cluster, provider), "Validation should fail when disk size is 0 or less")
	cluster.DiskSizeGB = 30

	provider.CredentialsFilePath = ""
	require.Error(t, g.validate(cluster, provider), "Validation should fail when credentials file path is empty")
	provider.CredentialsFilePath = "/path/to/credentials"

	provider.ProjectName = ""
	require.Error(t, g.validate(cluster, provider), "Validation should fail when project name is empty")
	provider.ProjectName = "my-project"

	delete(provider.CustomConfigurations, "target_provider")
	require.Error(t, g.validate(cluster, provider), "Validation should fail when target provider is empty")
	provider.CustomConfigurations["target_provider"] = "nimbus"
	require.Error(t, g.validate(cluster, provider), "Validation should fail when target provider is not supported")
	provider.CustomConfigurations["target_provider"] = "gcp"

	delete(provider.CustomConfigurations, "target_secret")
	require.Error(t, g.validate(cluster, provider), "Validation should fail when target secret is empty")
	provider.CustomConfigurations["target_secret"] = "secret_name"

	delete(provider.CustomConfigurations, "disk_type")
	require.Error(t, g.validate(cluster, provider), "Validation should fail when disk type is empty")
	provider.CustomConfigurations["disk_type"] = "pd-standard"

	delete(provider.CustomConfigurations, "workercidr")
	require.Error(t, g.validate(cluster, provider), "Validation should fail when workercidr is empty")
	provider.CustomConfigurations["workercidr"] = "10.250.0.0/19"

	delete(provider.CustomConfigurations, "autoscaler_min")
	require.Error(t, g.validate(cluster, provider), "Validation should fail when autoscaler_min is empty")
	provider.CustomConfigurations["autoscaler_min"] = 2

	delete(provider.CustomConfigurations, "autoscaler_max")
	require.Error(t, g.validate(cluster, provider), "Validation should fail when autoscaler_max is empty")
	provider.CustomConfigurations["autoscaler_max"] = 4

	delete(provider.CustomConfigurations, "max_surge")
	require.Error(t, g.validate(cluster, provider), "Validation should fail when max_surge is empty")
	provider.CustomConfigurations["max_surge"] = 4

	delete(provider.CustomConfigurations, "max_unavailable")
	require.Error(t, g.validate(cluster, provider), "Validation should fail when max_unavailable is empty")
	provider.CustomConfigurations["max_unavailable"] = 1
}

func TestLoadConfigurations(t *testing.T) {

	g := gardenerProvisioner{}

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
		Type:                types.Gardener,
		ProjectName:         "my-project",
		CredentialsFilePath: "/path/to/credentials",
		CustomConfigurations: map[string]interface{}{
			"target_provider": "gcp",
			"target_secret":   "secret-name",
			"disk_type":       "pd-standard",
			"zone":            "europe-west3-b",
		},
	}

	config := g.loadConfigurations(cluster, provider)

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
	g := gardenerProvisioner{
		operator: mockOp,
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
		Type:                types.Gardener,
		ProjectName:         "my-project",
		CredentialsFilePath: "/path/to/credentials",
		CustomConfigurations: map[string]interface{}{
			"target_provider": "gcp",
			"target_secret":   "secret-name",
			"disk_type":       "pd-standard",
			"zone":            "europe-west3-b",
			"workercidr":      "10.250.0.0/19",
			"autoscaler_min":  2,
			"autoscaler_max":  4,
			"max_surge":       4,
			"max_unavailable": 1,
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
	mockOp.On("Create", types.Gardener, g.loadConfigurations(cluster, provider)).Return(result, nil)

	cluster, err := g.Provision(cluster, provider)
	require.NoError(t, err, "Provision should succeed")
	require.Equal(t, result, cluster.ClusterInfo, "The cluster info returned from the operator should be in the cluster returned by Provision")

	badCluster := &types.Cluster{
		CPU: 1,
	}
	mockOp.On("Create", types.Gardener, g.loadConfigurations(badCluster, provider)).Return(badCluster, errors.New("Unable to provision cluster"))

	_, err = g.Provision(badCluster, provider)
	require.Error(t, err, "Provision should fail")
}

func TestDeProvision(t *testing.T) {
	mockOp := &mocks.Operator{}
	g := gardenerProvisioner{
		operator: mockOp,
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
		Type:                types.Gardener,
		ProjectName:         "my-project",
		CredentialsFilePath: "/path/to/credentials",
		CustomConfigurations: map[string]interface{}{
			"target_provider": "gcp",
			"target_secret":   "secret-name",
			"disk_type":       "pd-standard",
			"zone":            "europe-west3-b",
			"workercidr":      "10.250.0.0/19",
			"autoscaler_min":  2,
			"autoscaler_max":  4,
			"max_surge":       4,
			"max_unavailable": 1,
		},
	}
	goodState := &types.InternalState{
		TerraformState: terraform.NewState(),
	}
	cluster.ClusterInfo.InternalState = goodState
	mockOp.On("Delete", goodState, types.Gardener, g.loadConfigurations(cluster, provider)).Return(nil)

	err := g.Deprovision(cluster, provider)
	require.NoError(t, err, "Deprovision should succeed")

	badState := &types.InternalState{
		TerraformState: nil,
	}
	cluster.ClusterInfo.InternalState = badState
	mockOp.On("Delete", badState, types.Gardener, g.loadConfigurations(cluster, provider)).Return(errors.New("Unable to deprovision cluster"))

	err = g.Deprovision(cluster, provider)
	require.Error(t, err, "Deprovision should fail")
}
