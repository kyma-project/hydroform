package components

import (
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
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

	overridesProvider, err := overrides.New(k8sMock, make(map[string]interface{}), true)
	require.NoError(t, err)

	compDef := ComponentDefinition{
		Name:      "prerequisite1",
		Namespace: "namespace1",
	}
	componentList := []ComponentDefinition{compDef}

	installationCfg := config.Config{}

	provider := NewPrerequisitesProvider(overridesProvider, "", componentList, installationCfg)

	res, err := provider.GetComponents()
	require.NoError(t, err)
	require.Equal(t, 1, len(res), "Number of prerequisite components not as expected")
	require.Equal(t, compDef.Name, res[0].Name)
	require.Equal(t, compDef.Namespace, res[0].Namespace)
}
