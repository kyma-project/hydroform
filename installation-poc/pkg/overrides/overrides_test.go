package overrides

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_ReadOverridesFromCluster(t *testing.T) {
	component := "monitoring"

	// fake k8s with override ConfigMaps
	k8sMock := fake.NewSimpleClientset(
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "monitoring-overrides",
				Namespace: "kyma-installer",
				Labels:    map[string]string{"installer": "overrides", "component": component},
			},
			Data: map[string]string{
				"componentOverride1": "test1",
				"componentOverride2": "test2",
			},
		},
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "global-overrides",
				Namespace: "kyma-installer",
				Labels:    map[string]string{"installer": "overrides"},
			},
			Data: map[string]string{
				"globalOverride1": "test1",
				"globalOverride2": "test2",
			},
		},
	)

	// Read additional overrides file
	content, err := ioutil.ReadFile("../test/data/overrides.yaml")
	require.NoError(t, err)

	testProvider, err := New(k8sMock, string(content))
	require.NoError(t, err)

	err = testProvider.ReadOverridesFromCluster()
	require.NoError(t, err)

	// Monitoring should have two sample component overrides + one from additional overrides
	// + two global overrides from this file + one from additional overrides
	res := testProvider.OverridesFor(component)
	require.Equal(t, 6, len(res), "Number of component overrides not as expected")

	// Another component without any component override should only have two global overrides + one from additional overrides
	res2 := testProvider.OverridesFor("anotherComponent")
	require.Equal(t, 3, len(res2), "Number of global overrides not as expected")
}
