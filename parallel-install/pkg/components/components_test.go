package components

import (
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"

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

	overridesProvider, err := overrides.New(k8sMock, make(map[string]interface{}), logger.NewLogger(true))
	require.NoError(t, err)

	instCfg := &config.Config{
		ComponentList: &config.ComponentList{
			Components: []config.ComponentDefinition{
				{
					Name:      "comp1",
					Namespace: "ns1",
				},
				{
					Name:      "comp2",
					Namespace: "ns2",
				},
			},
		},
		KubeconfigSource: config.KubeconfigSource{
			Path:    "path",
			Content: "",
		},
	}

	cmpMetadataTpl := helm.NewKymaComponentMetadataTemplate("version", "profile").ForComponents()
	provider := NewComponentsProvider(overridesProvider, instCfg, instCfg.ComponentList.Components, cmpMetadataTpl)

	res := provider.GetComponents(false)
	require.Equal(t, 2, len(res), "Number of components not as expected")
	require.Equal(t, "comp1", res[0].Name)

	// test reversing
	res = provider.GetComponents(true)
	require.Equal(t, 2, len(res), "Number of components not as expected")
	require.Equal(t, "comp2", res[0].Name)
}

func Test_ReuseHelmValues(t *testing.T) {
	// fake k8s with override ConfigMaps
	k8sMock := fake.NewSimpleClientset(
		&v1.ConfigMap{},
	)

	overridesProvider, err := overrides.New(k8sMock, make(map[string]interface{}), logger.NewLogger(true))
	require.NoError(t, err)

	instCfg := &config.Config{
		ComponentList: &config.ComponentList{
			Components: []config.ComponentDefinition{},
		},
		ReuseHelmValues: true,
	}

	cmpMetadataTpl := helm.NewKymaComponentMetadataTemplate("version", "profile").ForComponents()
	provider := NewComponentsProvider(overridesProvider, instCfg, instCfg.ComponentList.Components, cmpMetadataTpl)
	require.Equal(t, provider.helmConfig.ReuseValues, true)

}
