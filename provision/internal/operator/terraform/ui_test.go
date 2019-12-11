package terraform

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAsk(t *testing.T) {
	ui := &HydroUI{}

	a, err := ui.Ask("Random question?")
	require.NoError(t, err, "Ask should never return an error")
	require.Equal(t, "yes", a, "Ask should always return yes")
}

func TestAskSecret(t *testing.T) {
	ui := &HydroUI{}

	a, err := ui.AskSecret("Random question?")
	require.NoError(t, err, "AskSecret should never return an error")
	require.Equal(t, "", a, "AskSecret should always return empty string")
}

func TestWarnAnError(t *testing.T) {
	ui := &HydroUI{}

	ui.Warn("WARNING")
	ui.Error("ERROR")

	require.Len(t, ui.Errors(), 2, "There should be 2 errors in total (1 errror and 1 warning)")
}
