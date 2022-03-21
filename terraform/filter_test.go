package terraform

import (
	"testing"

	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/stretchr/testify/require"
)

func TestFilterVars(t *testing.T) {
	t.Parallel()
	cfg := map[string]interface{}{
		"project":        "fake-project",
		"create_timeout": "30m",
		"update_timeout": "15m",
		"delete_timeout": "10m",
		"cluster_name":   "fake-cluster",
	}

	// filter GCP
	r := filterVars(cfg, types.GCP)
	require.Equal(t, cfg, r, "GCP should not filter out any variables")

	// filter Azure
	expected := map[string]interface{}{
		"cluster_name": "fake-cluster",
	}
	r = filterVars(cfg, types.Azure)
	require.Equal(t, expected, r, "Azure should filter out variables")

	// filter AWS
	r = filterVars(cfg, types.AWS)
	require.Equal(t, cfg, r, "AWS should not filter out any variables")

	// filter Gardener
	r = filterVars(cfg, types.Gardener)
	require.Equal(t, cfg, r, "Gardener should not filter out any variables")

	// filter Kind
	r = filterVars(cfg, types.Kind)
	require.Equal(t, cfg, r, "Kind should not filter out any variables")
}
