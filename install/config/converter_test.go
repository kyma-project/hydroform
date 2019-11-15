package config

import (
	"io/ioutil"
	"testing"

	"github.com/kyma-incubator/hydroform/install/k8s"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/hydroform/install/installation"

	"github.com/stretchr/testify/require"
)

func TestYAMLToConfiguration(t *testing.T) {
	configYamlBytes, err := ioutil.ReadFile("testdata/config.yaml")
	require.NoError(t, err)
	configYamlContent := string(configYamlBytes)

	decoder, err := k8s.DefaultDecoder()
	require.NoError(t, err)

	for _, testCase := range []struct {
		description    string
		yamlContent    string
		expectedConfig installation.Configuration
	}{
		{
			description: "should create configuration from files",
			yamlContent: configYamlContent,
			expectedConfig: installation.Configuration{
				Configuration: []installation.ConfigEntry{
					{Key: "global.config.key1", Value: "value1"},
					{Key: "global.config.key2", Value: "value2"},
					{Key: "global.secret.key1", Value: "secret1", Secret: true},
				},
				ComponentConfiguration: []installation.ComponentConfiguration{
					{
						Component: "istio",
						Configuration: []installation.ConfigEntry{
							{Key: "istio.config.key1", Value: "istio-config1"},
							{Key: "istio.secret.key1", Value: "istio-secret1", Secret: true},
							{Key: "istio.secret.key2", Value: "istio-secret2", Secret: true},
						},
					},
					{
						Component: "application-connector",
						Configuration: []installation.ConfigEntry{
							{Key: "ac.config.key1", Value: "ac-config1"},
							{Key: "ac.secret.key1", Value: "ac-secret1", Secret: true},
						},
					},
				},
			},
		},
		{
			description: "should create empty config from empty yaml",
			yamlContent: ``,
			expectedConfig: installation.Configuration{
				ComponentConfiguration: []installation.ComponentConfiguration{},
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// when
			config, err := YAMLToConfiguration(testCase.yamlContent, decoder)

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedConfig, config)
		})
	}

	t.Run("should return error if invalid yaml", func(t *testing.T) {
		// when
		_, err := YAMLToConfiguration("invalid", decoder)

		// then
		require.Error(t, err)
	})

	t.Run("should return error if yaml contains object different than Secrets and Config Maps", func(t *testing.T) {
		// when
		yamlContent := `
apiVersion: v1
kind: Secret
metadata:
  name: global-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
type: Opaque
data:
  global.secret.key1: "c2VjcmV0MQ=="
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tiller
  namespace: kube-system
  labels:
    kyma-project.io/installation: ""
`

		_, err := YAMLToConfiguration(yamlContent, decoder)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected object kind")
	})
}
