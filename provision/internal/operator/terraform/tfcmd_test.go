package terraform

import (
	"testing"

	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/stretchr/testify/require"
)

func TestInitArgs(t *testing.T) {
	// for now init args does not use the cluster and provider config for anything
	res := initArgs("", nil, "/path/to/cluster")

	require.Len(t, res, 2)
	require.Equal(t, res[0], "-var-file=/path/to/cluster/terraform.tfvars") // vars file
	require.Equal(t, res[1], "/path/to/cluster")                            // cluster config directory
}

func TestApplyArgs(t *testing.T) {
	// for now apply args does not use the cluster and provider config for anything
	res := applyArgs("", nil, "/path/to/cluster")

	require.Len(t, res, 4)
	require.Equal(t, res[0], "-state=/path/to/cluster/terraform.tfstate")   // state file
	require.Equal(t, res[1], "-var-file=/path/to/cluster/terraform.tfvars") // vars file
	require.Equal(t, res[2], "-auto-approve")                               // auto approve is important so that hydroform does not wait for user confirmation
	require.Equal(t, res[3], "/path/to/cluster")                            // cluster config directory
}

func TestImportArgs(t *testing.T) {

	cfg := map[string]interface{}{"project": "my-project", "namespace": "my-namespace", "location": "somewhere", "cluster_name": "my-cluster"}

	// test GCP
	res := importArgs(types.GCP, cfg, "/path/to/cluster")
	require.Len(t, res, 6)
	require.Equal(t, res[0], "-state=/path/to/cluster/terraform.tfstate")     // state file
	require.Equal(t, res[1], "-state-out=/path/to/cluster/terraform.tfstate") // state output file
	require.Equal(t, res[2], "-var-file=/path/to/cluster/terraform.tfvars")   // vars file
	require.Equal(t, res[3], "-config=/path/to/cluster")                      // config folder for import to know where the tf files are (if any)
	require.Equal(t, res[4], "google_container_cluster.gke_cluster")          // resource type for a GCP cluster
	require.Equal(t, res[5], "my-project/somewhere/my-cluster")               // cluster ID

	// test Gardener
	res = importArgs(types.Gardener, cfg, "/path/to/cluster")
	require.Len(t, res, 6)
	require.Equal(t, res[0], "-state=/path/to/cluster/terraform.tfstate")     // state file
	require.Equal(t, res[1], "-state-out=/path/to/cluster/terraform.tfstate") // state output file
	require.Equal(t, res[2], "-var-file=/path/to/cluster/terraform.tfvars")   // vars file
	require.Equal(t, res[3], "-config=/path/to/cluster")                      // config folder for import to know where the tf files are (if any)
	require.Equal(t, res[4], "gardener_shoot.gardener_cluster")               // resource type for a GCP cluster
	require.Equal(t, res[5], "my-namespace/my-cluster")                       // cluster ID
}
