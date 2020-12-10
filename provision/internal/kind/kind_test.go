package kind

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/states/statefile"
	"github.com/kyma-incubator/hydroform/provision/internal/operator/mocks"
	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestValidateInputs(t *testing.T) {
	t.Parallel()
	k := &kindProvisioner{}

	cluster := &types.Cluster{
		Name: "test-cluster",
	}
	provider := &types.Provider{
		Type:        types.Kind,
		ProjectName: "my-project",
		CustomConfigurations: map[string]interface{}{
			"node_image": "somerepo/image:v0.0.0",
		},
	}

	require.NoError(t, k.validateInputs(cluster, provider), "Validation should pass")

	cluster.Name = ""
	require.Error(t, k.validateInputs(cluster, provider), "Validation should fail when cluster name is empty")
	cluster.Name = "This_name_is_for_sure_way_too_long_for_the_cluster"
	require.Error(t, k.validateInputs(cluster, provider), "Validation should fail when cluster name is too long")
	cluster.Name = "-invalid-start"
	require.Error(t, k.validateInputs(cluster, provider), "Validation should fail when cluster name starts with '-'")
	cluster.Name = "invalid-end-"
	require.Error(t, k.validateInputs(cluster, provider), "Validation should fail when cluster name ends with '-'")
	cluster.Name = "hydro-cluster"

	provider.ProjectName = ""
	require.Error(t, k.validateInputs(cluster, provider), "Validation should fail when project name is empty")
	provider.ProjectName = "my-project"

	delete(provider.CustomConfigurations, "node_image")
	require.Error(t, k.validateInputs(cluster, provider), "Validation should fail when target provider is empty")
	provider.CustomConfigurations["target_provider"] = "somerepo/image:v0.0.0"
}

func TestLoadConfigurations(t *testing.T) {
	t.Parallel()
	k := &kindProvisioner{}

	cluster := &types.Cluster{
		Name: "test-cluster",
	}
	provider := &types.Provider{
		Type:        types.Kind,
		ProjectName: "my-project",
		CustomConfigurations: map[string]interface{}{
			"node_image": "somerepo/image:v0.0.0",
		},
	}

	config := k.loadConfigurations(cluster, provider)

	require.Equal(t, cluster.Name, config["cluster_name"])
	require.Equal(t, provider.ProjectName, config["project"])

	for k, v := range provider.CustomConfigurations {
		require.Equal(t, v, config[k], fmt.Sprintf("Custom config %s is incorrect", k))
	}
}

func TestProvision(t *testing.T) {
	t.Parallel()
	mockOp := &mocks.Operator{}
	k := kindProvisioner{
		provisionOperator: mockOp,
	}

	cluster := &types.Cluster{
		Name: "test-cluster",
	}
	provider := &types.Provider{
		Type:        types.Kind,
		ProjectName: "my-project",
		CustomConfigurations: map[string]interface{}{
			"node_image": "somerepo/image:v0.0.0",
		},
	}

	result := &types.ClusterInfo{
		CertificateAuthorityData: []byte{},
		Endpoint:                 "",
		Status: &types.ClusterStatus{
			Phase: types.Provisioned,
		},
		InternalState: &types.InternalState{
			TerraformState: nil,
		},
	}
	mockOp.On("Create", types.Kind, k.loadConfigurations(cluster, provider)).Return(result, nil)

	cluster, err := k.Provision(cluster, provider)
	require.NoError(t, err, "Provision should succeed")
	require.Equal(t, result, cluster.ClusterInfo, "The cluster info returned from the operator should be in the cluster returned by Provision")

	badCluster := &types.Cluster{
		Name: "",
	}
	mockOp.On("Create", types.Kind, k.loadConfigurations(badCluster, provider)).Return(badCluster, errors.New("Unable to provision cluster"))

	_, err = k.Provision(badCluster, provider)
	require.Error(t, err, "Provision should fail")
}

func TestDeprovision(t *testing.T) {
	t.Parallel()
	mockOp := &mocks.Operator{}
	k := kindProvisioner{
		provisionOperator: mockOp,
	}

	cluster := &types.Cluster{
		Name: "test-cluster",
	}
	provider := &types.Provider{
		Type:        types.Kind,
		ProjectName: "my-project",
		CustomConfigurations: map[string]interface{}{
			"node_image": "somerepo/image:v0.0.0",
		},
	}

	var state *statefile.File
	mockOp.On("Delete", state, types.Kind, k.loadConfigurations(cluster, provider)).Return(nil)

	err := k.Deprovision(cluster, provider)
	require.NoError(t, err, "Deprovision should succeed")

	provider.ProjectName = ""
	mockOp.On("Delete", state, types.Kind, k.loadConfigurations(cluster, provider)).Return(errors.New("Unable to deprovision cluster"))

	err = k.Deprovision(cluster, provider)
	require.Error(t, err, "Deprovision should fail")
}
