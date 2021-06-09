package helm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ReuseHelmValues(t *testing.T) {
	cfg := Config{
		ReuseValues: true,
	}
	client := NewClient(cfg)
	actionCfg, err := client.newActionConfig("comp", "path")
	require.NoError(t, err)
	upg := client.newUpgrade(actionCfg)
	require.Equal(t, true, upg.ReuseValues)
}
