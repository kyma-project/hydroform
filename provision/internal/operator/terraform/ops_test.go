package terraform

import (
	"testing"

	"github.com/hashicorp/terraform/command"
	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/stretchr/testify/require"
)

func TestWithUI(t *testing.T) {
	ops := &Options{}

	require.Equal(t, nil, ops.Meta.Ui, "Zero value UI should be nil")

	ui := &HydroUI{}
	WithUI(ui)(ops)

	require.Equal(t, ui, ops.Meta.Ui)
}

func TestWithDataDir(t *testing.T) {
	ops := &Options{}

	require.Equal(t, ".terraform", ops.Meta.DataDir(), "Default data dir should be .terraform")

	WithDataDir("/path/to/data")(ops)

	require.Equal(t, "/path/to/data", ops.Meta.DataDir())
}

func TestPersistent(t *testing.T) {
	ops := &Options{}

	require.False(t, ops.Persistent)

	Persistent()(ops)

	require.True(t, ops.Persistent)
}

func TestToTerraformOptions(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    types.Options
		Expected Options
	}{
		{
			Name:     "No options",
			Input:    types.Options{},
			Expected: Options{},
		},
		{
			Name: "Only data dir",
			Input: types.Options{
				DataDir: "/path/to/data",
			},
			Expected: Options{
				Meta: command.Meta{
					OverrideDataDir: "/path/to/data",
				},
			},
		},
		{
			Name: "Only persistence",
			Input: types.Options{
				Persistent: true,
			},
			Expected: Options{
				Persistent: true,
			},
		},
		{
			Name: "Datadir and Persistence",
			Input: types.Options{
				DataDir:    "/path/to/data",
				Persistent: true,
			},
			Expected: Options{
				Meta: command.Meta{
					OverrideDataDir: "/path/to/data",
				},
				Persistent: true,
			},
		},
	}

	for _, tc := range testCases {
		out := ToTerraformOptions(&tc.Input)

		ops := &Options{}
		for _, o := range out {
			o(ops)
		}

		require.Equal(t, tc.Expected, *ops, tc.Name)
	}
}
