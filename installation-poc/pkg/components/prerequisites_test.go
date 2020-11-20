package components

import (
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/config"
	"testing"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_PrerequisiteGetComponents(t *testing.T) {
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

	componentList := [][]string{[]string{"prerequisite1", "namespace1"}}

	installationCfg := config.Config{}

	provider := NewPrerequisitesProvider(overridesProvider, "", componentList, installationCfg)

	res, err := provider.GetComponents()
	require.NoError(t, err)
	require.Equal(t, 1, len(res), "Number of prerequisite components not as expected")
}
