package components

import (
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"io/ioutil"
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_GetComponents(t *testing.T) {
	// fake k8s with override ConfigMaps
	k8sMock := fake.NewSimpleClientset(
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

	overridesProvider, err := overrides.New(k8sMock, []string{""})
	require.NoError(t, err)

	// Read components file
	content, err := ioutil.ReadFile("../test/data/installationCR.yaml")
	require.NoError(t, err)

	installationCfg := config.Config{}

	provider := NewComponentsProvider(overridesProvider, "", string(content), installationCfg)

	res, err := provider.GetComponents()
	require.NoError(t, err)
	require.Equal(t, 21, len(res), "Number of components not as expected")
}
