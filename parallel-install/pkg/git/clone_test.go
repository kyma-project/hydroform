package git

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	repo       = "https://github.com/kyma-project/kyma"
	kyma117Rev = "2292dd21453af5f4f517c1c42a1cf5413d8c461c"
)

func TestCloneRevision(t *testing.T) {
	t.Parallel()

	os.RemoveAll("./clone") //ensure clone folder does not exist
	err := CloneRevision(repo, "./clone", kyma117Rev)
	defer os.RemoveAll("./clone")

	require.NoError(t, err, "Cloning Kyma 1.17 should not error")
	_, err = os.Stat("./clone")
	require.NoError(t, err, "cloned local kyma folder should not error")
	_, err = os.Stat("./clone/resources")
	require.NoError(t, err, "cloned local charts folder should not error")
}
