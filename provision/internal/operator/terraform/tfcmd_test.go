package terraform

import (
	"os"
	"testing"

	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/stretchr/testify/require"
)

func TestInitArgs(t *testing.T) {
	// test provider that has no module support
	res := initArgs("", nil, "/path/to/cluster")

	require.Len(t, res, 1)
	require.Equal(t, "/path/to/cluster", res[0]) // cluster config directory

	// test provider that has module but not an empty cluster dir => no modules will be initialized
	res = initArgs(types.Azure, nil, ".")

	require.Len(t, res, 1)
	require.Equal(t, ".", res[0]) // cluster config directory

	// test provider that has module and an empty cluster dir => modules will be initialized
	dir, err := clusterDir(".hf-test", "project", "cluster", types.Azure)
	defer os.RemoveAll(".hf-test")
	require.NoError(t, err)

	res = initArgs(types.Azure, nil, dir)
	require.Len(t, res, 2)
	require.Contains(t, res[0], "-from-module")
	require.Equal(t, res[1], dir) // cluster config directory

}

func TestApplyArgs(t *testing.T) {
	// for now apply args does not use the cluster and provider config for anything
	res := applyArgs("", nil, "/path/to/cluster")

	require.Len(t, res, 4)
	require.Equal(t, "-state=/path/to/cluster/terraform.tfstate", res[0])   // state file
	require.Equal(t, "-var-file=/path/to/cluster/terraform.tfvars", res[1]) // vars file
	require.Equal(t, "-auto-approve", res[2])                               // auto approve is important so that hydroform does not wait for user confirmation
	require.Equal(t, "/path/to/cluster", res[3])                            // cluster config directory
}

func TestImportArgs(t *testing.T) {

	cfg := map[string]interface{}{"project": "my-project", "namespace": "my-namespace", "location": "somewhere", "cluster_name": "my-cluster"}

	// test GCP
	res := importArgs(types.GCP, cfg, "/path/to/cluster")
	require.Len(t, res, 6)
	require.Equal(t, "-state=/path/to/cluster/terraform.tfstate", res[0])     // state file
	require.Equal(t, "-state-out=/path/to/cluster/terraform.tfstate", res[1]) // state output file
	require.Equal(t, "-var-file=/path/to/cluster/terraform.tfvars", res[2])   // vars file
	require.Equal(t, "-config=/path/to/cluster", res[3])                      // config folder for import to know where the tf files are (if any)
	require.Equal(t, "google_container_cluster.gke_cluster", res[4])          // resource type for a GCP cluster
	require.Equal(t, "my-project/somewhere/my-cluster", res[5])               // cluster ID

	// test Gardener
	res = importArgs(types.Gardener, cfg, "/path/to/cluster")
	require.Len(t, res, 6)
	require.Equal(t, "-state=/path/to/cluster/terraform.tfstate", res[0])     // state file
	require.Equal(t, "-state-out=/path/to/cluster/terraform.tfstate", res[1]) // state output file
	require.Equal(t, "-var-file=/path/to/cluster/terraform.tfvars", res[2])   // vars file
	require.Equal(t, "-config=/path/to/cluster", res[3])                      // config folder for import to know where the tf files are (if any)
	require.Equal(t, "gardener_shoot.gardener_cluster", res[4])               // resource type for a GCP cluster
	require.Equal(t, "my-namespace/my-cluster", res[5])                       // cluster ID
}
